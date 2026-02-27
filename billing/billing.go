package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"encore.dev/storage/sqldb"
	"encore.dev/storage/pubsub"
)

// Currency represents the currency type
type Currency string

const (
	CurrencyGEL Currency = "GEL"
	CurrencyUSD Currency = "USD"
)

// BillStatus represents the status of a bill
type BillStatus string

const (
	BillStatusOpen   BillStatus = "open"
	BillStatusClosed BillStatus = "closed"
)

// Bill represents a billing invoice
type Bill struct {
	ID          string      `json:"id"`
	Status      BillStatus  `json:"status"`
	Currency    Currency    `json:"currency"`
	TotalAmount float64     `json:"totalAmount"`
	LineItems   []LineItem  `json:"lineItems,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	ClosedAt    *time.Time  `json:"closedAt,omitempty"`
}

// LineItem represents a single line item on a bill
type LineItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Currency    Currency  `json:"currency"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateBillRequest represents the request to create a new bill
type CreateBillRequest struct {
	Currency Currency `json:"currency"`
}

// CreateBillResponse represents the response from creating a bill
type CreateBillResponse struct {
	Bill Bill `json:"bill"`
}

// AddLineItemRequest represents the request to add a line item
type AddLineItemRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    Currency `json:"currency"`
}

// AddLineItemResponse represents the response from adding a line item
type AddLineItemResponse struct {
	Bill Bill `json:"bill"`
}

// CloseBillResponse represents the response from closing a bill
type CloseBillResponse struct {
	Bill Bill `json:"bill"`
}

// GetBillResponse represents the response from getting a bill
type GetBillResponse struct {
	Bill Bill `json:"bill"`
}

// ListBillsRequest represents the request to list bills
type ListBillsRequest struct {
	Status string `query:"status"`
}

// ListBillsResponse represents the response from listing bills
type ListBillsResponse struct {
	Bills []Bill `json:"bills"`
}

// Exchange rates (in production, this would come from an external service)
var exchangeRates = map[Currency]float64{
	CurrencyGEL: 0.37, // 1 GEL = 0.37 USD
	CurrencyUSD: 1.0,
}

// ConvertToUSD converts an amount from one currency to USD
func ConvertToUSD(amount float64, currency Currency) float64 {
	if currency == CurrencyUSD {
		return amount
	}
	return amount * exchangeRates[CurrencyGEL]
}

// DB is the database connection
var DB = sqldb.Database("billing")

// billsTopic is used for publishing bill events
var billsTopic = pubsub.NewTopic("bills", pubsub.TopicConfig{
	DeliveryPolicy: pubsub.DeliverAll,
})

//encore:api public
func CreateBill(ctx context.Context, req *CreateBillRequest) (*CreateBillResponse, error) {
	if req.Currency == "" {
		req.Currency = CurrencyUSD
	}
	if req.Currency != CurrencyGEL && req.Currency != CurrencyUSD {
		return nil, fmt.Errorf("unsupported currency: %s", req.Currency)
	}

	bill := Bill{
		ID:          generateID(),
		Status:      BillStatusOpen,
		Currency:    req.Currency,
		LineItems:   []LineItem{},
		CreatedAt:   time.Now().UTC(),
	}

	// Store in database (using in-memory map for simplicity in this example)
	bills[bill.ID] = bill

	// Publish event for Temporal workflow
	billsTopic.Publish(ctx, bill.ID, BillEvent{
		BillID: bill.ID,
		Type:   "created",
		Time:   time.Now().UTC(),
	})

	return &CreateBillResponse{Bill: bill}, nil
}

//encore:api public
func AddLineItem(ctx context.Context, billID string, req *AddLineItemRequest) (*AddLineItemResponse, error) {
	bill, ok := bills[billID]
	if !ok {
		return nil, fmt.Errorf("bill not found: %s", billID)
	}

	if bill.Status == BillStatusClosed {
		return nil, fmt.Errorf("cannot add line item to closed bill: %s", billID)
	}

	if req.Currency != CurrencyGEL && req.Currency != CurrencyUSD {
		return nil, fmt.Errorf("unsupported currency: %s", req.Currency)
	}

	lineItem := LineItem{
		ID:          generateID(),
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CreatedAt:   time.Now().UTC(),
	}

	bill.LineItems = append(bill.LineItems, lineItem)

	// Update total amount (normalized to bill's currency)
	if bill.Currency == req.Currency {
		bill.TotalAmount += req.Amount
	} else if req.Currency == CurrencyGEL && bill.Currency == CurrencyUSD {
		bill.TotalAmount += ConvertToUSD(req.Amount, CurrencyGEL)
	} else if req.Currency == CurrencyUSD && bill.Currency == CurrencyGEL {
		bill.TotalAmount += ConvertToUSD(req.Amount, CurrencyUSD)
	}

	bills[billID] = bill

	// Publish event for Temporal workflow
	billsTopic.Publish(ctx, billID, BillEvent{
		BillID:     billID,
		Type:       "line_item_added",
		LineItemID: lineItem.ID,
		Time:       time.Now().UTC(),
	})

	return &AddLineItemResponse{Bill: bill}, nil
}

//encore:api public
func CloseBill(ctx context.Context, billID string) (*CloseBillResponse, error) {
	bill, ok := bills[billID]
	if !ok {
		return nil, fmt.Errorf("bill not found: %s", billID)
	}

	if bill.Status == BillStatusClosed {
		return nil, fmt.Errorf("bill already closed: %s", billID)
	}

	now := time.Now().UTC()
	bill.Status = BillStatusClosed
	bill.ClosedAt = &now
	bills[billID] = bill

	// Publish event for Temporal workflow
	billsTopic.Publish(ctx, billID, BillEvent{
		BillID: billID,
		Type:   "closed",
		Time:   now,
	})

	return &CloseBillResponse{Bill: bill}, nil
}

//encore:api public
func GetBill(ctx context.Context, billID string) (*GetBillResponse, error) {
	bill, ok := bills[billID]
	if !ok {
		return nil, fmt.Errorf("bill not found: %s", billID)
	}

	return &GetBillResponse{Bill: bill}, nil
}

//encore:api public
func ListBills(ctx context.Context, req *ListBillsRequest) (*ListBillsResponse, error) {
	var result []Bill

	for _, bill := range bills {
		if req.Status != "" && string(bill.Status) != req.Status {
			continue
		}
		result = append(result, bill)
	}

	return &ListBillsResponse{Bills: result}, nil
}

// In-memory storage (in production, use database)
var bills = make(map[string]Bill)

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("bill_%d", time.Now().UnixNano())
}

// BillEvent represents an event for the bill
type BillEvent struct {
	BillID     string
	Type       string
	LineItemID string
	Time       time.Time
}

// BillCreatedHandler handles bill creation events
// This would be connected to a Temporal workflow in production
var _ = billsTopic.Subscribe(func(ctx context.Context, event BillEvent) error {
	// In a real implementation, this would signal a Temporal workflow
	// to track the billing period and handle progressive accrual
	fmt.Printf("Bill event: %s - %s\n", event.BillID, event.Type)
	return nil
})

// Raw handler for pubsub
func init() {
	// Subscribe to bill events for Temporal workflow integration
}

// Sentinel error for closed bill
var ErrBillClosed = fmt.Errorf("bill is closed")

// ValidateBillOpen checks if a bill is open and returns an error if closed
func ValidateBillOpen(billID string) error {
	bill, ok := bills[billID]
	if !ok {
		return fmt.Errorf("bill not found: %s", billID)
	}
	if bill.Status == BillStatusClosed {
		return ErrBillClosed
	}
	return nil
}

// GetBillTotal returns the total amount in the bill's currency
func GetBillTotal(billID string) (float64, error) {
	bill, ok := bills[billID]
	if !ok {
		return 0, fmt.Errorf("bill not found: %s", billID)
	}
	return bill.TotalAmount, nil
}

// MarshalJSON implements custom JSON marshaling
func (b Bill) MarshalJSON() ([]byte, error) {
	type Alias Bill
	return json.Marshal(&struct {
		Alias
		TotalAmountDisplay string `json:"totalAmountDisplay"`
	}{
		Alias:               Alias(b),
		TotalAmountDisplay: fmt.Sprintf("%.2f %s", b.TotalAmount, b.Currency),
	})
}
