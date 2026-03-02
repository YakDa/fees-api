package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
)

// WorkflowManager handles billing period workflows
type WorkflowManager struct {
	client client.Client
}

// NewWorkflowManager creates a new workflow manager
func NewWorkflowManager(c client.Client) *WorkflowManager {
	return &WorkflowManager{client: c}
}

// StartBillingPeriod starts a new billing period workflow
//
// TODO: Implement with your workflow function
func (m *WorkflowManager) StartBillingPeriod(ctx context.Context, billID, currency string, billingPeriodDays int) (string, error) {
	workflowID := fmt.Sprintf("billing-period-%s", billID)
	
	// TODO: Uncomment when workflow is ready
	// input := workflow.BillingPeriodInput{...}
	// _, err := m.client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
	//     ID:        workflowID,
	//     TaskQueue: "billing",
	// }, yourWorkflowFunction, input)

	return workflowID, nil
}

// AddLineItem signals a line item to the billing workflow
func (m *WorkflowManager) AddLineItem(ctx context.Context, billID string, amount float64, currency string) error {
	workflowID := fmt.Sprintf("billing-period-%s", billID)

	err := m.client.SignalWorkflow(ctx, workflowID, "", "add-line-item", map[string]interface{}{
		"amount":   amount,
		"currency": currency,
	})
	if err != nil {
		return fmt.Errorf("failed to signal workflow: %w", err)
	}

	return nil
}

// CloseBill signals the workflow to close the bill
func (m *WorkflowManager) CloseBill(ctx context.Context, billID string) error {
	workflowID := fmt.Sprintf("billing-period-%s", billID)

	err := m.client.SignalWorkflow(ctx, workflowID, "", "close-bill", nil)
	if err != nil {
		return fmt.Errorf("failed to signal workflow: %w", err)
	}

	return nil
}

// GetBillState queries the current state of a billing period
func (m *WorkflowManager) GetBillState(ctx context.Context, billID string) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("billing-period-%s", billID)

	var result map[string]interface{}
	resp, err := m.client.QueryWorkflow(ctx, workflowID, "", "bill-state", &result)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}
	
	_ = resp // unused
	return result, nil
}

// CancelWorkflow cancels a billing workflow
func (m *WorkflowManager) CancelWorkflow(ctx context.Context, billID string) error {
	workflowID := fmt.Sprintf("billing-period-%s", billID)

	err := m.client.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	return nil
}
