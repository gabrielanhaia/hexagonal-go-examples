package httphandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/adapter/httphandler"
	"github.com/gabrielanhaia/hexagonal-go-examples/orderservice/internal/domain"
)

// --- Test doubles ---

type stubService struct {
	n int
}

func (s *stubService) PlaceOrder(_ context.Context, req domain.PlaceOrderRequest) (domain.Order, error) {
	s.n++
	order, _ := domain.NewOrder(fmt.Sprintf("ord-%d", s.n), req.CustomerID, req.Items)
	return order, nil
}

func (s *stubService) GetOrder(_ context.Context, id string) (domain.Order, error) {
	price, _ := domain.NewMoney(1000, "USD")
	order, _ := domain.NewOrder(id, "cust-1", []domain.LineItem{
		{ProductID: "p1", Quantity: 1, Price: price},
	})
	return order, nil
}

func (s *stubService) ConfirmOrder(_ context.Context, _ string) error { return nil }
func (s *stubService) CancelOrder(_ context.Context, _ string) error  { return nil }

// --- Tests ---

func TestCreate(t *testing.T) {
	svc := &stubService{}
	h := httphandler.New(svc, svc, svc, svc)

	body := `{"customer_id":"cust-1","items":[{"product_id":"p1","quantity":2,"price_cents":2500,"currency":"USD"}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["id"] != "ord-1" {
		t.Errorf("id = %v, want ord-1", resp["id"])
	}
	if resp["status"] != "pending" {
		t.Errorf("status = %v, want pending", resp["status"])
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &stubService{}
	h := httphandler.New(svc, svc, svc, svc)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	h.Create().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
