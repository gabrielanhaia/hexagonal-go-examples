package domain_test

import (
	"testing"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

func TestNewOrder_RequiresItems(t *testing.T) {
	_, err := domain.NewOrder("ord-1", "cust-1", nil)
	if err != domain.ErrOrderEmpty {
		t.Fatalf("expected ErrOrderEmpty, got %v", err)
	}
}

func TestOrder_Total(t *testing.T) {
	price, _ := domain.NewMoney(1500, "USD")
	items := []domain.LineItem{
		{ProductID: "prod-1", Quantity: 2, Price: price},
		{ProductID: "prod-2", Quantity: 1, Price: price},
	}
	order, _ := domain.NewOrder("ord-1", "cust-1", items)

	total, err := order.Total()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total.Amount() != 4500 {
		t.Fatalf("expected 4500, got %d", total.Amount())
	}
}

func TestOrder_Confirm(t *testing.T) {
	order := validOrder(t)
	if err := order.Confirm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != domain.OrderStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", order.Status)
	}
}

func TestOrder_ConfirmOnlyFromPending(t *testing.T) {
	order := validOrder(t)
	_ = order.Confirm()
	err := order.Confirm()
	if err != domain.ErrOrderNotPending {
		t.Fatalf("expected ErrOrderNotPending, got %v", err)
	}
}

func TestOrder_CancelAlreadyCanceled(t *testing.T) {
	order := validOrder(t)
	_ = order.Cancel()
	err := order.Cancel()
	if err != domain.ErrOrderCanceled {
		t.Fatalf("expected ErrOrderCanceled, got %v", err)
	}
}

func TestOrder_ApplyDiscount(t *testing.T) {
	price, _ := domain.NewMoney(2000, "USD")
	items := []domain.LineItem{{ProductID: "p1", Quantity: 1, Price: price}}
	order, _ := domain.NewOrder("ord-1", "cust-1", items)

	if err := order.ApplyDiscount(25); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total, _ := order.Total()
	if total.Amount() != 1500 {
		t.Fatalf("expected 1500, got %d", total.Amount())
	}
}

func TestMoney_Add_CurrencyMismatch(t *testing.T) {
	usd, _ := domain.NewMoney(1000, "USD")
	eur, _ := domain.NewMoney(500, "EUR")
	_, err := usd.Add(eur)
	if err == nil {
		t.Fatal("expected currency mismatch error")
	}
}

func validOrder(t *testing.T) domain.Order {
	t.Helper()
	price, _ := domain.NewMoney(1000, "USD")
	items := []domain.LineItem{{ProductID: "prod-1", Quantity: 2, Price: price}}
	order, err := domain.NewOrder("ord-1", "cust-1", items)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	return order
}
