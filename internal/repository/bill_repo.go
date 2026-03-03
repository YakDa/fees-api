package repository

import (
	"fees-api/internal/model"
	"sync"
)

// BillRepository defines the interface for bill data access
type BillRepository interface {
	Create(bill *model.Bill) error
	Get(id string) (*model.Bill, error)
	Update(bill *model.Bill) error
	List(status string) ([]model.Bill, error)
}

// InMemoryBillRepository is an in-memory implementation of BillRepository
type InMemoryBillRepository struct {
	mu    sync.RWMutex
	bills map[string]model.Bill
}

// NewInMemoryBillRepository creates a new in-memory bill repository
func NewInMemoryBillRepository() *InMemoryBillRepository {
	return &InMemoryBillRepository{
		bills: make(map[string]model.Bill),
	}
}

// Create creates a new bill
func (r *InMemoryBillRepository) Create(bill *model.Bill) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bills[bill.ID] = *bill
	return nil
}

// Get retrieves a bill by ID
func (r *InMemoryBillRepository) Get(id string) (*model.Bill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bill, ok := r.bills[id]
	if !ok {
		return nil, nil
	}
	return &bill, nil
}

// Update updates an existing bill
func (r *InMemoryBillRepository) Update(bill *model.Bill) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bills[bill.ID]; !ok {
		return nil
	}
	r.bills[bill.ID] = *bill
	return nil
}

// List returns all bills, optionally filtered by status
func (r *InMemoryBillRepository) List(status string) ([]model.Bill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []model.Bill
	for _, bill := range r.bills {
		if status != "" && string(bill.Status) != status {
			continue
		}
		result = append(result, bill)
	}
	return result, nil
}
