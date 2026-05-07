# CLAUDE.md — AI-Enabled SDLC Tutorial

This file briefs Claude Code on the conventions and non-negotiables for this project.
Read this before touching any code.

## Stack

- **Language:** Go 1.25
- **HTTP:** standard library (`net/http`) — no third-party router
- **Message broker:** RabbitMQ 3.13 via `github.com/rabbitmq/amqp091-go v1.11.0`
- **Storage:** in-memory maps with `sync.Mutex` — no database

## Project structure

```
services/
  orders/       — POST /orders, GET /orders/{id}, publishes order.placed
  fulfillment/  — consumes order.placed, publishes order.fulfilled
context/        — shared reference documents (base instructions, design docs, etc.)
CLAUDE.md       — this file
TUTORIAL.md     — narrative walkthrough of the five habits
```

Each service is a standalone Go module (its own `go.mod`). The two modules do not
import each other. All coordination is through RabbitMQ.

## Non-negotiables

### JSON field naming

All JSON tags must be **snake_case, full words, no abbreviations**.

```go
// correct
type Order struct {
    ID          string    `json:"order_id"`
    Quantity    int       `json:"quantity"`
    OrderStatus string    `json:"order_status"`
    CreatedAt   time.Time `json:"created_at"`
}

// wrong — never do this
type Order struct {
    ID  string `json:"orderId"`   // camelCase
    Qty int    `json:"qty"`       // abbreviation
    TS  string `json:"CreatedAt"` // PascalCase
}
```

This rule applies to every struct that has JSON tags: request bodies, response
bodies, and message payloads.

### Error handling

Wrap errors at every layer using `fmt.Errorf("%w", err)`. Use `errors.New` only
for new leaf errors that have no underlying cause.

```go
// correct
if err := db.Query(...); err != nil {
    return fmt.Errorf("query orders: %w", err)
}

// wrong — never discard the cause
return errors.New("query failed")
```

Never swallow errors silently. If an error is non-fatal (e.g., a failed publish),
log it with `log.Printf` and continue — do not ignore it entirely.

### Error response shape

Every HTTP error response must use exactly this shape:

```json
{"error": "description of what went wrong"}
```

Never use `{"message": "..."}`, `{"err": "..."}`, or any other key. Use the
`writeError(w, code, msg)` helper that already exists in the orders service.

### Logging

Use `log.Printf` / `log.Println` / `log.Fatalf` everywhere. Never use
`fmt.Printf` or `fmt.Println` for operational messages.

### No global state beyond the store

The only allowed package-level variables are:
- The in-memory store map and its mutex
- The `nextID` counter (orders service)
- The `records` map and its mutex (fulfillment service)

No other global variables. All dependencies (publishers, etc.) must be passed
as function parameters or captured in closures.

### No panic in handlers

HTTP handlers must never call `panic`. Return an error response instead.

### Error recovery paths

When a cleanup or error-recovery call (e.g., transitioning a record to `failed`,
rolling back state) returns an error, capture it and return it wrapped with
context. Do not drop errors in recovery branches.

```go
// correct
if ferr := transitionTo(r, StatusFailed); ferr != nil {
    return fmt.Errorf("set failed status: %w", ferr)
}

// wrong — the recovery itself failed silently
transitionTo(r, StatusFailed)
```

This applies equally to the primary path and to any cleanup or compensation
call inside an error branch.

## Message schemas

These schemas are the wire contract between the two services. Both services must
agree on field names and types.

### order.placed (published by orders, consumed by fulfillment)

```json
{
  "order_id": "1",
  "item": "widget",
  "quantity": 2
}
```

### order.fulfilled (published by fulfillment)

```json
{
  "order_id": "1",
  "status": "fulfilled"
}
```

## Claude Code instructions

- Read `context/01-base-instructions.md` before writing any Go code in this repo.
- When adding a new field to any struct that has JSON tags, follow the snake_case
  convention above and check both services to keep them consistent.
- When writing error handling, check that the error is wrapped with `%w` if there
  is an underlying cause.
- When writing HTTP handlers, check that all error paths use `writeError`.
- Run `go test .` from the service directory before declaring a task complete.
- Run `go build .` from the service directory to catch compile errors.
