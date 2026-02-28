package billing

import (
	"context"

	"fees-api/internal/model"
	"fees-api/internal/presentation/handlers"
	"fees-api/internal/repository"
	"fees-api/internal/service"
)

// Initialize the service (singleton pattern for Encore)
var (
	repo    = repository.NewInMemoryBillRepository()
	svc     = service.NewBillingService(repo)
	handler = handlers.NewBillingHandler(svc)
)

//encore:api public
func CreateBill(ctx context.Context, req *model.CreateBillRequest) (*model.CreateBillResponse, error) {
	return handler.CreateBill(ctx, req)
}

//encore:api public
func AddLineItem(ctx context.Context, billID string, req *model.AddLineItemRequest) (*model.AddLineItemResponse, error) {
	return handler.AddLineItem(ctx, billID, req)
}

//encore:api public
func CloseBill(ctx context.Context, billID string) (*model.CloseBillResponse, error) {
	return handler.CloseBill(ctx, billID)
}

//encore:api public
func GetBill(ctx context.Context, billID string) (*model.GetBillResponse, error) {
	return handler.GetBill(ctx, billID)
}

//encore:api public
func ListBills(ctx context.Context, req *model.ListBillsRequest) (*model.ListBillsResponse, error) {
	return handler.ListBills(ctx, req)
}
