package handlers

import (
	"context"

	"fees-api/internal/model"
	"fees-api/internal/service"
)

// BillingHandler handles HTTP requests for billing
type BillingHandler struct {
	svc *service.BillingService
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// CreateBill handles the CreateBill API
func (h *BillingHandler) CreateBill(ctx context.Context, req *model.CreateBillRequest) (*model.CreateBillResponse, error) {
	bill, err := h.svc.CreateBill(req)
	if err != nil {
		return nil, err
	}
	return &model.CreateBillResponse{Bill: *bill}, nil
}

// AddLineItem handles the AddLineItem API
func (h *BillingHandler) AddLineItem(ctx context.Context, billID string, req *model.AddLineItemRequest) (*model.AddLineItemResponse, error) {
	bill, err := h.svc.AddLineItem(billID, req)
	if err != nil {
		return nil, err
	}
	return &model.AddLineItemResponse{Bill: *bill}, nil
}

// CloseBill handles the CloseBill API
func (h *BillingHandler) CloseBill(ctx context.Context, billID string) (*model.CloseBillResponse, error) {
	bill, err := h.svc.CloseBill(billID)
	if err != nil {
		return nil, err
	}
	return &model.CloseBillResponse{Bill: *bill}, nil
}

// GetBill handles the GetBill API
func (h *BillingHandler) GetBill(ctx context.Context, billID string) (*model.GetBillResponse, error) {
	bill, err := h.svc.GetBill(billID)
	if err != nil {
		return nil, err
	}
	return &model.GetBillResponse{Bill: *bill}, nil
}

// ListBills handles the ListBills API
func (h *BillingHandler) ListBills(ctx context.Context, req *model.ListBillsRequest) (*model.ListBillsResponse, error) {
	bills, err := h.svc.ListBills(req.Status)
	if err != nil {
		return nil, err
	}
	return &model.ListBillsResponse{Bills: bills}, nil
}
