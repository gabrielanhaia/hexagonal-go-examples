# Hexagonal Go Examples

Companion code repository for **[Hexagonal Architecture in Go: Ports, Adapters, and Services That Last](https://xgabriel.com/hexagonal-go)** by Gabriel Anhaia.

## Structure

```
ch01/          Chapter 1:  The Spaghetti Service (the "before" example)
ch03/          Chapter 3:  Go interfaces as natural ports
ch04/          Chapter 4:  Domain — Money, Order, state machines
ch05/          Chapter 5:  Ports — small, consumer-defined interfaces
ch06/          Chapter 6:  Adapters — inbound (HTTP) and outbound (memory, console)
ch07/          Chapter 7:  Dependency rule — correct vs incorrect direction
ch15/          Chapter 15: Error handling across boundaries
ch16/          Chapter 16: Unit of Work — transactions without leaking SQL
ch17/          Chapter 17: Event-driven adapters — publisher, consumer
ch18/          Chapter 18: Observability — logging & metrics decorators
orderservice/  Chapters 8–14: The complete order service built throughout Part III
```

## Running

```bash
go test ./...
```

All examples target **Go 1.25+**.

## The Complete Order Service

The `orderservice/` directory contains the full hexagonal service built step by step in Part III:

```bash
cd orderservice
go run ./cmd/server/

# In another terminal:
curl -s -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":"cust-1","items":[{"product_id":"p1","quantity":2,"price_cents":2500,"currency":"USD"}]}'
```

## Book

Available at [xgabriel.com/hexagonal-go](https://xgabriel.com/hexagonal-go)

## License

MIT
