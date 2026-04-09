// Package events demonstrates event-driven adapters in a hexagonal architecture.
// The domain publishes events through a port. Adapters handle delivery.
package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// --- Domain event ---

type EventType string

const (
	EventOrderPlaced    EventType = "order.placed"
	EventOrderConfirmed EventType = "order.confirmed"
	EventOrderCanceled  EventType = "order.canceled"
)

type OrderEvent struct {
	Type       EventType
	OrderID    string
	CustomerID string
	OccurredAt time.Time
}

// --- Domain types ---

type LineItem struct {
	ProductID string
	Quantity  int
}

type Order struct {
	ID         string
	CustomerID string
	Items      []LineItem
	Status     string
}

// Confirm transitions the order to confirmed status.
func (o *Order) Confirm() error {
	if o.Status != "pending" {
		return fmt.Errorf("cannot confirm order in %q status", o.Status)
	}
	o.Status = "confirmed"
	return nil
}

// --- Port: ID generation ---

type IDGenerator interface {
	NewID() string
}

// --- Port: the domain publishes events through this ---

type EventPublisher interface {
	Publish(ctx context.Context, event OrderEvent) error
}

// --- Port: the domain's core operations ---

type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	FindByID(ctx context.Context, id string) (Order, error)
}

// --- Domain service that publishes events ---

type OrderService struct {
	repo      OrderRepository
	publisher EventPublisher
	idGen     IDGenerator
}

func NewOrderService(repo OrderRepository, publisher EventPublisher, idGen IDGenerator) *OrderService {
	return &OrderService{repo: repo, publisher: publisher, idGen: idGen}
}

func (s *OrderService) PlaceOrder(ctx context.Context, customerID string, items []LineItem) (Order, error) {
	order := Order{
		ID:         s.idGen.NewID(),
		CustomerID: customerID,
		Items:      items,
		Status:     "pending",
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return Order{}, fmt.Errorf("saving order: %w", err)
	}

	_ = s.publisher.Publish(ctx, OrderEvent{
		Type:       EventOrderPlaced,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		OccurredAt: time.Now(),
	})

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

	if err := s.repo.Save(ctx, order); err != nil {
		return err
	}

	_ = s.publisher.Publish(ctx, OrderEvent{
		Type:       EventOrderConfirmed,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		OccurredAt: time.Now(),
	})

	return nil
}

// --- In-memory adapters ---

type InMemoryRepo struct {
	mu     sync.RWMutex
	orders map[string]Order
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{orders: make(map[string]Order)}
}

func (r *InMemoryRepo) Save(_ context.Context, order Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryRepo) FindByID(_ context.Context, id string) (Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[id]
	if !ok {
		return Order{}, fmt.Errorf("order %s not found", id)
	}
	return order, nil
}

// InMemoryPublisher records events for testing.
type InMemoryPublisher struct {
	mu     sync.Mutex
	Events []OrderEvent
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{}
}

func (p *InMemoryPublisher) Publish(_ context.Context, event OrderEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Events = append(p.Events, event)
	return nil
}

// StaticIDGenerator returns a fixed ID for testing.
type StaticIDGenerator struct {
	ID string
}

func (g *StaticIDGenerator) NewID() string {
	return g.ID
}

// --- Inbound: event consumer as an inbound adapter ---

type OrderConfirmer interface {
	ConfirmOrder(ctx context.Context, id string) error
}

// EventConsumer is an inbound adapter that processes events.
// In production, this would read from Kafka/RabbitMQ/NATS.
type EventConsumer struct {
	confirmer OrderConfirmer
}

func NewEventConsumer(confirmer OrderConfirmer) *EventConsumer {
	return &EventConsumer{confirmer: confirmer}
}

// HandleEvent processes an incoming event. In production, the message broker
// adapter would call this method when a message arrives.
func (c *EventConsumer) HandleEvent(ctx context.Context, event OrderEvent) error {
	switch event.Type {
	case "payment.received":
		return c.confirmer.ConfirmOrder(ctx, event.OrderID)
	default:
		return nil // ignore unknown events
	}
}
