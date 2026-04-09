package memory

import (
	"context"
	"sync"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{orders: make(map[string]domain.Order)}
}

func (r *OrderRepository) Save(_ context.Context, order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

func (r *OrderRepository) FindByID(_ context.Context, id string) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

func (r *OrderRepository) ListByCustomer(_ context.Context, customerID string) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Order
	for _, o := range r.orders {
		if o.CustomerID == customerID {
			result = append(result, o)
		}
	}
	return result, nil
}
