package domain

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCanceled  OrderStatus = "canceled"
)

var (
	ErrOrderEmpty      = errors.New("order must have items")
	ErrOrderNotPending = errors.New(
		"operation requires pending status",
	)
	ErrOrderShipped = errors.New(
		"cannot modify a shipped order",
	)
	ErrOrderCanceled = errors.New(
		"cannot modify a canceled order",
	)
	ErrOrderNotFound = errors.New("order not found")
)

type LineItem struct {
	ProductID string
	Quantity  int
	Price     Money
}

type Order struct {
	ID         string
	CustomerID string
	Items      []LineItem
	Status     OrderStatus
	CreatedAt  time.Time
}

func NewOrder(
	id string,
	customerID string,
	items []LineItem,
) (Order, error) {
	if len(items) == 0 {
		return Order{}, ErrOrderEmpty
	}
	return Order{
		ID:         id,
		CustomerID: customerID,
		Items:      items,
		Status:     OrderStatusPending,
		CreatedAt:  time.Now(),
	}, nil
}

func (o *Order) Total() (Money, error) {
	if len(o.Items) == 0 {
		return Money{}, ErrOrderEmpty
	}

	total, err := NewMoney(0, o.Items[0].Price.Currency())
	if err != nil {
		return Money{}, err
	}

	for _, item := range o.Items {
		lineTotal, err := item.Price.Multiply(item.Quantity)
		if err != nil {
			return Money{}, err
		}
		total, err = total.Add(lineTotal)
		if err != nil {
			return Money{}, err
		}
	}

	return total, nil
}

func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return ErrOrderNotPending
	}
	o.Status = OrderStatusConfirmed
	return nil
}

func (o *Order) Cancel() error {
	if o.Status == OrderStatusShipped {
		return ErrOrderShipped
	}
	if o.Status == OrderStatusCanceled {
		return ErrOrderCanceled
	}
	o.Status = OrderStatusCanceled
	return nil
}

func (o *Order) ApplyDiscount(percent int) error {
	if o.Status != OrderStatusPending {
		return ErrOrderNotPending
	}

	for i := range o.Items {
		discounted, err := o.Items[i].Price.ApplyDiscount(
			percent,
		)
		if err != nil {
			return err
		}
		o.Items[i].Price = discounted
	}
	return nil
}
