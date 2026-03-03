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

// ============ Table-Driven Tests ============

func TestCreateBill(t *testing.T) {
	tests := []struct {
		name        string
		currency    model.Currency
		wantErr     bool
		checkBill   func(*testing.T, *model.Bill)
	}{
		{
			name:     "creates bill with USD",
			currency: model.CurrencyUSD,
			wantErr:  false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				if bill.ID == "" {
					t.Error("expected bill ID to be generated")
				}
				if bill.Status != model.BillStatusOpen {
					t.Errorf("expected status open, got %v", bill.Status)
				}
				if bill.Currency != model.CurrencyUSD {
					t.Errorf("expected USD, got %v", bill.Currency)
				}
			},
		},
		{
			name:     "creates bill with GEL",
			currency: model.CurrencyGEL,
			wantErr:  false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				if bill.Currency != model.CurrencyGEL {
					t.Errorf("expected GEL, got %v", bill.Currency)
				}
			},
		},
		{
			name:     "defaults to USD when empty",
			currency: "",
			wantErr:  false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				if bill.Currency != model.CurrencyUSD {
					t.Errorf("expected default USD, got %v", bill.Currency)
				}
			},
		},
		{
			name:     "rejects unsupported currency",
			currency: "EUR",
			wantErr:  true,
			checkBill: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockBillRepository()
			svc := NewBillingService(repo)

			req := &model.CreateBillRequest{Currency: tt.currency}
			bill, err := svc.CreateBill(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkBill != nil && bill != nil {
				tt.checkBill(t, bill)
			}
		})
	}
}

func TestAddLineItem(t *testing.T) {
	tests := []struct {
		name        string
		setupBill   func(*BillingService) string
		req         *model.AddLineItemRequest
		wantErr     bool
		checkBill   func(*testing.T, *model.Bill)
	}{
		{
			name: "adds line item successfully",
			setupBill: func(svc *BillingService) string {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				return bill.ID
			},
			req: &model.AddLineItemRequest{
				Description: "Service fee",
				Amount:      10.00,
				Currency:    model.CurrencyUSD,
			},
			wantErr: false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				if len(bill.LineItems) != 1 {
					t.Errorf("expected 1 line item, got %d", len(bill.LineItems))
				}
				if bill.TotalAmount != 10.00 {
					t.Errorf("expected total 10.00, got %f", bill.TotalAmount)
				}
			},
		},
		{
			name: "converts GEL to USD",
			setupBill: func(svc *BillingService) string {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				return bill.ID
			},
			req: &model.AddLineItemRequest{
				Description: "Service fee",
				Amount:      100.00,
				Currency:    model.CurrencyGEL,
			},
			wantErr: false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				// 100 GEL * 0.37 = 37 USD
				if bill.TotalAmount != 37.0 {
					t.Errorf("expected 37.0, got %f", bill.TotalAmount)
				}
			},
		},
		{
			name: "fails for closed bill",
			setupBill: func(svc *BillingService) string {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				svc.CloseBill(bill.ID)
				return bill.ID
			},
			req: &model.AddLineItemRequest{
				Description: "Service fee",
				Amount:      10.00,
				Currency:    model.CurrencyUSD,
			},
			wantErr: true,
			checkBill: nil,
		},
		{
			name: "fails for nonexistent bill",
			setupBill: func(svc *BillingService) string {
				return "nonexistent"
			},
			req: &model.AddLineItemRequest{
				Description: "Service fee",
				Amount:      10.00,
				Currency:    model.CurrencyUSD,
			},
			wantErr: true,
			checkBill: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockBillRepository()
			svc := NewBillingService(repo)

			billID := tt.setupBill(svc)
			bill, err := svc.AddLineItem(billID, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddLineItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkBill != nil && bill != nil {
				tt.checkBill(t, bill)
			}
		})
	}
}

func TestCloseBill(t *testing.T) {
	tests := []struct {
		name      string
		setupBill func(*BillingService) string
		wantErr   bool
		checkBill func(*testing.T, *model.Bill)
	}{
		{
			name: "closes open bill",
			setupBill: func(svc *BillingService) string {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				return bill.ID
			},
			wantErr: false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				if bill.Status != model.BillStatusClosed {
					t.Errorf("expected closed, got %v", bill.Status)
				}
				if bill.ClosedAt == nil {
					t.Error("expected ClosedAt to be set")
				}
			},
		},
		{
			name: "fails for already closed bill",
			setupBill: func(svc *BillingService) string {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				svc.CloseBill(bill.ID)
				return bill.ID
			},
			wantErr:   true,
			checkBill: nil,
		},
		{
			name: "fails for nonexistent bill",
			setupBill: func(svc *BillingService) string {
				return "nonexistent"
			},
			wantErr:   true,
			checkBill: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockBillRepository()
			svc := NewBillingService(repo)

			billID := tt.setupBill(svc)
			bill, err := svc.CloseBill(billID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CloseBill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkBill != nil && bill != nil {
				tt.checkBill(t, bill)
			}
		})
	}
}

func TestGetBill(t *testing.T) {
	tests := []struct {
		name      string
		setupBill func(*BillingService) string
		billID    string
		wantErr   bool
		checkBill func(*testing.T, *model.Bill)
	}{
		{
			name: "retrieves existing bill",
			setupBill: func(svc *BillingService) string {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				return bill.ID
			},
			billID:  "",
			wantErr: false,
			checkBill: func(t *testing.T, bill *model.Bill) {
				if bill.ID == "" {
					t.Error("expected bill ID")
				}
			},
		},
		{
			name:      "fails for nonexistent bill",
			setupBill: func(svc *BillingService) string { return "" },
			billID:    "nonexistent",
			wantErr:   true,
			checkBill: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockBillRepository()
			svc := NewBillingService(repo)

			expectedID := tt.setupBill(svc)
			if tt.billID != "" {
				expectedID = tt.billID
			}

			bill, err := svc.GetBill(expectedID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkBill != nil && bill != nil {
				tt.checkBill(t, bill)
			}
		})
	}
}

func TestListBills(t *testing.T) {
	tests := []struct {
		name      string
		setupBills func(*BillingService)
		status    string
		wantCount int
	}{
		{
			name: "lists all bills",
			setupBills: func(svc *BillingService) {
				svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyGEL})
			},
			status:    "",
			wantCount: 2,
		},
		{
			name: "filters by open status",
			setupBills: func(svc *BillingService) {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				svc.CloseBill(bill.ID)
				svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
			},
			status:    "open",
			wantCount: 1,
		},
		{
			name: "filters by closed status",
			setupBills: func(svc *BillingService) {
				bill, _ := svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
				svc.CloseBill(bill.ID)
				svc.CreateBill(&model.CreateBillRequest{Currency: model.CurrencyUSD})
			},
			status:    "closed",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockBillRepository()
			svc := NewBillingService(repo)

			tt.setupBills(svc)
			bills, err := svc.ListBills(tt.status)

			if err != nil {
				t.Errorf("ListBills() error = %v", err)
				return
			}

			if len(bills) != tt.wantCount {
				t.Errorf("got %d bills, want %d", len(bills), tt.wantCount)
			}
		})
	}
}
