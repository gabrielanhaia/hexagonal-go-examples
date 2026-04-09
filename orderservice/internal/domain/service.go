package domain

import (
	"context"
	"fmt"
)

// OrderService orchestrates domain operations. It depends on ports (interfaces),
// never on concrete adapters.
type OrderService struct {
	repo     OrderRepository
	notifier NotificationSender
	idgen    IDGenerator
}

// Ports used by the service — defined here in the domain.
type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	FindByID(ctx context.Context, id string) (Order, error)
}

type NotificationSender interface {
	OrderPlaced(ctx context.Context, order Order) error
}

type IDGenerator interface {
	NewID() string
}

func NewOrderService(repo OrderRepository, notifier NotificationSender, idgen IDGenerator) *OrderService {
	return &OrderService{
		repo:     repo,
		notifier: notifier,
		idgen:    idgen,
	}
}

type PlaceOrderRequest struct {
	CustomerID string
	Items      []LineItem
}

func (s *OrderService) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error) {
	order, err := NewOrder(s.idgen.NewID(), req.CustomerID, req.Items)
	if err != nil {
		return Order{}, fmt.Errorf("creating order: %w", err)
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return Order{}, fmt.Errorf("saving order: %w", err)
	}

	// Notification failure shouldn't fail the order placement.
	_ = s.notifier.OrderPlaced(ctx, order)

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s *OrderService) ConfirmOrder(ctx context.Context, id string) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := order.Confirm(); err != nil {
		return err
	}
	return s.repo.Save(ctx, order)
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := order.Cancel(); err != nil {
		return err
	}
	return s.repo.Save(ctx, order)
}
