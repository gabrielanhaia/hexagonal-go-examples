// Package main demonstrates a typical "spaghetti" Go service where HTTP handling,
// business logic, and database access are tangled together in a single handler.
// This is the BEFORE example — what NOT to do. The rest of the book shows how to fix it.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=app dbname=orders sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/orders", createOrderHandler)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// --- Parse the request ---
	var req struct {
		CustomerID string `json:"customer_id"`
		Items      []struct {
			ProductID string  `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
		CouponCode string `json:"coupon_code,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// --- Validate ---
	if req.CustomerID == "" {
		http.Error(w, "customer_id is required",
			http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "at least one item is required",
			http.StatusBadRequest)
		return
	}

	// --- Calculate totals and discounts ---
	var subtotal float64
	for _, item := range req.Items {
		subtotal += item.Price * float64(item.Quantity)
	}

	discount := 0.0
	if req.CouponCode != "" {
		var pct float64
		err := db.QueryRowContext(r.Context(),
			"SELECT discount_pct FROM coupons WHERE code = $1 "+
				"AND expires_at > NOW()",
			req.CouponCode,
		).Scan(&pct)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "database error",
				http.StatusInternalServerError)
			return
		}
		discount = subtotal * (pct / 100)
	}

	total := subtotal - discount

	// --- Insert the order ---
	orderID := uuid.New().String()
	_, err := db.ExecContext(r.Context(),
		"INSERT INTO orders (id, customer_id, subtotal, "+
			"discount, total, created_at) "+
			"VALUES ($1, $2, $3, $4, $5, NOW())",
		orderID, req.CustomerID, subtotal, discount, total,
	)
	if err != nil {
		http.Error(w, "failed to create order",
			http.StatusInternalServerError)
		return
	}

	// --- Insert line items ---
	for _, item := range req.Items {
		_, err := db.ExecContext(r.Context(),
			"INSERT INTO order_items (order_id, product_id, "+
				"quantity, price) VALUES ($1, $2, $3, $4)",
			orderID, item.ProductID, item.Quantity, item.Price,
		)
		if err != nil {
			log.Printf("failed to insert item: %v", err)
		}
	}

	// --- Send notification ---
	notifBody, _ := json.Marshal(map[string]string{
		"order_id":    orderID,
		"customer_id": req.CustomerID,
	})
	resp, err := http.Post(
		"http://notification-service:8080/notify",
		"application/json",
		bytes.NewReader(notifBody),
	)
	if err != nil {
		log.Printf("notification failed: %v", err)
	} else {
		resp.Body.Close()
	}

	// --- Respond ---
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": orderID,
	})
}
