package service

import (
	"testing"

	"fees-api/internal/model"
)

// mockBillRepository is a mock implementation of BillRepository for testing
type mockBillRepository struct {
	bills map[string]model.Bill
}

func newMockBillRepository() *mockBillRepository {
	return &mockBillRepository{
		bills: make(map[string]model.Bill),
	}
}

func (m *mockBillRepository) Create(bill *model.Bill) error {
	m.bills[bill.ID] = *bill
	return nil
}

func (m *mockBillRepository) Get(id string) (*model.Bill, error) {
	bill, ok := m.bills[id]
	if !ok {
		return nil, nil
	}
	return &bill, nil
}

func (m *mockBillRepository) Update(bill *model.Bill) error {
	m.bills[bill.ID] = *bill
	return nil
}

func (m *mockBillRepository) List(status string) ([]model.Bill, error) {
	var result []model.Bill
	for _, bill := range m.bills {
		if status != "" && string(bill.Status) != status {
			continue
		}
		result = append(result, bill)
	}
	return result, nil
}

func TestCreateBill(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	req := &model.CreateBillRequest{Currency: model.CurrencyUSD}
	bill, err := svc.CreateBill(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bill.ID == "" {
		t.Error("expected bill ID to be generated")
	}

	if bill.Status != model.BillStatusOpen {
		t.Errorf("expected status to be open, got %v", bill.Status)
	}

	if bill.Currency != model.CurrencyUSD {
		t.Errorf("expected currency to be USD, got %v", bill.Currency)
	}
}

func TestCreateBillDefaultCurrency(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	req := &model.CreateBillRequest{}
	bill, err := svc.CreateBill(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bill.Currency != model.CurrencyUSD {
		t.Errorf("expected default currency to be USD, got %v", bill.Currency)
	}
}

func TestCreateBillUnsupportedCurrency(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	req := &model.CreateBillRequest{Currency: "EUR"}
	_, err := svc.CreateBill(req)

	if err == nil {
		t.Fatal("expected error for unsupported currency")
	}
}

func TestAddLineItem(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	// Create a bill first
	bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})

	// Add a line item
	req := &model.AddLineItemRequest{
		Description: "Service fee",
		Amount:      10.00,
		Currency:    model.CurrencyUSD,
	}
	updatedBill, err := svc.AddLineItem(bill.ID, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(updatedBill.LineItems) != 1 {
		t.Errorf("expected 1 line item, got %d", len(updatedBill.LineItems))
	}

	if updatedBill.TotalAmount != 10.00 {
		t.Errorf("expected total 10.00, got %f", updatedBill.TotalAmount)
	}
}

func TestAddLineItemToClosedBill(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	// Create and close a bill
	bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
	svc.CloseBill(bill.ID)

	// Try to add line item
	req := &model.AddLineItemRequest{
		Description: "Service fee",
		Amount:      10.00,
		Currency:    model.CurrencyUSD,
	}
	_, err := svc.AddLineItem(bill.ID, req)

	if err == nil {
		t.Fatal("expected error when adding to closed bill")
	}
}

func TestAddLineItemCurrencyConversion(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	// Create a bill in USD
	bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})

	// Add a line item in GEL
	req := &model.AddLineItemRequest{
		Description: "Service fee",
		Amount:      100.00,
		Currency:    model.CurrencyGEL,
	}
	updatedBill, err := svc.AddLineItem(bill.ID, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 100 GEL * 0.37 = 37 USD
	expected := 37.0
	if updatedBill.TotalAmount != expected {
		t.Errorf("expected total %f, got %f", expected, updatedBill.TotalAmount)
	}
}

func TestCloseBill(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
	closedBill, err := svc.CloseBill(bill.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if closedBill.Status != model.BillStatusClosed {
		t.Errorf("expected status to be closed, got %v", closedBill.Status)
	}

	if closedBill.ClosedAt == nil {
		t.Error("expected ClosedAt to be set")
	}
}

func TestCloseBillAlreadyClosed(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
	svc.CloseBill(bill.ID)

	_, err := svc.CloseBill(bill.ID)

	if err == nil {
		t.Fatal("expected error when closing already closed bill")
	}
}

func TestGetBill(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	created, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
	retrieved, err := svc.GetBill(created.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestGetBillNotFound(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	_, err := svc.GetBill("nonexistent")

	if err == nil {
		t.Fatal("expected error for nonexistent bill")
	}
}

func TestListBills(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	// Create some bills
	svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
	svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyGEL})

	bills, err := svc.ListBills("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bills) != 2 {
		t.Errorf("expected 2 bills, got %d", len(bills))
	}
}

func TestListBillsByStatus(t *testing.T) {
	repo := newMockBillRepository()
	svc := NewBillingService(repo)

	// Create and close a bill
	bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
	svc.CloseBill(bill.ID)

	// Create another open bill
	svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})

	openBills, _ := svc.ListBills("open")
	closedBills, _ := svc.ListBills("closed")

	if len(openBills) != 1 {
		t.Errorf("expected 1 open bill, got %d", len(openBills))
	}

	if len(closedBills) != 1 {
		t.Errorf("expected 1 closed bill, got %d", len(closedBills))
	}
}
