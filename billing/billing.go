package billing

import (
	"context"

	"fees-api/internal/model"
)

//encore:api public method=POST path=/bills
func CreateBill(ctx context.Context, req *model.CreateBillRequest) (*model.CreateBillResponse, error) {
	svc := GetService()
	
	bill, err := svc.svc.CreateBill(req)
	if err != nil {
		return nil, err
	}

	// Automatically start the billing workflow
	_ = svc.startWorkflow(ctx, bill.ID, string(bill.Currency), 30)

	return &model.CreateBillResponse{Bill: *bill}, nil
}

//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *model.AddLineItemRequest) (*model.AddLineItemResponse, error) {
	svc := GetService()
	
	bill, err := svc.svc.AddLineItem(billID, req)
	if err != nil {
		return nil, err
	}

	// Automatically signal the workflow
	_ = svc.signalAddItem(ctx, billID, req.Amount, string(req.Currency))

	return &model.AddLineItemResponse{Bill: *bill}, nil
}

//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) (*model.CloseBillResponse, error) {
	svc := GetService()
	
	bill, err := svc.svc.CloseBill(billID)
	if err != nil {
		return nil, err
	}

	// Automatically signal the workflow to close
	_ = svc.signalCloseBill(ctx, billID)

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
