// Package interfaces demonstrates Go's implicit interface satisfaction —
// the language feature that makes hexagonal architecture feel native in Go.
package interfaces

import "fmt"

// OrderRepository is a port — it describes what the domain needs,
// not how it's implemented.
type OrderRepository interface {
	Save(order Order) error
	FindByID(id string) (Order, error)
}

// NotificationSender is another port — the domain needs to send
// notifications but doesn't care whether it's email, SMS, or Slack.
type NotificationSender interface {
	Send(to string, message string) error
}

// Order is a simple domain entity.
type Order struct {
	ID         string
	CustomerID string
	TotalCents int64
	Status     string
}

// PostgresOrderRepository implements OrderRepository using PostgreSQL.
// Notice: no "implements" keyword. It satisfies the interface by having
// the right methods with the right signatures. That's it.
type PostgresOrderRepository struct {
	// In a real app, this would hold a *sql.DB
	connString string
}

func NewPostgresOrderRepository(connString string) *PostgresOrderRepository {
	return &PostgresOrderRepository{connString: connString}
}

func (r *PostgresOrderRepository) Save(order Order) error {
	fmt.Printf("PostgreSQL: saving order %s\n", order.ID)
	return nil
}

func (r *PostgresOrderRepository) FindByID(id string) (Order, error) {
	fmt.Printf("PostgreSQL: finding order %s\n", id)
	return Order{ID: id, Status: "pending"}, nil
}

// InMemoryOrderRepository also implements OrderRepository — useful for tests.
// Same interface, completely different implementation.
type InMemoryOrderRepository struct {
	orders map[string]Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{orders: make(map[string]Order)}
}

func (r *InMemoryOrderRepository) Save(order Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(id string) (Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return Order{}, fmt.Errorf("order %s not found", id)
	}
	return order, nil
}

// OrderService is the domain logic. It depends on the PORT (interface),
// not on any specific adapter. It has no idea whether it's talking to
// PostgreSQL, an in-memory map, or a mock.
type OrderService struct {
	repo   OrderRepository
	notify NotificationSender
}

func NewOrderService(repo OrderRepository, notify NotificationSender) *OrderService {
	return &OrderService{repo: repo, notify: notify}
}

func (s *OrderService) PlaceOrder(customerID string, totalCents int64) (Order, error) {
	order := Order{
		ID:         fmt.Sprintf("ord_%s", customerID),
		CustomerID: customerID,
		TotalCents: totalCents,
		Status:     "pending",
	}

	if err := s.repo.Save(order); err != nil {
		return Order{}, fmt.Errorf("saving order: %w", err)
	}

	_ = s.notify.Send(customerID, fmt.Sprintf("Order %s placed!", order.ID))

	return order, nil
}

// ConsoleNotifier is a simple adapter for NotificationSender.
type ConsoleNotifier struct{}

func (n *ConsoleNotifier) Send(to string, message string) error {
	fmt.Printf("Notification to %s: %s\n", to, message)
	return nil
}
