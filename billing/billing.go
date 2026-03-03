package billing

import (
	"context"

	"fees-api/internal/model"
	"fees-api/workflow"

	"go.temporal.io/sdk/client"
)

//encore:api public method=POST path=/bills
func CreateBill(ctx context.Context, req *model.CreateBillRequest) (*model.CreateBillResponse, error) {
	svc := GetService()
	bill, err := svc.svc.CreateBill(req)
	if err != nil {
		return nil, err
	}
	return &model.CreateBillResponse{Bill: *bill}, nil
}

//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *model.AddLineItemRequest) (*model.AddLineItemResponse, error) {
	svc := GetService()
	bill, err := svc.svc.AddLineItem(billID, req)
	if err != nil {
		return nil, err
	}
	return &model.AddLineItemResponse{Bill: *bill}, nil
}

//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) (*model.CloseBillResponse, error) {
	svc := GetService()
	bill, err := svc.svc.CloseBill(billID)
	if err != nil {
		return nil, err
	}
	return &model.CloseBillResponse{Bill: *bill}, nil
}

//encore:api public method=GET path=/bills/:billID
func GetBill(ctx context.Context, billID string) (*model.GetBillResponse, error) {
	svc := GetService()
	bill, err := svc.svc.GetBill(billID)
	if err != nil {
		return nil, err
	}
	return &model.GetBillResponse{Bill: *bill}, nil
}

//encore:api public method=GET path=/bills
func ListBills(ctx context.Context, req *model.ListBillsRequest) (*model.ListBillsResponse, error) {
	svc := GetService()
	bills, err := svc.svc.ListBills(req.Status)
	if err != nil {
		return nil, err
	}
	return &model.ListBillsResponse{Bills: bills}, nil
}

// ============ Temporal Workflow Endpoints ============

// StartWorkflowRequest is the request to start a billing workflow
type StartWorkflowRequest struct {
	Currency          string `json:"currency"`
	BillingPeriodDays int    `json:"billingPeriodDays"`
}

// StartWorkflowResponse is the response from starting a workflow
type StartWorkflowResponse struct {
	WorkflowID string `json:"workflowId"`
}

//encore:api public method=POST path=/bills/:billID/workflow
func StartBillingWorkflow(ctx context.Context, billID string, req *StartWorkflowRequest) (*StartWorkflowResponse, error) {
	svc := GetService()

	input := workflow.BillingPeriodInput{
		BillID:            billID,
		Currency:          req.Currency,
		BillingPeriodDays: req.BillingPeriodDays,
	}

	workflowID := "billing-period-" + billID

	we, err := svc.client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueueName,
	}, workflow.BillingPeriodWorkflow, input)
	if err != nil {
		return nil, err
	}

	return &StartWorkflowResponse{WorkflowID: we.GetID()}, nil
}

// SignalWorkflowRequest is the request to signal a workflow
type SignalWorkflowRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

//encore:api public method=POST path=/bills/:billID/workflow/add-item
func SignalAddLineItem(ctx context.Context, billID string, req *SignalWorkflowRequest) error {
	svc := GetService()
	workflowID := "billing-period-" + billID

	return svc.client.SignalWorkflow(ctx, workflowID, "", "add-line-item", workflow.AddLineItemSignalInput{
		Amount:   req.Amount,
		Currency: req.Currency,
	})
}

//encore:api public method=POST path=/bills/:billID/workflow/close
func SignalCloseBill(ctx context.Context, billID string) error {
	svc := GetService()
	workflowID := "billing-period-" + billID

	return svc.client.SignalWorkflow(ctx, workflowID, "", "close-bill", nil)
}
