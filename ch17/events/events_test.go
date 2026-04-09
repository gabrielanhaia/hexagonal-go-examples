package events

import (
	"context"
	"testing"
)

func TestPlaceOrder_PublishesEvent(t *testing.T) {
	repo := NewInMemoryRepo()
	pub := NewInMemoryPublisher()
	idGen := &StaticIDGenerator{ID: "ord-1"}
	svc := NewOrderService(repo, pub, idGen)

	order, err := svc.PlaceOrder(context.Background(), "cust-1", []LineItem{
		{ProductID: "prod-a", Quantity: 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != "ord-1" {
		t.Errorf("order id = %q, want ord-1", order.ID)
	}

	if len(pub.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.Events))
	}
	if pub.Events[0].Type != EventOrderPlaced {
		t.Errorf("event type = %q, want %s", pub.Events[0].Type, EventOrderPlaced)
	}
	if pub.Events[0].OrderID != "ord-1" {
		t.Errorf("order id = %q, want ord-1", pub.Events[0].OrderID)
	}
}

func TestConfirmOrder_PublishesEvent(t *testing.T) {
	repo := NewInMemoryRepo()
	pub := NewInMemoryPublisher()
	idGen := &StaticIDGenerator{ID: "ord-1"}
	svc := NewOrderService(repo, pub, idGen)

	svc.PlaceOrder(context.Background(), "cust-1", []LineItem{
		{ProductID: "prod-a", Quantity: 1},
	})

	err := svc.ConfirmOrder(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(pub.Events))
	}
	if pub.Events[1].Type != EventOrderConfirmed {
		t.Errorf("event type = %q, want %s", pub.Events[1].Type, EventOrderConfirmed)
	}
}

func TestConfirmOrder_RejectsNonPending(t *testing.T) {
	repo := NewInMemoryRepo()
	pub := NewInMemoryPublisher()
	idGen := &StaticIDGenerator{ID: "ord-1"}
	svc := NewOrderService(repo, pub, idGen)

	svc.PlaceOrder(context.Background(), "cust-1", []LineItem{
		{ProductID: "prod-a", Quantity: 1},
	})
	svc.ConfirmOrder(context.Background(), "ord-1")

	// Confirming again should fail because status is now "confirmed"
	err := svc.ConfirmOrder(context.Background(), "ord-1")
	if err == nil {
		t.Fatal("expected error confirming already-confirmed order")
	}
}

func TestEventConsumer_PaymentReceived(t *testing.T) {
	repo := NewInMemoryRepo()
	pub := NewInMemoryPublisher()
	idGen := &StaticIDGenerator{ID: "ord-1"}
	svc := NewOrderService(repo, pub, idGen)

	svc.PlaceOrder(context.Background(), "cust-1", []LineItem{
		{ProductID: "prod-a", Quantity: 1},
	})

	consumer := NewEventConsumer(svc)
	err := consumer.HandleEvent(context.Background(), OrderEvent{
		Type:    "payment.received",
		OrderID: "ord-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order, _ := repo.FindByID(context.Background(), "ord-1")
	if order.Status != "confirmed" {
		t.Errorf("status = %q, want confirmed", order.Status)
	}
}

func TestEventConsumer_UnknownEvent(t *testing.T) {
	consumer := NewEventConsumer(nil) // no confirmer needed
	err := consumer.HandleEvent(context.Background(), OrderEvent{Type: "unknown.event"})
	if err != nil {
		t.Fatalf("unexpected error for unknown event: %v", err)
	}
}
