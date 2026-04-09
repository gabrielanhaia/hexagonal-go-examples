package ports

import (
	"context"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

// NotificationSender sends notifications about order events.
type NotificationSender interface {
	OrderPlaced(ctx context.Context, order domain.Order) error
}
