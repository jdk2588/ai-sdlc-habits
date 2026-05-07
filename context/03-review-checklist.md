# Review Checklist — Order Processing Services

Use this checklist in every code review session. BLOCK items must be fixed
before merge. FLAG items are advisory — document the exception in the PR if you
choose not to fix.

---

## BLOCK — do not merge with these present

### B1: Unhandled errors

Every call that returns an error must do one of:
- Return it to the caller (with `fmt.Errorf("context: %w", err)` wrapping).
- Log it with `log.Printf` and continue (for genuinely non-fatal paths).
- Explicitly handle it (e.g., `errors.Is` branch).

**What to look for:**

```go
// BLOCK: error silently dropped
transitionTo(r, StatusFailed)

// BLOCK: error assigned to blank identifier
_, _ = someCall()

// OK: non-fatal error logged and execution continues
if err := pub.Publish(...); err != nil {
    log.Printf("service: failed to publish: %v", err)
}
```

### B2: Wrong error-wrapping style

Leaf errors (no underlying cause) must use `errors.New`. Errors that wrap a
cause must use `fmt.Errorf("context: %w", err)`.

**What to look for:**

```go
// BLOCK: fmt.Errorf used as a leaf error (no %w, no underlying cause)
return fmt.Errorf("quantity must be greater than zero")

// BLOCK: errors.New used when an underlying error exists
return errors.New("connection failed") // loses the original err

// OK
return errors.New("quantity must be greater than zero")
return fmt.Errorf("connect to RabbitMQ: %w", err)
```

### B3: Wrong HTTP error response shape

All error responses must use `writeError(w, code, msg)` and the key must be
`"error"`.

**What to look for:**

```go
// BLOCK: inline error response not using writeError
json.NewEncoder(w).Encode(map[string]string{"message": "not found"})

// BLOCK: wrong key
json.NewEncoder(w).Encode(map[string]string{"err": "..."})

// OK
writeError(w, http.StatusNotFound, "order not found")
```

### B4: Non-snake_case JSON tags

Every JSON tag must be `snake_case` and use full words, no abbreviations.

**What to look for:**

```go
// BLOCK: camelCase, PascalCase, or abbreviated tags
OrderID  string `json:"orderId"`
Qty      int    `json:"qty"`
ID       string `json:"OrderID"`

// OK
ID       string `json:"order_id"`
Quantity int    `json:"quantity"`
```

### B5: Illegal global state

The only allowed package-level variables are the store map and its mutex, the
ID counter (orders service), and the records map and its mutex (fulfillment
service).

**What to look for:**

```go
// BLOCK: new package-level variable
var defaultTimeout = 5 * time.Second
var cachedConfig config
```

### B6: panic in handlers

No HTTP handler or AMQP delivery handler may call `panic`.

**What to look for:**

```go
// BLOCK: any panic call in handler code
panic("unexpected state")
```

### B7: Missing input validation

Every POST or PUT handler that reads from `r.Body` must call a validation
function before using the decoded values.

**What to look for:**

- Handler decodes into a struct and immediately uses the fields without a
  `validateX(req)` call.
- Validation function exists but is not called on all fields.

---

## FLAG — advisory, document if skipped

### F1: Missing tests on public functions

Every exported function should have at least one test. Every handler should
have tests for: valid input, at least one invalid input, and not-found (if
applicable).

Flag when a new exported function has no test file entry.

### F2: Undocumented message schema changes

Any change to a JSON field name in a message struct (`PlacedMessage`,
`FulfilledMessage`, or similar) must be reflected in CLAUDE.md under
"Message schemas" in the same commit.

Flag when a message struct field name changed but CLAUDE.md was not updated.

### F3: fmt.Printf instead of log.Printf

Operational messages must use the `log` package.

Flag any `fmt.Printf` or `fmt.Println` calls outside of test files.

### F4: log.Printf style inconsistency

Log messages should follow the pattern `"service: description %s"` where
`service` is `"orders"` or `"fulfillment"`.

Flag messages that omit the service prefix or use a different separator.

---

## How to run a review session with Claude Code

1. Stage or open the diff for the change you want to review.

2. Start a new Claude Code conversation and paste this prompt:

   ```
   I am reviewing a change to the order processing services. Read the diff
   below, then check each BLOCK item from context/03-review-checklist.md.
   For each item, output: PASS, FAIL, or N/A. For each FAIL, show the exact
   line and explain what rule it breaks.

   <diff>
   [paste diff here]
   </diff>
   ```

3. For each FAIL, open a follow-up prompt:

   ```
   Fix the BLOCK failure for [item name]. Do not change any other behavior.
   Show only the changed lines.
   ```

4. Apply each fix, then run from the affected service directory:

   ```
   go test .
   go build .
   ```

5. After all BLOCK items pass, repeat steps 2-4 with the FLAG items.

Keep each review conversation focused on one change. Starting a new
conversation for each fix avoids the model losing earlier context.
