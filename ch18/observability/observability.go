// Package observability demonstrates where logging, metrics, and tracing
// belong in a hexagonal architecture: in the adapters and middleware,
// never in the domain.
package observability

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// --- Domain (zero observability imports) ---

type Order struct {
	ID         string
	CustomerID string
	Status     string
}

type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	FindByID(ctx context.Context, id string) (Order, error)
	ListByCustomer(ctx context.Context, customerID string) ([]Order, error)
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) PlaceOrder(ctx context.Context, order Order) error {
	return s.repo.Save(ctx, order)
}

func (s *OrderService) ListOrders(ctx context.Context, customerID string) ([]Order, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

// --- Metrics port: the domain defines what can be recorded ---

type MetricsRecorder interface {
	RecordDuration(name string, duration time.Duration)
	IncCounter(name string)
}

// --- Logging decorator: wraps a port to add logging ---

type LoggingRepository struct {
	next   OrderRepository
	logger *slog.Logger
}

func NewLoggingRepository(next OrderRepository, logger *slog.Logger) *LoggingRepository {
	return &LoggingRepository{next: next, logger: logger}
}

func (r *LoggingRepository) Save(ctx context.Context, order Order) error {
	start := time.Now()
	err := r.next.Save(ctx, order)
	r.logger.Info("repository.Save",
		"order_id", order.ID,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err,
	)
	return err
}

func (r *LoggingRepository) FindByID(ctx context.Context, id string) (Order, error) {
	start := time.Now()
	order, err := r.next.FindByID(ctx, id)
	r.logger.Info("repository.FindByID",
		"order_id", id,
		"found", err == nil,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return order, err
}

func (r *LoggingRepository) ListByCustomer(ctx context.Context, customerID string) ([]Order, error) {
	start := time.Now()
	orders, err := r.next.ListByCustomer(ctx, customerID)
	r.logger.Info("repository.ListByCustomer",
		"customer_id", customerID,
		"count", len(orders),
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err,
	)
	return orders, err
}

// --- Metrics decorator: wraps a port to record metrics ---

type MetricsRepository struct {
	next    OrderRepository
	metrics MetricsRecorder
}

func NewMetricsRepository(next OrderRepository, metrics MetricsRecorder) *MetricsRepository {
	return &MetricsRepository{next: next, metrics: metrics}
}

func (r *MetricsRepository) Save(ctx context.Context, order Order) error {
	start := time.Now()
	err := r.next.Save(ctx, order)
	r.metrics.RecordDuration("repository.save", time.Since(start))
	r.metrics.IncCounter("repository.save")
	if err != nil {
		r.metrics.IncCounter("repository.save.error")
	}
	return err
}

func (r *MetricsRepository) FindByID(ctx context.Context, id string) (Order, error) {
	start := time.Now()
	order, err := r.next.FindByID(ctx, id)
	r.metrics.RecordDuration("repository.find", time.Since(start))
	r.metrics.IncCounter("repository.find")
	if err != nil {
		r.metrics.IncCounter("repository.find.error")
	}
	return order, err
}

func (r *MetricsRepository) ListByCustomer(ctx context.Context, customerID string) ([]Order, error) {
	start := time.Now()
	orders, err := r.next.ListByCustomer(ctx, customerID)
	r.metrics.RecordDuration("repository.list_by_customer", time.Since(start))
	r.metrics.IncCounter("repository.list_by_customer")
	if err != nil {
		r.metrics.IncCounter("repository.list_by_customer.error")
	}
	return orders, err
}

// --- In-memory repo for testing ---

type InMemoryRepo struct {
	orders map[string]Order
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{orders: make(map[string]Order)}
}

func (r *InMemoryRepo) Save(_ context.Context, order Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryRepo) FindByID(_ context.Context, id string) (Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return Order{}, fmt.Errorf("order %s not found", id)
	}
	return order, nil
}

func (r *InMemoryRepo) ListByCustomer(_ context.Context, customerID string) ([]Order, error) {
	var result []Order
	for _, o := range r.orders {
		if o.CustomerID == customerID {
			result = append(result, o)
		}
	}
	return result, nil
}

// --- In-memory metrics recorder for testing ---

type InMemoryMetrics struct {
	Counters  map[string]int
	Durations map[string]time.Duration
}

func NewInMemoryMetrics() *InMemoryMetrics {
	return &InMemoryMetrics{
		Counters:  make(map[string]int),
		Durations: make(map[string]time.Duration),
	}
}

func (m *InMemoryMetrics) RecordDuration(name string, d time.Duration) {
	m.Durations[name] += d
}

func (m *InMemoryMetrics) IncCounter(name string) {
	m.Counters[name]++
}
