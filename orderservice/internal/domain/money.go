package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeAmount   = errors.New("amount cannot be negative")
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrInvalidDiscount  = errors.New("discount must be between 0 and 100")
)

type Money struct {
	amount   int64
	currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add %s to %s: %w", m.currency, other.currency, ErrCurrencyMismatch)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Multiply(quantity int) (Money, error) {
	result := m.amount * int64(quantity)
	if result < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{amount: result, currency: m.currency}, nil
}

func (m Money) ApplyDiscount(percent int) (Money, error) {
	if percent < 0 || percent > 100 {
		return Money{}, ErrInvalidDiscount
	}
	discounted := m.amount * int64(100-percent) / 100
	return Money{amount: discounted, currency: m.currency}, nil
}

func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.amount, m.currency)
}
