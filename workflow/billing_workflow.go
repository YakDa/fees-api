package workflow

import (
	"context"
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

	// Set up timer for billing period end
	timerFuture := workflow.NewTimer(ctx, time.Duration(input.BillingPeriodDays)*24*time.Hour)

	// Set up signal channels
	addLineItemChan := workflow.GetSignalChannel(ctx, "add-line-item")
	closeBillChan := workflow.GetSignalChannel(ctx, "close-bill")

	// Selector for handling events
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		// Timer fired - close the bill
		state.Status = "closed"
		now := workflow.Now(ctx)
		state.ClosedAt = &now
	})
	selector.AddReceive(addLineItemChan, func(c workflow.ReceiveChannel, more bool) {
		var signalInput AddLineItemSignalInput
		c.Receive(ctx, &signalInput)

		// Execute activity to add line item
		state.LineItemCount++
		state.TotalAmount += signalInput.Amount
	})
	selector.AddReceive(closeBillChan, func(c workflow.ReceiveChannel, more bool) {
		state.Status = "closed"
		now := workflow.Now(ctx)
		state.ClosedAt = &now
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

// CreateBillingActivity creates a billing period
func CreateBillingActivity(ctx context.Context, input ActivityInput) (string, error) {
	// This would call the Encore API to create a bill
	// For now, return the workflow ID
	return input.BillID, nil
}

// AddLineItemActivity adds a line item
func AddLineItemActivity(ctx context.Context, input ActivityInput) error {
	// This would call the Encore API to add a line item
	return nil
}

// CloseBillActivity closes a bill
func CloseBillActivity(ctx context.Context, input ActivityInput) error {
	// This would call the Encore API to close the bill
	return nil
}
