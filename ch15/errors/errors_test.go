package errors

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestGetOrder_NotFound(t *testing.T) {
	repo := NewInMemoryRepo()
	svc := NewOrderService(repo)

	_, err := svc.GetOrder(context.Background(), "nonexistent")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestGetOrder_Found(t *testing.T) {
	repo := NewInMemoryRepo()
	repo.Add(Order{ID: "ord-1", CustomerID: "cust-1", Status: "pending"})
	svc := NewOrderService(repo)

	order, err := svc.GetOrder(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != "ord-1" {
		t.Errorf("id = %q, want ord-1", order.ID)
	}
}

// domainErrorToStatus demonstrates error mapping in an HTTP adapter.
func domainErrorToStatus(err error) int {
	switch {
	case errors.Is(err, ErrOrderNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrOrderEmpty):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotPending):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{ErrOrderNotFound, http.StatusNotFound},
		{ErrOrderEmpty, http.StatusBadRequest},
		{ErrNotPending, http.StatusConflict},
		{errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		got := domainErrorToStatus(tt.err)
		if got != tt.status {
			t.Errorf("domainErrorToStatus(%v) = %d, want %d", tt.err, got, tt.status)
		}
	}
}
