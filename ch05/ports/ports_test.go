package ports

import (
	"context"
	"fmt"
	"testing"
)

// --- Test doubles that satisfy the ports ---

type inMemoryRepo struct {
	orders map[string]Order
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{orders: make(map[string]Order)}
}

func (r *inMemoryRepo) Save(_ context.Context, order Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *inMemoryRepo) FindByID(_ context.Context, id string) (Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return Order{}, fmt.Errorf("order %s not found", id)
	}
	return order, nil
}

type spyNotifier struct {
	calls []Order
}

func (n *spyNotifier) OrderPlaced(_ context.Context, order Order) error {
	n.calls = append(n.calls, order)
	return nil
}

type stubPricer struct {
	prices map[string]Money
}

func (p *stubPricer) GetPrice(_ context.Context, productID string) (Money, error) {
	price, ok := p.prices[productID]
	if !ok {
		return Money{}, fmt.Errorf("product %s not found", productID)
	}
	return price, nil
}

// --- Tests demonstrating port-based testing ---

func TestPlaceOrder(t *testing.T) {
	repo := newInMemoryRepo()
	notifier := &spyNotifier{}
	pricer := &stubPricer{
		prices: map[string]Money{
			"prod_a": {Cents: 1000, Currency: "USD"},
			"prod_b": {Cents: 2500, Currency: "USD"},
		},
	}

	svc := NewOrderService(repo, notifier, pricer)

	order, err := svc.PlaceOrder(context.Background(), "cust_1", []string{"prod_a", "prod_b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.Total.Cents != 3500 {
		t.Errorf("total = %d, want 3500", order.Total.Cents)
	}

	if len(order.Items) != 2 {
		t.Errorf("items = %d, want 2", len(order.Items))
	}

	// Verify it was saved
	saved, err := repo.FindByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("order not in repo: %v", err)
	}
	if saved.CustomerID != "cust_1" {
		t.Errorf("saved customer = %q, want cust_1", saved.CustomerID)
	}

	// Verify notification was sent
	if len(notifier.calls) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.calls))
	}
}

func TestPlaceOrder_UnknownProduct(t *testing.T) {
	repo := newInMemoryRepo()
	notifier := &spyNotifier{}
	pricer := &stubPricer{prices: map[string]Money{}}

	svc := NewOrderService(repo, notifier, pricer)

	_, err := svc.PlaceOrder(context.Background(), "cust_1", []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for unknown product")
	}
}

func TestGetOrder(t *testing.T) {
	repo := newInMemoryRepo()
	notifier := &spyNotifier{}
	pricer := &stubPricer{
		prices: map[string]Money{"prod_a": {Cents: 1000, Currency: "USD"}},
	}

	svc := NewOrderService(repo, notifier, pricer)

	placed, _ := svc.PlaceOrder(context.Background(), "cust_1", []string{"prod_a"})

	got, err := svc.GetOrder(context.Background(), placed.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != placed.ID {
		t.Errorf("got ID %q, want %q", got.ID, placed.ID)
	}
}

// This test demonstrates that OrderSaver works as a standalone port.
// A service that only writes orders can depend on just OrderSaver.
func TestOrderSaverPort(t *testing.T) {
	repo := newInMemoryRepo()

	// Accept the narrow interface — proves a write-only consumer works
	var saver OrderSaver = repo
	err := saver.Save(context.Background(), Order{ID: "ord_99", CustomerID: "cust_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// But we can still read through the wider interface
	_, err = repo.FindByID(context.Background(), "ord_99")
	if err != nil {
		t.Error("order should be findable through the full repo")
	}
}
