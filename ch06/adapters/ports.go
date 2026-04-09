package adapters

import "context"

// Ports — defined by the domain, implemented by adapters.

type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	FindByID(ctx context.Context, id string) (Order, error)
}

type NotificationSender interface {
	OrderPlaced(ctx context.Context, order Order) error
}
