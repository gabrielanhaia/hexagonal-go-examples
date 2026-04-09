// Package memory provides an in-memory implementation of OrderRepository.
// This is an outbound adapter — it fulfills a port defined by the domain.
// Useful for tests and prototyping.
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/gabrielanhaia/hexagonal-go-examples/ch06/adapters"
)

type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]adapters.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]adapters.Order),
	}
}

func (r *OrderRepository) Save(_ context.Context, order adapters.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

func (r *OrderRepository) FindByID(_ context.Context, id string) (adapters.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[id]
	if !ok {
		return adapters.Order{}, fmt.Errorf("order %s not found", id)
	}
	return order, nil
}
