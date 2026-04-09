package ports

// IDGenerator produces unique identifiers for new orders.
type IDGenerator interface {
	NewID() string
}
