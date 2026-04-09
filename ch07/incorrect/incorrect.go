// Package incorrect shows the WRONG way: domain depends on infrastructure.
// This is the anti-pattern. The domain directly imports database/sql,
// making it impossible to test without a real database and impossible to
// swap the storage layer.
package incorrect

import (
	"context"
	"database/sql"
	"fmt"
)

type Order struct {
	ID         string
	CustomerID string
	TotalCents int64
}

// OrderService DIRECTLY depends on *sql.DB — a concrete infrastructure type.
// The dependency arrow points OUTWARD (domain -> infrastructure). This is wrong.
type OrderService struct {
	db *sql.DB // <-- THIS IS THE PROBLEM
}

func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(ctx context.Context, customerID string, totalCents int64) (Order, error) {
	if customerID == "" {
		return Order{}, fmt.Errorf("customer ID is required")
	}

	order := Order{
		ID:         fmt.Sprintf("ord_%s_%d", customerID, totalCents),
		CustomerID: customerID,
		TotalCents: totalCents,
	}

	// Business logic is now welded to SQL.
	// Want to test the validation above? You need a database connection.
	// Want to use a different storage? Rewrite the service.
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO orders (id, customer_id, total_cents) VALUES ($1, $2, $3)",
		order.ID, order.CustomerID, order.TotalCents,
	)
	if err != nil {
		return Order{}, fmt.Errorf("inserting order: %w", err)
	}

	return order, nil
}
