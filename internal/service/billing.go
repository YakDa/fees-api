package service

import (
	"fmt"
	"math"
	"time"

	"fees-api/internal/model"
	"fees-api/internal/repository"
	billingerrors "fees-api/pkg/errors"
)

// Exchange rates to USD (base currency)
var exchangeRatesToUSD = map[model.Currency]float64{
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
	// Input validation
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if len(req.Description) > 500 {
		return nil, fmt.Errorf("description too long (max 500 characters)")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

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

	// Convert float64 to int64 cents to avoid floating point errors
	amountCents := floatToCents(req.Amount)

	lineItem := model.LineItem{
		ID:          generateID(),
		Description: req.Description,
		Amount:      amountCents,
		Currency:    req.Currency,
		CreatedAt:   time.Now().UTC(),
	}

	bill.LineItems = append(bill.LineItems, lineItem)

	// Update total amount (normalized to bill's currency)
	bill.TotalAmount = s.convertAndAdd(bill.TotalAmount, bill.Currency, amountCents, req.Currency)

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

// ConvertToUSD converts amount (in cents) from one currency to USD cents
func (s *BillingService) ConvertToUSD(amountCents int64, currency model.Currency) int64 {
	if currency == model.CurrencyUSD {
		return amountCents
	}
	// Convert GEL cents to USD cents using exchange rate
	usdFloat := float64(amountCents) * exchangeRatesToUSD[model.CurrencyGEL]
	return int64(math.Round(usdFloat))
}

// convertAndAdd converts amount (in cents) to bill's currency and adds to total (also in cents)
func (s *BillingService) convertAndAdd(totalCents int64, billCurrency model.Currency, amountCents int64, amountCurrency model.Currency) int64 {
	if billCurrency == amountCurrency {
		return totalCents + amountCents
	}
	// Convert amount to USD cents first, then to bill's currency
	usdCents := s.ConvertToUSD(amountCents, amountCurrency)
	
	// Convert from USD to bill's currency
	if billCurrency == model.CurrencyUSD {
		return totalCents + usdCents
	}
	
	// USD to GEL: divide by rate (1 GEL = 0.37 USD, so 1 USD = 1/0.37 GEL)
	gelFloat := float64(usdCents) / exchangeRatesToUSD[model.CurrencyGEL]
	return totalCents + int64(math.Round(gelFloat))
}

// floatToCents converts a float64 dollar amount to int64 cents
func floatToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("bill_%d", time.Now().UnixNano())
}
