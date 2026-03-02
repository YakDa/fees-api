package service

import (
	"context"
	"fmt"
	"time"

	"fees-api/internal/model"
	"fees-api/internal/repository"
	billingerrors "fees-api/pkg/errors"
)

// Exchange rates (in production, this would come from an external service)
var exchangeRates = map[model.Currency]float64{
	model.CurrencyGEL: 0.37, // 1 GEL = 0.37 USD
	model.CurrencyUSD: 1.0,
}

// BillingService handles business logic for billing
type BillingService struct {
	repo              repository.BillRepository
	useTemporal       bool
	billingPeriodDays int
}

// NewBillingService creates a new billing service
func NewBillingService(repo repository.BillRepository) *BillingService {
	return &BillingService{
		repo:              repo,
		useTemporal:       false, // Disabled by default
		billingPeriodDays: 30,     // Default 30 days
	}
}

// NewBillingServiceWithTemporal creates a billing service with Temporal integration
func NewBillingServiceWithTemporal(repo repository.BillRepository, billingPeriodDays int) *BillingService {
	return &BillingService{
		repo:              repo,
		useTemporal:       true,
		billingPeriodDays: billingPeriodDays,
	}
}

// CreateBill creates a new bill
func (s *BillingService) CreateBill(req *model.CreateBillRequest) (*model.Bill, error) {
	if req.Currency == "" {
		req.Currency = model.CurrencyUSD
	}
	if req.Currency != model.CurrencyGEL && req.Currency != model.CurrencyUSD {
		return nil, billingerrors.UnsupportedCurrency(string(req.Currency))
	}

	bill := &model.Bill{
		ID:        generateID(),
		Status:    model.BillStatusOpen,
		Currency:  req.Currency,
		LineItems: []model.LineItem{},
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(bill); err != nil {
		return nil, err
	}

	return bill, nil
}

// AddLineItem adds a line item to a bill
func (s *BillingService) AddLineItem(billID string, req *model.AddLineItemRequest) (*model.Bill, error) {
	bill, err := s.repo.Get(billID)
	if err != nil {
		return nil, err
	}
	if bill == nil {
		return nil, billingerrors.BillNotFound(billID)
	}

	if bill.Status == model.BillStatusClosed {
		return nil, billingerrors.BillClosed(billID)
	}

	if req.Currency != model.CurrencyGEL && req.Currency != model.CurrencyUSD {
		return nil, billingerrors.UnsupportedCurrency(string(req.Currency))
	}

	lineItem := model.LineItem{
		ID:          generateID(),
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CreatedAt:   time.Now().UTC(),
	}

	bill.LineItems = append(bill.LineItems, lineItem)

	// Update total amount (normalized to bill's currency)
	bill.TotalAmount = s.convertAndAdd(bill.TotalAmount, bill.Currency, req.Amount, req.Currency)

	if err := s.repo.Update(bill); err != nil {
		return nil, err
	}

	return bill, nil
}

// CloseBill closes a bill
func (s *BillingService) CloseBill(billID string) (*model.Bill, error) {
	bill, err := s.repo.Get(billID)
	if err != nil {
		return nil, err
	}
	if bill == nil {
		return nil, billingerrors.BillNotFound(billID)
	}

	if bill.Status == model.BillStatusClosed {
		return nil, billingerrors.BillClosed(billID)
	}

	now := time.Now().UTC()
	bill.Status = model.BillStatusClosed
	bill.ClosedAt = &now

	if err := s.repo.Update(bill); err != nil {
		return nil, err
	}

	return bill, nil
}

// GetBill retrieves a bill by ID
func (s *BillingService) GetBill(billID string) (*model.Bill, error) {
	bill, err := s.repo.Get(billID)
	if err != nil {
		return nil, err
	}
	if bill == nil {
		return nil, billingerrors.BillNotFound(billID)
	}
	return bill, nil
}

// ListBills lists all bills, optionally filtered by status
func (s *BillingService) ListBills(status string) ([]model.Bill, error) {
	return s.repo.List(status)
}

// ============ HTTP Handlers (for Encore API) ============

// BillingHandler handles HTTP requests for billing
type BillingHandler struct {
	svc *BillingService
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(svc *BillingService) *BillingHandler {
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

// ConvertToUSD converts an amount from one currency to USD
func (s *BillingService) ConvertToUSD(amount float64, currency model.Currency) float64 {
	if currency == model.CurrencyUSD {
		return amount
	}
	return amount * exchangeRates[model.CurrencyGEL]
}

// convertAndAdd converts amount to bill's currency and adds to total
func (s *BillingService) convertAndAdd(total float64, billCurrency model.Currency, amount float64, amountCurrency model.Currency) float64 {
	if billCurrency == amountCurrency {
		return total + amount
	}
	// Convert amount to USD first, then to bill's currency
	usdAmount := s.ConvertToUSD(amount, amountCurrency)
	// Assume 1:1 for USD to bill currency (simplified)
	return total + usdAmount
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("bill_%d", time.Now().UnixNano())
}
