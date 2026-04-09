// Package errors demonstrates error handling across hexagonal boundaries.
// Domain errors, adapter errors, and the translation between them.
package errors

import (
	"context"
	"errors"
	"fmt"
)

// --- Domain errors ---

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderEmpty    = errors.New("order must have items")
	ErrNotPending    = errors.New("operation requires pending status")
)

// DomainError wraps a domain error with additional context.
type DomainError struct {
	Op  string // operation that failed
	Err error  // underlying domain error
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.Err)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// --- Domain types ---

type Order struct {
	ID         string
	CustomerID string
	Status     string
}

// --- Port ---

type OrderRepository interface {
	FindByID(ctx context.Context, id string) (Order, error)
}

// --- Service showing error wrapping ---

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		// The service returns domain errors. If the adapter already
		// returned a domain error (like ErrOrderNotFound), pass it through.
		// If it returned an infrastructure error, wrap it.
		if errors.Is(err, ErrOrderNotFound) {
			return Order{}, err
		}
		return Order{}, &DomainError{Op: "GetOrder", Err: fmt.Errorf("repository: %w", err)}
	}
	return order, nil
}

// --- Adapter that translates infrastructure errors ---

type InMemoryRepo struct {
	orders map[string]Order
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{orders: make(map[string]Order)}
}

func (r *InMemoryRepo) FindByID(_ context.Context, id string) (Order, error) {
	order, ok := r.orders[id]
	if !ok {
		// Translate "map miss" into domain error
		return Order{}, ErrOrderNotFound
	}
	return order, nil
}

func (r *InMemoryRepo) Add(order Order) {
	r.orders[order.ID] = order
}
