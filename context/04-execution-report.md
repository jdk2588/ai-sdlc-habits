# Phase 4 Execution Report — Fulfillment Failure Handling

Written after implementation.

## What was built

### processor.go — `fulfill()` signature change

`fulfill()` now accepts a `proc func(*FulfillmentRecord) error` parameter. The
processor runs between the `processing` and `fulfilled` state transitions. If the
processor returns an error, the record transitions to `failed` and is persisted
via `upsert()` before the error is returned to the caller.

```
fulfill(r *FulfillmentRecord, proc func(*FulfillmentRecord) error) error
```

### consumer.go — `handleDelivery()` updated

`handleDelivery()` now passes a no-op processor to `fulfill()`:

```go
fulfill(r, func(_ *FulfillmentRecord) error { return nil })
```

A real processor (inventory check, warehouse API call, etc.) can be substituted
without touching the state machine.

### processor_test.go — new test + signature updates

`TestFulfill_ProcessingFails` injects a processor that returns an error and
asserts that:
- `fulfill()` returns a non-nil error
- `r.Status` is `StatusFailed`

The two existing `fulfill()` tests (`TestFulfill_HappyPath`,
`TestFulfill_InvalidStartState`) were updated to pass the no-op processor.

A shared `noop` helper avoids repeating the inline function literal.

## What changed from the plan

No deviations. The implementation matched the guide exactly.

One observation not captured in the guide: the existing `fulfill()` had a
failure branch (`transitionTo(StatusFulfilled)` error) that was unreachable
because the transition is always valid from `processing`. The new design
removes that dead code by replacing it with a `proc()` call that is actually
injectable and testable.

## How to run it

**Trigger a failure manually:**

There is no live failure path yet (the no-op processor always succeeds). To
observe `failed` status in a test:

```go
r := &FulfillmentRecord{OrderID: "test-1", Status: StatusPlaced}
err := fulfill(r, func(_ *FulfillmentRecord) error {
    return errors.New("simulated failure")
})
// r.Status == StatusFailed, err != nil
```

**To observe failure via RabbitMQ:**

1. Replace the no-op in `handleDelivery` with a processor that returns an error
   for specific orders (e.g., based on item name).
2. `docker compose up`, then `make orders` and `make fulfillment`.
3. POST an order with the triggering item.
4. Check `records` in the fulfillment process — it will show `status: failed`.
5. A future GET /fulfillments/{id} endpoint would surface this over HTTP.

**All tests pass:**

```
ok  github.com/one2n/sdlc-tutorial/fulfillment   (9 tests)
ok  github.com/one2n/sdlc-tutorial/orders        (11 tests)
```

## What was learned

The pre-existing failed path in `fulfill()` was dead code that looked correct
but could never fire. Writing the implementation guide first forced the question
"how would processing actually fail?" — which surfaced the gap before touching
code. The injectable processor pattern is a direct answer: separate the state
machine logic from the work logic so failures can be injected in tests without
any mocking framework.

The guide noted orders-service observability as out of scope. That held. The
fulfillment record is updated correctly; surfacing it via HTTP or events is the
next natural step.
