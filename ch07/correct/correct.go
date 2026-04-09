// Package correct shows the RIGHT way: domain defines ports, adapters implement them.
// Dependencies point inward. The domain package has zero external imports.
package correct

import (
	"context"
	"fmt"
)

// --- Domain layer (center of the hexagon) ---
// Notice: no imports from any adapter, database, or HTTP package.

type Order struct {
	ID         string
	CustomerID string
	TotalCents int64
}

// OrderRepository is a PORT — defined by the domain.
// The domain says "I need something that can do this."
type OrderRepository interface {
	Save(ctx context.Context, order Order) error
}

// OrderService contains business logic. It depends on the PORT (interface),
// not on any concrete adapter.
type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, customerID string, totalCents int64) (Order, error) {
	if customerID == "" {
		return Order{}, fmt.Errorf("customer ID is required")
	}
	if totalCents <= 0 {
		return Order{}, fmt.Errorf("total must be positive")
	}

	order := Order{
		ID:         fmt.Sprintf("ord_%s_%d", customerID, totalCents),
		CustomerID: customerID,
		TotalCents: totalCents,
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return Order{}, fmt.Errorf("saving order: %w", err)
	}

	return order, nil
}

// --- Adapter layer (edge of the hexagon) ---
// This would normally be in a separate package.
// It DEPENDS ON the domain (imports the Order type and implements the port).
// The domain does NOT depend on it.

type InMemoryRepository struct {
	Orders map[string]Order
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{Orders: make(map[string]Order)}
}

// Save implements OrderRepository. No "implements" keyword needed.
func (r *InMemoryRepository) Save(_ context.Context, order Order) error {
	r.Orders[order.ID] = order
	return nil
}
