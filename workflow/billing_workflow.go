package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/workflow"
)

// BillingPeriodInput is the input for starting the billing period workflow
type BillingPeriodInput struct {
	BillID            string `json:"billId"`
	Currency          string `json:"currency"`
	BillingPeriodDays int    `json:"billingPeriodDays"`
}

// BillState represents the current state of a bill in the workflow
type BillState struct {
	BillID        string     `json:"billId"`
	Status        string     `json:"status"`
	Currency      string     `json:"currency"`
	TotalAmount   float64    `json:"totalAmount"`
	LineItemCount int        `json:"lineItemCount"`
	StartedAt     time.Time  `json:"startedAt"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
}

// AddLineItemSignalInput is the input for adding a line item signal
type AddLineItemSignalInput struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// BillingPeriodWorkflow manages the lifecycle of a billing period
func BillingPeriodWorkflow(ctx workflow.Context, input BillingPeriodInput) error {
	// Initialize state
	state := BillState{
		BillID:        input.BillID,
		Status:        "open",
		Currency:      input.Currency,
		TotalAmount:   0,
		LineItemCount: 0,
		StartedAt:     workflow.Now(ctx),
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Set up timer for billing period end
	timerDuration := time.Duration(input.BillingPeriodDays) * 24 * time.Hour
	timerFuture := workflow.NewTimer(ctx, timerDuration)

	// Set up signal channels
	addLineItemChan := workflow.GetSignalChannel(ctx, "add-line-item")
	closeBillChan := workflow.GetSignalChannel(ctx, "close-bill")

	// Selector for handling events
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		// Timer fired - period ended, auto-close the bill
		state.Status = "closed"
		now := workflow.Now(ctx)
		state.ClosedAt = &now

		// Call activity to close the bill via HTTP API
		err := workflow.ExecuteActivity(ctx, CloseBillActivity, CloseBillActivityInput{
			BillID: input.BillID,
		}).Get(ctx, nil)
		if err != nil {
			// Log error but don't fail workflow - bill state is already closed
			state.Status = "close-failed"
		}
	})
	selector.AddReceive(addLineItemChan, func(c workflow.ReceiveChannel, more bool) {
		var signalInput AddLineItemSignalInput
		c.Receive(ctx, &signalInput)

		// Execute activity to add line item
		err := workflow.ExecuteActivity(ctx, AddLineItemActivity, ActivityInput{
			BillID:   input.BillID,
			Amount:   signalInput.Amount,
			Currency: signalInput.Currency,
		}).Get(ctx, nil)
		if err == nil {
			state.LineItemCount++
			state.TotalAmount += signalInput.Amount
		}
	})
	selector.AddReceive(closeBillChan, func(c workflow.ReceiveChannel, more bool) {
		state.Status = "closed"
		now := workflow.Now(ctx)
		state.ClosedAt = &now

		// Call activity to close the bill via HTTP API
		_ = workflow.ExecuteActivity(ctx, CloseBillActivity, CloseBillActivityInput{
			BillID: input.BillID,
		}).Get(ctx, nil)
	})

	// Register query handler
	workflow.SetQueryHandler(ctx, "bill-state", func() (BillState, error) {
		return state, nil
	})

	// Wait until closed
	for state.Status == "open" {
		selector.Select(ctx)
	}

	return nil
}

// ============ Activities ============

// ActivityInput represents input for activities
type ActivityInput struct {
	BillID      string  `json:"billId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

// CloseBillActivityInput represents input for closing a bill
type CloseBillActivityInput struct {
	BillID string `json:"billId"`
}

// API base URL - in production this would be configurable
const apiBaseURL = "http://127.0.0.1:4000"

// CreateBillingActivity creates a billing period (placeholder)
func CreateBillingActivity(ctx context.Context, input ActivityInput) (string, error) {
	return input.BillID, nil
}

// AddLineItemActivity adds a line item via HTTP API
func AddLineItemActivity(ctx context.Context, input ActivityInput) error {
	url := fmt.Sprintf("%s/bills/%s/items", apiBaseURL, input.BillID)
	
	payload := map[string]interface{}{
		"description": input.Description,
		"amount":      input.Amount,
		"currency":    input.Currency,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	return nil
}

// CloseBillActivity closes a bill via HTTP API
func CloseBillActivity(ctx context.Context, input CloseBillActivityInput) error {
	url := fmt.Sprintf("%s/bills/%s/close", apiBaseURL, input.BillID)
	
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	return nil
}
