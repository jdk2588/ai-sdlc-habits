# Phase 4 Implementation Guide — Fulfillment Failure Handling

Written before implementation. Approved before any code is changed.

## What to build

When the fulfillment service fails to process an order, the order record must
transition to `failed` status instead of remaining stuck in `processing`. The
failure must be reflected in the fulfillment service's in-memory store so it
is observable.

## Why

Without explicit failure handling, a processing error leaves an order in
`processing` indefinitely. Operators cannot distinguish "still in flight" from
"permanently broken". Transitioning to `failed` makes failures visible without
requiring DLQ infrastructure.

## Out of scope

- Retries and backoff logic
- Dead-letter queue (DLQ) configuration
- Alerting or metrics on failure rate
- Persisting failure reason or error message alongside the record
- The orders service subscribing to fulfillment events (future work)

## Interfaces involved

### `fulfill()` — processor.go

Currently the function drives the state machine directly
(`placed -> processing -> fulfilled`). The failing branch
(`transitionTo(StatusFulfilled)` error) is unreachable because the transition
is always valid.

**Change:** add an injectable `processor` function parameter. The processor
runs real fulfillment work between the `processing` and `fulfilled` transitions.
When the processor returns an error, `fulfill()` transitions to `failed` and
persists the failed record.

```
fulfill(r *FulfillmentRecord, proc func(*FulfillmentRecord) error) error
```

New happy-path flow:

```
placed -> processing (upsert) -> proc(r) -> fulfilled (upsert)
```

Failure flow when proc returns error:

```
placed -> processing (upsert) -> proc(r) -> failed (upsert) -> return err
```

### `handleDelivery()` — consumer.go

Must pass a `processor` to `fulfill()`. For Phase 4 the processor is a no-op
(`func(_ *FulfillmentRecord) error { return nil }`). A real processor (inventory
check, warehouse API call, etc.) can be substituted later without changing the
state machine.

### Tests — processor_test.go

New test: `TestFulfill_ProcessingFails`

Injects a processor that returns an error. Asserts that:
- `fulfill()` returns a non-nil error
- `r.Status` is `StatusFailed` after the call

Existing tests `TestFulfill_HappyPath` and `TestFulfill_InvalidStartState`
need their signatures updated to pass a no-op processor.

## What the failure signal looks like

After a failure, `records[orderID].Status == "failed"`. The record is persisted
via `upsert()` before `fulfill()` returns. An HTTP read endpoint (future work)
or the `order.fulfilled` message with `status: "failed"` (also future) would
surface this externally.
