// Package uow demonstrates the Unit of Work pattern for managing transactions
// without leaking infrastructure into the domain.
package uow

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// --- Domain types ---

type LineItem struct {
	ProductID string
	Quantity  int
}

type Order struct {
	ID         string
	CustomerID string
	Items      []LineItem
	Status     string
}

// generateID produces a new unique identifier.
func generateID() string {
	return uuid.New().String()
}

// --- Ports ---

type OrderSaver interface {
	SaveOrder(ctx context.Context, order Order) error
}

type InventoryReserver interface {
	ReserveInventory(ctx context.Context, productID string, qty int) error
}

// UnitOfWork is the port that provides transactional execution.
// The domain says "run these operations as one atomic unit"
// without knowing about SQL transactions, Redis pipelines, etc.
type UnitOfWork interface {
	Execute(ctx context.Context, fn func(tx UnitOfWorkTx) error) error
}

// UnitOfWorkTx provides transactional versions of the ports.
type UnitOfWorkTx interface {
	OrderSaver
	InventoryReserver
}

// --- Domain service using UoW ---

type PlaceOrderService struct {
	uow UnitOfWork
}

func NewPlaceOrderService(uow UnitOfWork) *PlaceOrderService {
	return &PlaceOrderService{uow: uow}
}

func (s *PlaceOrderService) PlaceOrder(ctx context.Context, customerID string, items []LineItem) (Order, error) {
	order := Order{
		ID:         generateID(),
		CustomerID: customerID,
		Items:      items,
		Status:     "pending",
	}

	err := s.uow.Execute(ctx, func(tx UnitOfWorkTx) error {
		if err := tx.SaveOrder(ctx, order); err != nil {
			return fmt.Errorf("saving order: %w", err)
		}
		for _, item := range items {
			if err := tx.ReserveInventory(ctx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("reserving inventory: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// --- In-memory UoW adapter for testing ---

type InMemoryUoW struct {
	mu        sync.Mutex
	orders    map[string]Order
	inventory map[string]int // product → reserved quantity
}

func NewInMemoryUoW() *InMemoryUoW {
	return &InMemoryUoW{
		orders:    make(map[string]Order),
		inventory: make(map[string]int),
	}
}

func (u *InMemoryUoW) Execute(_ context.Context, fn func(tx UnitOfWorkTx) error) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Create a transactional view — on error, changes are discarded
	tx := &inMemoryTx{
		orders:    make(map[string]Order),
		inventory: make(map[string]int),
	}

	if err := fn(tx); err != nil {
		return err // discard changes
	}

	// Commit: apply buffered changes
	for id, order := range tx.orders {
		u.orders[id] = order
	}
	for product, qty := range tx.inventory {
		u.inventory[product] += qty
	}
	return nil
}

func (u *InMemoryUoW) Orders() map[string]Order   { return u.orders }
func (u *InMemoryUoW) Inventory() map[string]int   { return u.inventory }

type inMemoryTx struct {
	orders    map[string]Order
	inventory map[string]int
}

func (tx *inMemoryTx) SaveOrder(_ context.Context, order Order) error {
	tx.orders[order.ID] = order
	return nil
}

func (tx *inMemoryTx) ReserveInventory(_ context.Context, productID string, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	tx.inventory[productID] += qty
	return nil
}
