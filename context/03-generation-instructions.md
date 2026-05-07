# Generation Instructions — Order Processing Services

This document tells an AI coding assistant how to generate new code and refactor
existing code in this repo. Read `CLAUDE.md` and `context/01-base-instructions.md`
first — this document supplements them with workflow-level guidance.

## Before writing any code

Identify which service owns the change:

- `services/orders/` — accepts and validates order creation, stores orders in
  memory, publishes `order.placed`.
- `services/fulfillment/` — consumes `order.placed`, runs the state machine,
  publishes `order.fulfilled`.

Neither service queries the other's store. The only coupling is through RabbitMQ.
If a change touches both services, implement both sides in the same session so
the wire contract stays consistent.

## Building new code

### Handler pattern

All handlers that require dependencies use the closure form:

```go
func handleDoThing(dep dependency) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. decode / validate input
        // 2. call store
        // 3. call publisher (non-fatal if it fails — log and continue)
        // 4. write response
    }
}
```

Return early on error (guard clauses), success at the bottom. Never nest happy
path logic inside an `if err == nil` block.

### Adding a new struct

1. Name Go fields in PascalCase, no abbreviations (`Quantity` not `Qty`).
2. Add `json:"snake_case_full_word"` to every field that crosses a wire.
3. Add `json:"-"` to fields intentionally excluded from serialization.
4. Cross-check the JSON field names against the wire contract table in CLAUDE.md.

### Adding a new HTTP endpoint

1. Register the route in `main.go`.
2. Write the handler in `handlers.go`.
3. Write input validation in `validate.go` if there is a request body.
4. Write tests in `handlers_test.go`: valid input (expect 2xx), invalid input
   (expect 400), not found (expect 404 if applicable).

### Adding a new AMQP message type

1. Define a typed struct for the message body with snake_case JSON tags.
2. Use `json.Marshal` on the struct — never `fmt.Sprintf` to build JSON strings.
3. Add the schema to CLAUDE.md under "Message schemas" before publishing the code.
4. Update both services if the schema is shared between publisher and consumer.

### Error construction rules

| Situation | What to write |
|-----------|---------------|
| New error with no underlying cause | `errors.New("description")` |
| Wrapping a cause from a prior call | `fmt.Errorf("what you were doing: %w", err)` |
| Adding context without wrapping | Do not do this — always use `%w` when `err` exists |

Never construct a leaf error with `fmt.Errorf("message")` when `errors.New`
would suffice — `fmt.Errorf` without `%w` loses type information and is
confusing to readers.

## Refactoring existing code

### Safe refactors (proceed without asking)

- Renaming unexported identifiers within a package.
- Extracting a helper from a handler when the helper has no new side effects.
- Changing log message text.
- Changing test assertions or test setup.

### Risky refactors — confirm with the team first

| Change | Risk |
|--------|------|
| Renaming a JSON field | Breaks the wire contract; both services and CLAUDE.md must update atomically |
| Renaming an AMQP queue | Both services and the consumer tag must update together |
| Changing `OrderStatus` string constants | Every comparison across both services must update |
| Adding a package-level variable | May violate the global-state rule |

### Refactor checklist

Before declaring a refactor done:

- [ ] JSON tags unchanged unless wire contract change was intentional and documented
- [ ] `go test .` passes from every affected service directory
- [ ] No new package-level variables beyond the store and its mutex
- [ ] Error handling not weakened: no swallowed errors, no `%w` removed

## Always

- Wrap errors at every call site: `fmt.Errorf("what you were doing: %w", err)`.
- Use `writeError(w, code, msg)` for every HTTP error response.
- Use `json.Marshal` for all JSON construction — never `fmt.Sprintf`.
- Use `log.Printf` for all operational log messages.
- Use `errors.New` for leaf errors (no underlying cause).
- Declare string constants for status values used in more than one place.
- Hold the mutex for the minimum required scope — acquire, do work, release.
- Return errors from functions; do not panic.
- Log non-fatal errors with `log.Printf` before continuing.

## Never

- Package-level variables beyond the store map, its mutex, and the ID counter.
- `panic` in any HTTP handler or AMQP delivery handler.
- `fmt.Sprintf` to build JSON strings.
- A key other than `"error"` in HTTP error response bodies.
- Silent error drops — bare call to a function that returns an error with no
  check, no log, no return.
- Importing packages not already in `go.mod` without confirming first.
- Changing a JSON field name in a shared message struct without updating both
  services and CLAUDE.md in the same commit.
