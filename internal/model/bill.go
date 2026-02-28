package model

import "time"

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
	LineItems   []LineItem `json:"lineItems,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	ClosedAt    *time.Time `json:"closedAt,omitempty"`
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
