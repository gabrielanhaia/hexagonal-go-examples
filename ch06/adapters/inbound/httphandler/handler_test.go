package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielanhaia/hexagonal-go-examples/ch06/adapters"
)

// stubPlacer is a test double that satisfies the OrderPlacer port.
type stubPlacer struct {
	lastCustomerID string
	lastItems      []adapters.OrderItem
}

func (s *stubPlacer) PlaceOrder(_ context.Context, customerID string, items []adapters.OrderItem) (*adapters.Order, error) {
	s.lastCustomerID = customerID
	s.lastItems = items
	return &adapters.Order{
		ID:         "ord_test",
		CustomerID: customerID,
		Items:      items,
		Status:     adapters.OrderStatusPending,
		Total:      adapters.Money{Cents: 5000, Currency: "USD"},
	}, nil
}

func TestCreateOrder(t *testing.T) {
	placer := &stubPlacer{}
	handler := NewOrderHandler(placer)

	body := `{"customer_id":"cust_1","items":[{"product_id":"prod_a","quantity":2,"price_cents":2500}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp createOrderResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.OrderID != "ord_test" {
		t.Errorf("order_id = %q, want ord_test", resp.OrderID)
	}
	if resp.Total != 5000 {
		t.Errorf("total = %d, want 5000", resp.Total)
	}

	// Verify the adapter translated the HTTP request correctly
	if placer.lastCustomerID != "cust_1" {
		t.Errorf("customer = %q, want cust_1", placer.lastCustomerID)
	}
	if len(placer.lastItems) != 1 {
		t.Fatalf("items = %d, want 1", len(placer.lastItems))
	}
	if placer.lastItems[0].ProductID != "prod_a" {
		t.Errorf("product = %q, want prod_a", placer.lastItems[0].ProductID)
	}
}

func TestCreateOrder_WrongMethod(t *testing.T) {
	handler := NewOrderHandler(&stubPlacer{})

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rec := httptest.NewRecorder()

	handler.Create().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCreateOrder_InvalidJSON(t *testing.T) {
	handler := NewOrderHandler(&stubPlacer{})

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	handler.Create().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
