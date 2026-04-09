package domain_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

// --- Test doubles ---

type inMemoryRepo struct {
	orders map[string]domain.Order
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{orders: make(map[string]domain.Order)}
}

func (r *inMemoryRepo) Save(_ context.Context, order domain.Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *inMemoryRepo) FindByID(_ context.Context, id string) (domain.Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

type spyNotifier struct {
	calls []domain.Order
}

func (n *spyNotifier) OrderPlaced(_ context.Context, order domain.Order) error {
	n.calls = append(n.calls, order)
	return nil
}

type seqIDGen struct{ n int }

func (g *seqIDGen) NewID() string {
	g.n++
	return fmt.Sprintf("ord-%d", g.n)
}

// --- Tests ---

func TestPlaceOrder(t *testing.T) {
	repo := newInMemoryRepo()
	notifier := &spyNotifier{}
	idgen := &seqIDGen{}
	svc := domain.NewOrderService(repo, notifier, idgen)

	price, _ := domain.NewMoney(2500, "USD")
	req := domain.PlaceOrderRequest{
		CustomerID: "cust-1",
		Items:      []domain.LineItem{{ProductID: "prod-a", Quantity: 2, Price: price}},
	}

	order, err := svc.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.ID != "ord-1" {
		t.Errorf("id = %q, want ord-1", order.ID)
	}
	if order.Status != domain.OrderStatusPending {
		t.Errorf("status = %q, want pending", order.Status)
	}

	saved, err := repo.FindByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("not saved: %v", err)
	}
	if saved.CustomerID != "cust-1" {
		t.Errorf("saved customer = %q, want cust-1", saved.CustomerID)
	}

	if len(notifier.calls) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.calls))
	}
}

func TestPlaceOrder_EmptyItems(t *testing.T) {
	svc := domain.NewOrderService(newInMemoryRepo(), &spyNotifier{}, &seqIDGen{})

	_, err := svc.PlaceOrder(context.Background(), domain.PlaceOrderRequest{
		CustomerID: "cust-1",
		Items:      nil,
	})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

func TestGetOrder(t *testing.T) {
	repo := newInMemoryRepo()
	svc := domain.NewOrderService(repo, &spyNotifier{}, &seqIDGen{})

	price, _ := domain.NewMoney(1000, "USD")
	placed, _ := svc.PlaceOrder(context.Background(), domain.PlaceOrderRequest{
		CustomerID: "cust-1",
		Items:      []domain.LineItem{{ProductID: "p1", Quantity: 1, Price: price}},
	})

	got, err := svc.GetOrder(context.Background(), placed.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != placed.ID {
		t.Errorf("got ID %q, want %q", got.ID, placed.ID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	svc := domain.NewOrderService(newInMemoryRepo(), &spyNotifier{}, &seqIDGen{})
	_, err := svc.GetOrder(context.Background(), "nonexistent")
	if err != domain.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestConfirmOrder(t *testing.T) {
	repo := newInMemoryRepo()
	svc := domain.NewOrderService(repo, &spyNotifier{}, &seqIDGen{})

	price, _ := domain.NewMoney(1000, "USD")
	order, _ := svc.PlaceOrder(context.Background(), domain.PlaceOrderRequest{
		CustomerID: "cust-1",
		Items:      []domain.LineItem{{ProductID: "p1", Quantity: 1, Price: price}},
	})

	err := svc.ConfirmOrder(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	confirmed, _ := svc.GetOrder(context.Background(), order.ID)
	if confirmed.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want confirmed", confirmed.Status)
	}
}

func TestCancelOrder(t *testing.T) {
	repo := newInMemoryRepo()
	svc := domain.NewOrderService(repo, &spyNotifier{}, &seqIDGen{})

	price, _ := domain.NewMoney(1000, "USD")
	order, _ := svc.PlaceOrder(context.Background(), domain.PlaceOrderRequest{
		CustomerID: "cust-1",
		Items:      []domain.LineItem{{ProductID: "p1", Quantity: 1, Price: price}},
	})

	err := svc.CancelOrder(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	canceled, _ := svc.GetOrder(context.Background(), order.ID)
	if canceled.Status != domain.OrderStatusCanceled {
		t.Errorf("status = %q, want canceled", canceled.Status)
	}
}
