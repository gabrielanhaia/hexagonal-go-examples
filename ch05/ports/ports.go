// Package ports demonstrates how to design Go interfaces as hexagonal ports.
// Each interface is small, focused, and defined by the domain's needs — not
// by the adapter's capabilities.
package ports

import (
	"context"
	"time"
)

// --- Domain types (repeated here for self-contained example) ---

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
)

type Money struct {
	Cents    int64
	Currency string
}

type Order struct {
	ID         string
	CustomerID string
	Items      []OrderItem
	Status     OrderStatus
	Total      Money
	CreatedAt  time.Time
}

type OrderItem struct {
	ProductID string
	Quantity  int
	Price     Money
}

// --- Ports: small, focused interfaces ---

// OrderSaver is a write-only port. If your service only needs to save orders,
// it should depend on this — not on a fat interface that also includes
// Find, List, Delete, and Count.
type OrderSaver interface {
	Save(ctx context.Context, order Order) error
}

// OrderFinder is a read-only port.
type OrderFinder interface {
	FindByID(ctx context.Context, id string) (Order, error)
}

// OrderRepository combines read and write. Use this when a service genuinely
// needs both, but prefer the smaller interfaces when it doesn't.
type OrderRepository interface {
	OrderSaver
	OrderFinder
}

// OrderLister is a separate port for listing/searching.
// Not every consumer needs this capability.
type OrderLister interface {
	ListByCustomer(ctx context.Context, customerID string) ([]Order, error)
}

// NotificationSender is a port for sending notifications.
// One method. The domain doesn't care if it's email, SMS, Slack, or a carrier pigeon.
type NotificationSender interface {
	OrderPlaced(ctx context.Context, order Order) error
}

// PriceLookup is a port for retrieving product prices.
// The domain needs prices but doesn't care where they come from.
type PriceLookup interface {
	GetPrice(ctx context.Context, productID string) (Money, error)
}

// CouponValidator is a port for validating discount coupons.
type CouponValidator interface {
	Validate(ctx context.Context, code string) (discountPercent int, err error)
}

// --- A service that depends on ports, not implementations ---

type OrderService struct {
	repo     OrderRepository
	notifier NotificationSender
	pricer   PriceLookup
}

func NewOrderService(repo OrderRepository, notifier NotificationSender, pricer PriceLookup) *OrderService {
	return &OrderService{
		repo:     repo,
		notifier: notifier,
		pricer:   pricer,
	}
}

func (s *OrderService) PlaceOrder(ctx context.Context, customerID string, productIDs []string) (*Order, error) {
	var items []OrderItem
	var totalCents int64

	for _, pid := range productIDs {
		price, err := s.pricer.GetPrice(ctx, pid)
		if err != nil {
			return nil, err
		}
		items = append(items, OrderItem{
			ProductID: pid,
			Quantity:  1,
			Price:     price,
		})
		totalCents += price.Cents
	}

	order := &Order{
		ID:         "ord_" + customerID, // simplified
		CustomerID: customerID,
		Items:      items,
		Status:     OrderStatusPending,
		Total:      Money{Cents: totalCents, Currency: "USD"},
	}

	if err := s.repo.Save(ctx, *order); err != nil {
		return nil, err
	}

	// Fire-and-forget notification (in production, you'd handle this differently)
	_ = s.notifier.OrderPlaced(ctx, *order)

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
