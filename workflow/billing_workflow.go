package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// BillingPeriodWorkflow manages the lifecycle of a billing period
// This workflow is started at the beginning of a fee period
// and allows for progressive accrual of fees
type BillingPeriodWorkflow struct{}

const (
	// Signal to add a line item
	AddLineItemSignal = "add-line-item"
	// Signal to close the bill
	CloseBillSignal = "close-bill"
	// Query to get current bill state
	BillStateQuery = "bill-state"
)

// BillState represents the current state of a bill in the workflow
type BillState struct {
	BillID        string    `json:"billId"`
	Status        string    `json:"status"`
	Currency      string    `json:"currency"`
	TotalAmount   float64   `json:"totalAmount"`
	LineItemCount int       `json:"lineItemCount"`
	StartedAt     time.Time `json:"startedAt"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
}

// BillingPeriodInput is the input for starting the billing period workflow
type BillingPeriodInput struct {
	BillID            string        `json:"billId"`
	Currency          string        `json:"currency"`
	BillingPeriodDays int           `json:"billingPeriodDays"` // e.g., 30 for monthly
}

// RunBillingPeriodWorkflow is the main workflow function
func RunBillingPeriodWorkflow(ctx workflow.Context, input BillingPeriodInput) error {
	// Initialize the bill state
	state := BillState{
		BillID:        input.BillID,
		Status:        "open",
		Currency:      input.Currency,
		TotalAmount:   0,
		LineItemCount: 0,
		StartedAt:     workflow.Now(ctx),
	}

	// Set up a timer for the billing period end
	billingPeriodTimer := workflow.NewTimer(ctx, time.Duration(input.BillingPeriodDays)*24*time.Hour)

	// Channel for receiving signals
	addLineItemChan := workflow.GetSignalChannel(ctx, AddLineItemSignal)
	closeBillChan := workflow.GetSignalChannel(ctx, CloseBillSignal)

	selector := workflow.NewSelector(ctx)
	selector.AddFuture(billingPeriodTimer, func(f workflow.Future) {
		// Billing period ended - close the bill automatically
		state.Status = "closed"
		now := workflow.Now(ctx)
		state.ClosedAt = &now
	})
	selector.AddReceive(addLineItemChan, func(c workflow.ReceiveChannel, more bool) {
		var lineItem struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		}
		c.Receive(ctx, &lineItem)

		// Add the line item to the bill
		// In a real implementation, this would call the Encore API
		state.LineItemCount++
		if lineItem.Currency == state.Currency {
			state.TotalAmount += lineItem.Amount
		}
		// Note: Currency conversion would be handled by the API
	})
	selector.AddReceive(closeBillChan, func(c workflow.ReceiveChannel, more bool) {
		var _ struct{} // Empty signal
		state.Status = "closed"
		now := workflow.Now(ctx)
		state.ClosedAt = &now
	})

	// Register query handler
	err := workflow.SetQueryHandler(ctx, BillStateQuery, func() (BillState, error) {
		return state, nil
	})
	if err != nil {
		return err
	}

	// Wait until either the billing period ends or the bill is closed
	for state.Status == "open" {
		selector.Select(ctx)
	}

	// Workflow complete - the bill is now closed
	// In production, this would trigger final invoice generation
	fmt.Printf("Billing period complete for bill %s: total amount %.2f %s\n",
		state.BillID, state.TotalAmount, state.Currency)

	return nil
}

// TemporalWorkflowOptions returns the workflow options for billing period
func TemporalWorkflowOptions(billID string) workflow.Options {
	return workflow.Options{
		TaskQueue:           "billing",
		ID:                  fmt.Sprintf("billing-period-%s", billID),
		ExecutionStartToCloseTimeout: 365 * 24 * time.Hour, // 1 year max
		WorkflowIDReusePolicy: workflow.WorkflowIDReusePolicyAllowDuplicate,
	}
}

// AddLineItemSignal is the signal to add a line item to the bill
type AddLineItemSignal struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// CloseBillSignal is the signal to close the bill
type CloseBillSignal struct{}

// BillingPeriodActivity handles the actual bill operations
// In production, these would call the Encore APIs
type BillingPeriodActivity struct{}

// CreateBillingPeriod creates a new billing period workflow
func CreateBillingPeriod(ctx context.Context, input BillingPeriodInput) (string, error) {
	// This would be implemented using Temporal client
	// In production, use temporalClient.StartWorkflow(ctx, options, RunBillingPeriodWorkflow, input)
	return fmt.Sprintf("billing-period-%s", input.BillID), nil
}

// CloseBillingPeriod closes an existing billing period
func CloseBillingPeriod(ctx context.Context, workflowID string) error {
	// This would signal the workflow to close
	return nil
}

// GetBillingPeriodState queries the current state of a billing period
func GetBillingPeriodState(ctx context.Context, workflowID string) (*BillState, error) {
	// This would query the workflow state
	return nil, fmt.Errorf("not implemented")
}
