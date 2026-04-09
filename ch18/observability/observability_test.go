package observability

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestLoggingDecorator(t *testing.T) {
	inner := NewInMemoryRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	repo := NewLoggingRepository(inner, logger)

	svc := NewOrderService(repo)

	err := svc.PlaceOrder(context.Background(), Order{ID: "ord-1", CustomerID: "cust-1", Status: "pending"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order, err := svc.GetOrder(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != "ord-1" {
		t.Errorf("id = %q, want ord-1", order.ID)
	}
}

func TestLoggingDecorator_ListByCustomer(t *testing.T) {
	inner := NewInMemoryRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	repo := NewLoggingRepository(inner, logger)

	svc := NewOrderService(repo)

	svc.PlaceOrder(context.Background(), Order{ID: "ord-1", CustomerID: "cust-1", Status: "pending"})
	svc.PlaceOrder(context.Background(), Order{ID: "ord-2", CustomerID: "cust-1", Status: "pending"})

	orders, err := svc.ListOrders(context.Background(), "cust-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("orders = %d, want 2", len(orders))
	}
}

func TestMetricsDecorator(t *testing.T) {
	inner := NewInMemoryRepo()
	metrics := NewInMemoryMetrics()
	repo := NewMetricsRepository(inner, metrics)

	svc := NewOrderService(repo)

	svc.PlaceOrder(context.Background(), Order{ID: "ord-1", CustomerID: "cust-1", Status: "pending"})
	svc.PlaceOrder(context.Background(), Order{ID: "ord-2", CustomerID: "cust-2", Status: "pending"})
	svc.GetOrder(context.Background(), "ord-1")
	svc.GetOrder(context.Background(), "nonexistent")

	if metrics.Counters["repository.save"] != 2 {
		t.Errorf("save count = %d, want 2", metrics.Counters["repository.save"])
	}
	if metrics.Counters["repository.find"] != 2 {
		t.Errorf("find count = %d, want 2", metrics.Counters["repository.find"])
	}
	if metrics.Counters["repository.find.error"] != 1 {
		t.Errorf("find error count = %d, want 1", metrics.Counters["repository.find.error"])
	}
}

func TestMetricsDecorator_ListByCustomer(t *testing.T) {
	inner := NewInMemoryRepo()
	metrics := NewInMemoryMetrics()
	repo := NewMetricsRepository(inner, metrics)

	svc := NewOrderService(repo)

	svc.PlaceOrder(context.Background(), Order{ID: "ord-1", CustomerID: "cust-1", Status: "pending"})
	svc.ListOrders(context.Background(), "cust-1")

	if metrics.Counters["repository.list_by_customer"] != 1 {
		t.Errorf("list count = %d, want 1", metrics.Counters["repository.list_by_customer"])
	}
}

func TestDecoratorChaining(t *testing.T) {
	inner := NewInMemoryRepo()
	metrics := NewInMemoryMetrics()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Chain: metrics → logging → actual repo
	repo := NewMetricsRepository(NewLoggingRepository(inner, logger), metrics)

	svc := NewOrderService(repo)
	svc.PlaceOrder(context.Background(), Order{ID: "ord-1", CustomerID: "cust-1", Status: "pending"})

	if metrics.Counters["repository.save"] != 1 {
		t.Errorf("save count = %d, want 1", metrics.Counters["repository.save"])
	}
}
