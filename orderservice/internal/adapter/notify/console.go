package notify

import (
	"context"
	"fmt"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) OrderPlaced(_ context.Context, order domain.Order) error {
	fmt.Printf("[NOTIFICATION] Order %s placed by customer %s\n",
		order.ID, order.CustomerID)
	return nil
}
