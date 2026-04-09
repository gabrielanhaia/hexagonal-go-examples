package interfaces

import (
	"testing"
)

// SpyNotifier records calls for test assertions.
type SpyNotifier struct {
	Calls []struct {
		To      string
		Message string
	}
}

func (n *SpyNotifier) Send(to string, message string) error {
	n.Calls = append(n.Calls, struct {
		To      string
		Message string
	}{to, message})
	return nil
}

func TestPlaceOrder_SavesAndNotifies(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	notifier := &SpyNotifier{}
	service := NewOrderService(repo, notifier)

	order, err := service.PlaceOrder("cust_123", 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.CustomerID != "cust_123" {
		t.Errorf("got customer %q, want %q", order.CustomerID, "cust_123")
	}
	if order.TotalCents != 5000 {
		t.Errorf("got total %d, want %d", order.TotalCents, 5000)
	}
	if order.Status != "pending" {
		t.Errorf("got status %q, want %q", order.Status, "pending")
	}

	// Verify the order was actually saved
	saved, err := repo.FindByID(order.ID)
	if err != nil {
		t.Fatalf("order not found in repo: %v", err)
	}
	if saved.ID != order.ID {
		t.Errorf("saved order ID %q doesn't match %q", saved.ID, order.ID)
	}

	// Verify notification was sent
	if len(notifier.Calls) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.Calls))
	}
	if notifier.Calls[0].To != "cust_123" {
		t.Errorf("notification sent to %q, want %q", notifier.Calls[0].To, "cust_123")
	}
}

func TestPlaceOrder_WithPostgresRepo(t *testing.T) {
	// This demonstrates that PostgresOrderRepository satisfies OrderRepository
	// without any explicit "implements" declaration.
	// We're just proving the types are compatible — not hitting a real DB.
	var repo OrderRepository = NewPostgresOrderRepository("host=localhost")
	_ = repo // Compiles. That's the proof.
}

func TestInMemoryRepo_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryOrderRepository()

	_, err := repo.FindByID("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent order, got nil")
	}
}
