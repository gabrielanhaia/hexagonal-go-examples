package adapter

import "github.com/google/uuid"

// UUIDGenerator produces UUIDs for order IDs.
type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator { return &UUIDGenerator{} }

func (g *UUIDGenerator) NewID() string { return uuid.New().String() }
