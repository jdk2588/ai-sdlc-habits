# AI-Enabled SDLC — Talk Companion

Two Go services. RabbitMQ in between. Two developers. No shared conventions.

```
git checkout phase-0
```

---

## Phase 0: The Chaos

Both developers built the same system. Here is what diverged.

**JSON naming**
```go
// Developer 1
type Order struct { ID string `json:"orderId"` Qty int `json:"qty"` }

// Developer 2
type Order struct { ID string `json:"order_id"` Quantity int `json:"quantity"` }
```
The consumer silently deserialized empty values. Orders processed with zero quantities.

**Error response shape**
```
Developer 1 → {"message": "not found"}
Developer 2 → {"error": "not found"}
```

**Logging**
```
Developer 1 → fmt.Printf("processing order %s", id)   // no timestamp
Developer 2 → log.Printf("fulfillment: processing %s", id)  // timestamp
```

These are not careless mistakes. Both developers were experienced. The AI amplified each developer's individual style instead of converging on a shared one.

---

## The Five Habits

| # | Habit | Format |
|---|-------|--------|
| 1 | Brief your AI | **demo** |
| 2 | Design before you code | explain |
| 3 | Write your team's rules | **demo** |
| 4 | Keep a decision log | explain |
| 5 | Capture what you learn | explain |

---

## Habit 1: Brief Your AI

```
git checkout phase-1
```

**Problem:** The AI has no memory of yesterday's session and no knowledge of your project's specific conventions. Every session starts from zero.

**What we're about to demo:** Start a Claude Code session in this repo — with `CLAUDE.md` in place — and ask it to add a new struct. Then compare the output to what Phase 0 produced without it.

**Watch for:**
- Does the AI use snake_case JSON tags without being told?
- Does it use `log.Printf` instead of `fmt.Printf`?
- Does it use `{"error": "..."}` for error responses?

---
*— switch to terminal —*

---

**What was produced:**

| File | What it does |
|------|-------------|
| `CLAUDE.md` | Stack, non-negotiables, message schemas — read automatically at session start |
| `context/01-base-instructions.md` | Naming conventions, code patterns, file layout — the expanded brief |

The AI now starts every session with the same ground truth both developers agreed on.

---

## Habit 2: Design Before You Code

```
git checkout phase-2
```

**Problem:** Without a design document, both developers implement what seems reasonable to them. Wire schemas diverge. State machine transitions get defined differently in code than in conversation.

**Fix:** Write the design first. Treat it as a gate — implementation does not start until the document is reviewed and approved.

The gate is the habit. Writing the design is easy. Waiting for approval is the discipline.

**Produced:**

| File | What it does |
|------|-------------|
| `context/02-design.md` | Exact HTTP interface, AMQP message schemas, state machine transitions, error paths — all agreed before a line of code was written |

---

## Habit 3: Write Your Team's Rules

```
git checkout phase-3
```

**Problem:** Two developers, same AI tool, different output quality. The senior engineer's judgment — what to generate, what to block, what to flag — lives in their head, not in the AI's context.

**What we're about to demo:** Run the review checklist against the Phase 2 code. Watch the AI find two violations it introduced itself.

**Watch for:**
- A dropped return value in a recovery path (`transitionTo` error silently lost)
- A leaf error constructed with `fmt.Errorf` instead of `errors.New`

---
*— switch to terminal —*

---

**What was produced:**

| File | What it does |
|------|-------------|
| `context/03-generation-instructions.md` | How to build new code: handler pattern, struct rules, error construction table |
| `context/03-review-checklist.md` | BLOCK and FLAG items — violations with examples of wrong and correct forms |

Both AIs now run reviews against the checklist, not against each developer's intuition.

---

## Habit 4: Keep a Decision Log

```
git checkout phase-4
```

**Problem:** AI memory is ephemeral. Every session starts with re-explaining context. Decisions made last week are invisible to today's session.

**Fix:** Write an implementation guide *before* the code and an execution report *after*. The guide specifies the exact change — function signatures, state flows, test spec, out-of-scope list. The report records what was actually built and where it deviated.

The out-of-scope list is not decoration. Without it, the AI reasonably adds retries, dead-letter queues, and alerting. The scope boundary prevents each of those from appearing.

**Produced:**

| File | What it does |
|------|-------------|
| `context/04-implementation-guide.md` | What to build and why — written before implementation |
| `context/04-execution-report.md` | What was built, what changed, how to run it — written immediately after |

---

## Habit 5: Capture What You Learn

```
git checkout phase-5
```

**Problem:** The AI makes the same mistake in session after session. No shared knowledge base accumulates from the mistakes.

**Fix:** After each story, three questions:

1. **What did the AI consistently get wrong?** → Add a guardrail to `CLAUDE.md`
2. **What prompt worked really well?** → Save it as a template
3. **What was missing from the briefing?** → Add it to `context/01-base-instructions.md`

Guardrails belong in a PR, not a chat message. When a session goes well, note what prompt worked.

**Produced:**

| File | What it does |
|------|-------------|
| `context/05-retro.md` | Three-question retro from Phase 4 — what went wrong, what worked, what was missing |
| `context/05-prompt-templates.md` | Reusable prompts extracted from sessions that worked |
| `CLAUDE.md` | Updated with new guardrail: error capture in recovery paths |

---

## The Contrast

```
git checkout phase-0   # look at services/orders/handlers.go
git checkout phase-5   # look at the same file
```

At `phase-0`: any new code reproduces whoever wrote it last. No shared ground truth.

At `phase-5`: any new code starts from the same briefing. The habits compound — each retro makes the briefing more precise, a more precise briefing means fewer violations, fewer violations means a shorter feedback loop.

**The Phase 5 codebase is not the goal. The process that produced it is.**
