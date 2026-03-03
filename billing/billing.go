package billing

import (
	"context"

	"fees-api/internal/model"
	"fees-api/internal/repository"
	"fees-api/internal/service"
)

// Initialize the service (singleton pattern for Encore)
var (
	repo    = repository.NewInMemoryBillRepository()
	svc     = service.NewBillingService(repo)
	handler = service.NewBillingHandler(svc)
)

//encore:api public method=POST path=/bills
func CreateBill(ctx context.Context, req *model.CreateBillRequest) (*model.CreateBillResponse, error) {
	return handler.CreateBill(ctx, req)
}

//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *model.AddLineItemRequest) (*model.AddLineItemResponse, error) {
	return handler.AddLineItem(ctx, billID, req)
}

//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) (*model.CloseBillResponse, error) {
	return handler.CloseBill(ctx, billID)
}

//encore:api public method=GET path=/bills/:billID
func GetBill(ctx context.Context, billID string) (*model.GetBillResponse, error) {
	return handler.GetBill(ctx, billID)
}

//encore:api public method=GET path=/bills
func ListBills(ctx context.Context, req *model.ListBillsRequest) (*model.ListBillsResponse, error) {
	return handler.ListBills(ctx, req)
}

// TODO: Add Temporal integration endpoints
//
// To enable Temporal workflows:
//
// 1. Start Temporal local server:
//    ./tools/temporalite_0.3.0_darwin_arm64/temporalite start -n default
//
// 2. Initialize Temporal client in billing service
//
// 3. Add endpoints to start/complete workflows:
//
// //encore:api public method=POST path=/bills/:billID/workflow
// func StartBillingWorkflow(ctx context.Context, billID string, req *model.StartWorkflowRequest) (*model.StartWorkflowResponse, error) {
//     workflowID, err := temporalManager.StartBillingPeriod(ctx, billID, req.Currency, req.BillingPeriodDays)
//     if err != nil {
//         return nil, err
//     }
//     return &model.StartWorkflowResponse{WorkflowID: workflowID}, nil
// }
