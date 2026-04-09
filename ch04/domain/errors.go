package domain

// Domain errors are defined in money.go and order.go alongside
// the types they relate to. This file exists only as a reference
// for readers looking for the full error catalogue.
//
// Money errors (money.go):
//   ErrNegativeAmount   — amount cannot be negative
//   ErrCurrencyMismatch — currency mismatch
//   ErrInvalidDiscount  — discount must be between 0 and 100
//
// Order errors (order.go):
//   ErrOrderEmpty       — order must have items
//   ErrOrderNotPending  — operation requires pending status
//   ErrOrderShipped     — cannot modify a shipped order
//   ErrOrderCanceled    — cannot modify a canceled order
//   ErrOrderNotFound    — order not found
