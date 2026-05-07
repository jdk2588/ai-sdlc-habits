# Phase 5 Retro — Fulfillment Failure Handling

This retro covers the feature built in Phase 4: adding a `failed` status path
to the fulfillment state machine via an injectable processor function. The three
questions below reference concrete observations from that session.

---

## Q1: What did AI consistently get wrong?

**Error capture in recovery paths.**

Before Phase 4, the Phase 3 review found a B1 violation in the original
`fulfill()` function. The call to `transitionTo(r, StatusFailed)` returned an
error that was silently dropped:

```go
// pre-Phase-4 processor.go — B1 violation found by Phase 3 review
transitionTo(r, StatusFailed) // return value dropped
```

This is a specific pattern: AI handles errors on the primary path correctly but
drops errors from recovery or cleanup calls. The implicit assumption is that if
you are already in an error branch, a failure inside that branch does not need
to be propagated.

The Phase 4 implementation guide had to name this pattern explicitly and show
the required form:

```go
// correct — recovery error captured and propagated
if ferr := transitionTo(r, StatusFailed); ferr != nil {
    return fmt.Errorf("set failed status: %w", ferr)
}
```

The same underlying confusion appeared a second time in `validate.go` (B2 in
Phase 3 review): leaf validation errors used `fmt.Errorf` without `%w`, because
AI defaults to `fmt.Errorf` as the safe error constructor even when no
underlying error is being wrapped.

Both instances involved AI choosing the visually "safer" call without checking
whether the return value or wrapping form matched the project rules.

**Guardrail added to CLAUDE.md: see "Error recovery paths".**

---

## Q2: What prompt worked really well?

**The implementation guide format — function signature + state flow + test spec.**

The Phase 4 implementation guide specified:
1. Exact function signature: `fulfill(r *FulfillmentRecord, proc func(*FulfillmentRecord) error) error`
2. Arrow diagrams for both code paths (happy and failure)
3. Test name and assertions for the new test and the two tests that needed updating
4. An explicit out-of-scope list (retries, DLQ, alerting, failure reason storage)

The result: zero deviations between the guide and the implementation. No
corrections were needed. The execution report entry "What changed from the plan:
none" is the signal that the guide was precise enough.

The elements that mattered most:

- **Named state flows** removed ambiguity about which transitions fire in the
  failure path (`placed -> processing -> proc(r) -> failed -> return err`).
- **Test name + assertions** gave AI an unambiguous definition of "done" for the
  test file — no guessing about what to assert.
- **Out-of-scope list** prevented scope creep. Without it, retry logic or a
  DLQ exchange declaration would have appeared.

The reusable template is in `context/05-prompt-templates.md`.

---

## Q3: What was missing from the briefing?

**Function-parameter dependency injection was not documented.**

`context/01-base-instructions.md` documents handler injection via closures:

```go
func handleCreateOrder(pub publisher) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) { ... }
}
```

It does not cover injecting a dependency as a typed function parameter:

```go
func fulfill(r *FulfillmentRecord, proc func(*FulfillmentRecord) error) error
```

When Phase 4 required an injectable processor, the pattern had to be specified
from scratch in the implementation guide. Without that explicit spec, AI's
first approach would have been to hardcode the processing logic inside
`fulfill()` or introduce a package-level variable for the processor — both of
which violate the no-global-state rule.

**Addition to `context/01-base-instructions.md`: see "Dependency injection via function parameters".**
