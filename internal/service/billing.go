package service

import (
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
	repo repository.BillRepository
}

// NewBillingService creates a new billing service
func NewBillingService(repo repository.BillRepository) *BillingService {
	return &BillingService{repo: repo}
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
		ID:          generateID(),
		Status:      model.BillStatusOpen,
		Currency:    req.Currency,
		LineItems:   []model.LineItem{},
		CreatedAt:   time.Now().UTC(),
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
