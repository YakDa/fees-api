package errors

import "fmt"

// Bill errors
var (
	ErrBillNotFound = fmt.Errorf("bill not found")
	ErrBillClosed   = fmt.Errorf("bill is closed")
)

// BillNotFoundError returns an error for bill not found
func BillNotFound(billID string) error {
	return fmt.Errorf("bill not found: %s", billID)
}

// BillClosedError returns an error for closed bill
func BillClosed(billID string) error {
	return fmt.Errorf("cannot modify closed bill: %s", billID)
}

// UnsupportedCurrencyError returns an error for unsupported currency
func UnsupportedCurrency(currency string) error {
	return fmt.Errorf("unsupported currency: %s", currency)
}
