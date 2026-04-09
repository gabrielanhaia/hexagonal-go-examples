package correct

import (
	"context"
	"testing"
)

func TestCreateOrder_Correct(t *testing.T) {
	// The domain service accepts any OrderRepository implementation.
	// We pass an in-memory adapter for testing. In production, we'd pass Postgres.
	// The service doesn't know or care which one it gets.
	repo := NewInMemoryRepository()
	svc := NewOrderService(repo)

	order, err := svc.CreateOrder(context.Background(), "cust_1", 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.CustomerID != "cust_1" {
		t.Errorf("customer = %q, want cust_1", order.CustomerID)
	}

	// The adapter stored the order — we can verify through the adapter directly
	stored, ok := repo.Orders[order.ID]
	if !ok {
		t.Fatal("order not found in repository")
	}
	if stored.TotalCents != 5000 {
		t.Errorf("stored total = %d, want 5000", stored.TotalCents)
	}
}

func TestCreateOrder_Validation(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewOrderService(repo)

	_, err := svc.CreateOrder(context.Background(), "", 5000)
	if err == nil {
		t.Error("expected error for empty customer ID")
	}

	_, err = svc.CreateOrder(context.Background(), "cust_1", -100)
	if err == nil {
		t.Error("expected error for negative total")
	}
}
