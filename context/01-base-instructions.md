# Base Instructions — Order Processing Services

This document is the expanded briefing for AI coding assistants working on this
repo. It supplements `CLAUDE.md`. Read both before writing code.

## Stack details

### HTTP

Use only the Go standard library `net/http`. Do not add a third-party router such
as `gorilla/mux`, `chi`, or `gin`.

Handler pattern — always use the closure form so dependencies can be injected:

```go
func handleCreateOrder(pub publisher) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // implementation
    }
}
```

For handlers that need no dependencies, the plain function signature is fine:

```go
func handleGetOrder(w http.ResponseWriter, r *http.Request) { ... }
```

### RabbitMQ

Use `github.com/rabbitmq/amqp091-go v1.11.0`. Do not use the older
`github.com/streadway/amqp` package.

AMQP setup follows this sequence every time:
1. `amqp.Dial(url)` — connect
2. `conn.Channel()` — open a channel
3. `ch.QueueDeclare(...)` for each queue used — idempotent, safe to repeat
4. Pass the channel to publishers/consumers

Error handling during AMQP setup: wrap every error with `fmt.Errorf("step: %w", err)`.

### In-memory store

Orders: a `map[string]*Order` guarded by a `sync.Mutex`, plus a `nextID int`
counter for sequential IDs. All access to the map must hold the mutex.

Fulfillment: a `map[string]*FulfillmentRecord` guarded by a `sync.RWMutex`.
Reads use `RLock`, writes use `Lock`.

## Naming conventions

### Struct fields

Go field names are PascalCase. No abbreviations.

| Wrong | Correct |
|-------|---------|
| `Qty` | `Quantity` |
| `ID` for the JSON `order_id` | `ID` in Go, `json:"order_id"` tag |

### JSON tags

All JSON tags are snake_case, full words. Apply this to every struct that is
serialized or deserialized — request bodies, response bodies, and AMQP messages.

```go
type Order struct {
    ID          string    `json:"order_id"`
    Item        string    `json:"item"`
    Quantity    int       `json:"quantity"`
    Price       float64   `json:"price"`
    OrderStatus string    `json:"order_status"`
    CreatedAt   time.Time `json:"created_at"`
}
```

Fields that are intentionally excluded from JSON must use `json:"-"`, not omitted
tags.

### Function names

Follow standard Go conventions: exported functions are PascalCase, unexported are
camelCase. Handlers are named `handleX` (unexported) and registered in `main`.
Store helpers are named `saveX`, `findX`, `upsertX` (unexported).

### Status values

Order status values are plain strings: `"placed"`, `"processing"`, `"fulfilled"`,
`"failed"`. Define them as package-level string constants if used in multiple
places:

```go
const (
    StatusPlaced     = "placed"
    StatusProcessing = "processing"
    StatusFulfilled  = "fulfilled"
    StatusFailed     = "failed"
)
```

## Code patterns

### Error wrapping

Always wrap errors with context. Use `fmt.Errorf("what you were doing: %w", err)`.
Use `errors.New("message")` only when creating a new leaf error with no underlying
cause.

```go
// correct
conn, err := amqp.Dial(url)
if err != nil {
    return fmt.Errorf("connect to RabbitMQ: %w", err)
}

// correct
var errInvalidTransition = errors.New("invalid state transition")

// wrong
return errors.New("connection failed") // loses the original error
```

### HTTP error responses

Use the shared `writeError` helper for all error responses:

```go
func writeError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

Do not write error responses inline. Do not use a different key name. A 404 is
still `{"error": "order not found"}`, not `{"message": "..."}`.

### AMQP message serialization

Use `json.Marshal` to build message bodies. Never use `fmt.Sprintf` to construct
JSON strings manually — field names are easy to get wrong and hard to review.

```go
// correct
type placedMessage struct {
    OrderID  string `json:"order_id"`
    Item     string `json:"item"`
    Quantity int    `json:"quantity"`
}
body, err := json.Marshal(placedMessage{...})

// wrong
body := fmt.Sprintf(`{"orderId":"%s","qty":%d}`, id, qty)
```

### Logging

Use the `log` package throughout. The `log` package adds timestamps automatically,
which is useful in Docker Compose output. Never use `fmt.Printf` for operational
messages.

```go
log.Printf("orders: connected to RabbitMQ")
log.Printf("orders: failed to publish order.placed for %s: %v", orderID, err)
log.Fatalf("orders: server error: %v", err)
```

## Non-negotiables (never violate)

- Do not add global state beyond the store map and its mutex.
- Do not call `panic` in any HTTP handler.
- Do not swallow errors silently — at minimum log them with `log.Printf`.
- Do not use `fmt.Sprintf` to build JSON strings.
- Do not use a JSON key other than `"error"` in error responses.
- Do not import packages not already in `go.mod` without asking first.

## What each service owns

`services/orders/` is responsible for:
- Accepting and validating order creation requests
- Persisting orders in the in-memory store
- Publishing `order.placed` messages

`services/fulfillment/` is responsible for:
- Consuming `order.placed` messages
- Running the order through the fulfillment state machine
- Publishing `order.fulfilled` messages

Neither service queries the other's store. The only coupling is through RabbitMQ.
