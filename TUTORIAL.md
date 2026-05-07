# AI-Enabled SDLC: The Story of Two Developers and Five Habits

This is the story of Developer 1 and Developer 2. Both are experienced engineers.
Both use AI assistants daily. Both were given the same task: build an order
processing system in Go with RabbitMQ. Neither was given any shared conventions
to start from.

What they produced without shared conventions - and how five habits changed it -
is what this repository shows.

The five habits are from the One2N talk "AI Enabled SDLC: We have come far, we have a long way to go". The accompanying slide deck is in this repository: [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf).

The commit history is the artifact. Each phase is a discrete set of commits.
Check out a phase tag to see the codebase at that point in the story. Read
the files in `context/` to see the thinking that preceded the code.

---

## The system

The system is intentionally simple so the habits, not the architecture, are the
story.

Two services. `services/orders/` accepts order creation requests over HTTP,
persists them in memory, and publishes an `order.placed` event to RabbitMQ.
`services/fulfillment/` consumes those events, runs each order through a state
machine (`placed -> processing -> fulfilled`), and publishes `order.fulfilled`.
The two services share no code and do not query each other. The only coupling
is through RabbitMQ.

Both Developer 1 and Developer 2 touched both services. That is deliberate. The
friction of two developers producing divergent patterns in the same files is the
demonstration vehicle.

---

## Phase 0: Chaos

> Tag: `phase-0`
> Slides 3-7 in [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf): the usual AI workflow loop and the characteristics of an AI agent that produce these divergences.

Developer 1 started with the orders service. Developer 2 started with
fulfillment. When they each expanded into the other's service, they brought
their own patterns - and their AI brought those patterns too, because each AI
had only one developer's session history to learn from.

The divergence was not careless. Both developers wrote defensively. Both handled
errors. But they did it in incompatible ways, and the AI amplified each
developer's individual style rather than moving toward a shared one.

### JSON field naming

Developer 1's order struct used camelCase JSON tags - a habit from previous
JavaScript-adjacent work:

```go
type Order struct {
    ID       string  `json:"orderId"`
    Quantity int     `json:"qty"`
    Price    float64 `json:"unitPrice"`
}
```

Developer 2's struct used snake_case with full words:

```go
type Order struct {
    ID       string  `json:"order_id"`
    Quantity int     `json:"quantity"`
    Price    float64 `json:"price"`
}
```

When Developer 2 added the `order.placed` AMQP message consumer in the
fulfillment service, the field names in the message body did not match what the
orders service published. The consumer silently deserialized empty values. Orders
were processed, but with zero quantities.

### Error response shape

Developer 1 returned errors as `{"message": "description"}`. Developer 2 used
`{"error": "description"}`. A client written against one half of the API would
fail against the other half.

### Logging

Developer 1's code used `fmt.Printf` for operational messages:

```go
fmt.Printf("Processing order %s\n", orderID)
```

Developer 2's code used `log.Printf`:

```go
log.Printf("fulfillment: processing order %s", orderID)
```

In Docker Compose output, log entries from Developer 2's code had timestamps.
Developer 1's did not. Correlating events across services during debugging
required mentally compensating for the missing timestamps.

### Error handling in recovery paths

When the fulfillment state machine encountered a failure during processing, one
version of the code tried to transition the order to `failed` status - but
dropped the return value:

```go
// the return value is ignored - if this call fails, the failure is silent
transitionTo(r, StatusFailed)
return fmt.Errorf("processing failed: %w", err)
```

The state machine appeared to handle failures, but if the transition to `failed`
itself failed, the order would remain stuck in `processing` with no observable
signal. The error was in the recovery path, which is the branch that runs when
things are already going wrong - the worst place for a silent failure.

### The pattern

The individual violations were fixable. The structural problem was not: any new
code written by either developer's AI would reproduce that developer's patterns.
There was no shared ground truth for the AI to reference. The divergence was not
a one-time incident. It was a process.

---

## Habit 1: Brief Your AI

> Tag: `phase-1`
> Artifacts: `CLAUDE.md`, `context/01-base-instructions.md`
> Slides 9-11 in [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf): what to put in a briefing document and the gotchas around keeping it current.

The first habit is giving your AI a briefing document before it writes a single
line.

Developer 1 and Developer 2 agreed on the non-negotiables: snake_case JSON tags
with full words, `log.Printf` for all operational messages, `{"error": "..."}` as
the only error response shape, `fmt.Errorf("context: %w", err)` for wrapping
errors, and no global state beyond the store.

They wrote these into `CLAUDE.md` at the repository root. Claude Code reads
`CLAUDE.md` automatically at the start of every session. Any AI-assisted session
in this repo now starts with the same ground truth - not the ground truth from
one developer's history, but the ground truth the team agreed on together.

`CLAUDE.md` covers what to use (the stack, the patterns) but not why the rules
exist or how to apply them in less obvious situations. For that, they wrote
`context/01-base-instructions.md`.

`context/01-base-instructions.md` is longer and more granular. It shows the AMQP
setup sequence, the correct handler closure pattern, the distinction between
`errors.New` for leaf errors and `fmt.Errorf("%w")` for wrapping, and the
in-memory store access patterns. It also explains the naming rules in enough
detail that an AI generating a new struct has an unambiguous guide.

After Phase 1, both developers' AIs were reading the same document before
writing code. New structs had snake_case tags. Error responses used `{"error":
"..."}`. Log messages used `log.Printf` with a service prefix.

The Phase 0 divergence patterns did not disappear from the existing code
immediately - they required explicit cleanup. But no new code reproduced them.

---

## Habit 2: Design Before You Code

> Tag: `phase-2`
> Artifact: `context/02-design.md`
> Slides 12-13 in [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf): the five elements of a design document and the importance of explicit approval gates.

The second habit is writing the design before the implementation begins, and
treating that design as a gate.

Phase 2 added two features: `GET /orders/{id}` in the orders service, and the
fulfillment state machine. Both were needed to make the system observable and
to show meaningful state transitions. Both touched the boundary between the two
services - the `order.placed` and `order.fulfilled` wire schemas.

Without a design document, each developer would have implemented what seemed
reasonable to them. The wire schemas would have diverged again. The state machine
transitions might have been defined differently in code than in the prose
explanation. Integration would have required negotiation after the fact.

Developer 1 and Developer 2 wrote `context/02-design.md` first. The document
specifies the exact HTTP interface for `GET /orders/{id}` - the request shape,
the 200 response fields with types, the 404 response - and the exact AMQP
message schemas for `order.placed` and `order.fulfilled`. It defines the state
machine transitions explicitly:

```
placed -> processing    (start of processing)
processing -> fulfilled  (processing succeeded)
processing -> failed     (processing error)
```

It also names the error paths: what happens when RabbitMQ is unavailable at
startup, what happens when an `order.placed` message cannot be deserialized,
what happens when a state transition fails. The document ends with an execution
gate declaration: implementation did not begin until this document was reviewed
and approved.

The gate is the habit. Writing the design is easy. Waiting for approval before
starting implementation is the discipline. The gate forces the question "did we
agree?" before anyone's AI produces code that the other half of the team would
need to reverse-engineer.

After Phase 2, the `GET /orders/{id}` handler and the fulfillment state machine
matched the wire contracts in the design document. No integration negotiation
was needed.

---

## Habit 3: Write Your Team's Rules

> Tag: `phase-3`
> Artifacts: `context/03-generation-instructions.md`, `context/03-review-checklist.md`
> Slides 14-15 in [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf): generation instructions, review checklists, and why rules that try to cover everything enforce nothing.

The third habit is writing explicit generation and review rules that an AI can
apply mechanically.

`CLAUDE.md` and `context/01-base-instructions.md` tell an AI what the patterns
are. `context/03-generation-instructions.md` tells an AI how to build new code
and what to check before declaring a change done. It specifies the handler
pattern, the steps for adding a new struct, the steps for adding a new HTTP
endpoint, and a table of error construction rules:

| Situation | What to write |
|-----------|---------------|
| New error with no underlying cause | `errors.New("description")` |
| Wrapping a cause from a prior call | `fmt.Errorf("context: %w", err)` |

It also lists the refactors that require team confirmation before proceeding -
renaming a JSON field, renaming an AMQP queue, adding a package-level variable
- and explains the risk for each.

`context/03-review-checklist.md` is the counterpart: the document that catches
violations. It defines BLOCK items (violations that must be fixed before merge)
and FLAG items (advisory violations worth noting). Each BLOCK item includes an
example of the violation and the correct form.

When Developer 1 and Developer 2 ran a review of the Phase 2 code against the
checklist, two BLOCK violations surfaced.

The first was in the fulfillment processor. The call to transition a failed order
to `StatusFailed` dropped its return value:

```go
transitionTo(r, StatusFailed) // B1: return value silently dropped
return fmt.Errorf("processing failed: %w", err)
```

This is the same violation that appeared in Phase 0. It survived into Phase 2
because neither developer's AI had a rule for it. The checklist named it
explicitly, and the fix was clear:

```go
if ferr := transitionTo(r, StatusFailed); ferr != nil {
    return fmt.Errorf("set failed status: %w", ferr)
}
```

The second was in the validation layer. A leaf validation error used `fmt.Errorf`
without `%w`:

```go
return fmt.Errorf("quantity must be greater than zero") // B2: use errors.New
```

This loses type information and is misleading - `fmt.Errorf` signals to a reader
that an underlying error is being wrapped. The correct form is `errors.New`, which
signals that this is the origin of the error chain.

Both violations followed the same pattern: the AI chose the visually "safer"
call without checking whether it matched the project's specific rules. The
checklist made the project's rules machine-checkable.

After Phase 3, reviews ran against the checklist, not against each developer's
intuition. Both AIs could run a review by reading the checklist and applying it
to a diff.

---

## Habit 4: Keep a Decision Log

> Tag: `phase-4`
> Artifacts: `context/04-implementation-guide.md`, `context/04-execution-report.md`
> Slides 16-17 in [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf): implementation guides, execution reports, and documenting rejected alternatives alongside the happy path.

The fourth habit is writing an implementation guide before writing the code, and
an execution report after - creating a before/after record of every significant
change.

Phase 4 added failure handling to the fulfillment state machine. The existing
`fulfill()` function had a failure branch that was unreachable: the call to
`transitionTo(r, StatusFulfilled)` would always succeed from the `processing`
state, so the error branch after it could never fire. Failures during actual
fulfillment work - an inventory check, a warehouse API call, whatever a real
processor would do - had no place to go.

The implementation guide in `context/04-implementation-guide.md` specified the
exact change before any code was touched:

- The new `fulfill()` signature: `fulfill(r *FulfillmentRecord, proc func(*FulfillmentRecord) error) error`
- The happy-path state flow: `placed -> processing (upsert) -> proc(r) -> fulfilled (upsert)`
- The failure-path state flow: `placed -> processing (upsert) -> proc(r) -> failed (upsert) -> return err`
- The new test: `TestFulfill_ProcessingFails` - injects a failing processor, asserts `r.Status == StatusFailed`
- The existing tests that needed signature updates
- An explicit out-of-scope list: retries, dead-letter queue, alerting, persisting the failure reason

The out-of-scope list is not decoration. Without it, an AI generating code for
"failure handling" would reasonably add retry logic or a dead-letter queue
exchange declaration. The scope boundary prevented each of those from appearing.

The execution report in `context/04-execution-report.md` records what was
actually built and what changed from the plan. In this case: nothing changed.
The implementation matched the guide exactly. The report records this as a
signal - zero deviations means the guide was precise enough.

It also captures a new observation: the original `fulfill()` contained dead code
that looked correct but could never fire. Writing the implementation guide first
surfaced this before touching the code. The guide forced the question "how would
processing actually fail?" - which led directly to the injectable processor
pattern.

The injectable processor is what makes the failure path testable without a
mocking framework. The production code passes a no-op processor. A test passes a
failing processor. The state machine is unchanged in both cases.

---

## Habit 5: Capture What You Learn

> Tag: `phase-5`
> Artifacts: `context/05-retro.md`, `context/05-prompt-templates.md`, updated `CLAUDE.md`
> Slides 18-19 in [One2N-AI-SDLC.pdf](One2N-AI-SDLC.pdf): the three retro questions, and the gotchas around retro ownership and persisting guardrails in the repo.

The fifth habit is running a structured retrospective after each significant
change and feeding the findings back into the team's shared documents.

`context/05-retro.md` asks three questions about the Phase 4 implementation:

1. What did AI consistently get wrong?
2. What prompt worked really well?
3. What was missing from the briefing?

The answers are specific. AI got the error capture in recovery paths wrong twice
- once in Phase 0 (when the violation was introduced), once in Phase 3 (when
the review checklist caught it). The pattern is: AI handles errors on the
primary path correctly but drops errors from recovery or cleanup calls. The
implicit assumption is that when you are already in an error branch, a failure
inside that branch does not need to be propagated.

The retro named this pattern explicitly. The finding was added to `CLAUDE.md`
as a new "Error recovery paths" section with a correct and incorrect example.
Future AIs in this repo will read that rule before writing any error handling.
The pattern does not need to be caught by review again.

The retro also identified what was missing from the briefing: function-parameter
dependency injection. `context/01-base-instructions.md` documented the closure
form for handler injection, but not the pattern of passing a typed function
parameter to non-handler functions. Without that pattern documented, an AI
implementing the injectable processor would have defaulted to either hardcoding
the logic inside `fulfill()` or introducing a package-level variable - both
violations of the no-global-state rule. The pattern was added to
`context/01-base-instructions.md` after the retro.

`context/05-prompt-templates.md` extracts the two prompt formats that produced
reliable results: the implementation guide format (function signature + state
flows + test spec + out-of-scope list) and the checklist review format (diff +
PASS/FAIL/N/A output + per-FAIL follow-up). Both templates include the evidence
that they worked.

The retrospective does not need to be long. It needs to be honest about what the
AI got wrong, specific about what made the good prompt work, and concrete about
what is now missing from the shared documents.

---

## The contrast

Check out `phase-0` and look at the code. Then check out `phase-5` and look at
the same files.

The surface difference is visible: consistent JSON tags, consistent error
response shapes, consistent logging. But the structural difference is the one
that compounds.

At `phase-0`, any new code written by either developer's AI would reproduce that
developer's patterns. There was no shared ground truth. Adding a new struct, a
new endpoint, or a new message type was a coin flip on which conventions would
land in the diff.

At `phase-5`, any new code written by either developer's AI starts from the same
briefing. `CLAUDE.md` and `context/01-base-instructions.md` define the stack and
patterns. `context/03-generation-instructions.md` defines how to build new code.
`context/03-review-checklist.md` defines how to catch violations. The
implementation guide format in `context/05-prompt-templates.md` defines how to
specify a change before writing it. The retro format defines how to capture what
the AI got wrong and feed it back into the documents.

The habits compound. Each retro makes the briefing more precise. A more precise
briefing means fewer violations in the next review. Fewer violations in the next
review means a shorter feedback loop. Over time, the AI's first pass gets closer
to what the team would have written.

The Phase 5 codebase is not the goal. The process that produced it is.

---

## The files

```
CLAUDE.md                             - stack, non-negotiables, message schemas
context/01-base-instructions.md       - stack details, naming conventions, code patterns
context/02-design.md                  - GET /orders/{id}, state machine, wire schemas
context/03-generation-instructions.md - how to build new code; risky refactor list
context/03-review-checklist.md        - BLOCK and FLAG items for every review session
context/04-implementation-guide.md    - Phase 4 feature spec (before implementation)
context/04-execution-report.md        - Phase 4 results (after implementation)
context/05-retro.md                   - what AI got wrong, what worked, what was missing
context/05-prompt-templates.md        - reusable prompts extracted from Phase 4 and 5
```
