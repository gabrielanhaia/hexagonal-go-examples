// Package httphandler provides an inbound HTTP adapter that translates
// HTTP requests into domain operations and domain responses into HTTP responses.
// The handler knows about HTTP. The service it calls does not.
package httphandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gabrielanhaia/hexagonal-go-examples/ch06/adapters"
)

// OrderPlacer is the port this inbound adapter calls into.
// It's defined here, at the adapter boundary, because the adapter
// is the consumer. It doesn't need the full service — just this one method.
type OrderPlacer interface {
	PlaceOrder(ctx context.Context, customerID string, items []adapters.OrderItem) (*adapters.Order, error)
}

type OrderHandler struct {
	placer OrderPlacer
}

func NewOrderHandler(placer OrderPlacer) *OrderHandler {
	return &OrderHandler{placer: placer}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id"`
	Items      []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
		Price     int64  `json:"price_cents"`
	} `json:"items"`
}

type createOrderResponse struct {
	OrderID string `json:"order_id"`
	Total   int64  `json:"total_cents"`
	Status  string `json:"status"`
}

func (h *OrderHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Translate HTTP request into domain types
		var items []adapters.OrderItem
		for _, item := range req.Items {
			items = append(items, adapters.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     adapters.Money{Cents: item.Price, Currency: "USD"},
			})
		}

		// Call the domain through the port
		order, err := h.placer.PlaceOrder(r.Context(), req.CustomerID, items)
		if err != nil {
			// Translate domain errors into HTTP responses
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Translate domain response into HTTP response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createOrderResponse{
			OrderID: order.ID,
			Total:   order.Total.Cents,
			Status:  string(order.Status),
		})
	}
}
