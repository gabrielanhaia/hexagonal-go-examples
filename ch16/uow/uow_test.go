package uow

import (
	"context"
	"testing"
)

func TestPlaceOrder_Success(t *testing.T) {
	uow := NewInMemoryUoW()
	svc := NewPlaceOrderService(uow)

	items := []LineItem{
		{ProductID: "prod-a", Quantity: 2},
		{ProductID: "prod-b", Quantity: 1},
	}

	order, err := svc.PlaceOrder(context.Background(), "cust-1", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Order should have a generated ID and correct fields
	if order.ID == "" {
		t.Error("order ID should not be empty")
	}
	if order.CustomerID != "cust-1" {
		t.Errorf("customer id = %q, want cust-1", order.CustomerID)
	}
	if order.Status != "pending" {
		t.Errorf("status = %q, want pending", order.Status)
	}

	// Both order and inventory committed
	if _, ok := uow.Orders()[order.ID]; !ok {
		t.Error("order not saved")
	}
	if uow.Inventory()["prod-a"] != 2 {
		t.Errorf("prod-a reserved = %d, want 2", uow.Inventory()["prod-a"])
	}
	if uow.Inventory()["prod-b"] != 1 {
		t.Errorf("prod-b reserved = %d, want 1", uow.Inventory()["prod-b"])
	}
}

func TestPlaceOrder_RollbackOnReservationFailure(t *testing.T) {
	uow := NewInMemoryUoW()
	svc := NewPlaceOrderService(uow)

	items := []LineItem{
		{ProductID: "prod-a", Quantity: 2},
		{ProductID: "prod-b", Quantity: -1}, // invalid — will fail
	}

	_, err := svc.PlaceOrder(context.Background(), "cust-1", items)
	if err == nil {
		t.Fatal("expected error for invalid reservation")
	}

	// Nothing committed — transaction rolled back
	if len(uow.Orders()) != 0 {
		t.Error("no orders should be saved after rollback")
	}
	if len(uow.Inventory()) != 0 {
		t.Errorf("inventory = %v, want empty after rollback", uow.Inventory())
	}
}
