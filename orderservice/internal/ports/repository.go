package ports

import (
	"context"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

// OrderSaver persists a single order.
type OrderSaver interface {
	Save(ctx context.Context, order domain.Order) error
}

// OrderFinder retrieves a single order by ID.
type OrderFinder interface {
	FindByID(ctx context.Context, id string) (domain.Order, error)
}

// OrderLister returns orders for a customer.
type OrderLister interface {
	ListByCustomer(ctx context.Context, customerID string) ([]domain.Order, error)
}

// OrderRepository composes read and write operations.
type OrderRepository interface {
	OrderSaver
	OrderFinder
}
