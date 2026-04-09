// Package console provides a NotificationSender that writes to stdout.
// This is an outbound adapter — useful for development and debugging.
package console

import (
	"context"
	"fmt"

	"github.com/gabrielanhaia/hexagonal-go-examples/ch06/adapters"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) OrderPlaced(_ context.Context, order adapters.Order) error {
	fmt.Printf("[NOTIFICATION] Order %s placed by customer %s (total: %s %.2f)\n",
		order.ID, order.CustomerID, order.Total.Currency, float64(order.Total.Cents)/100)
	return nil
}
