---
context: AI-Enabled SDLC Tutorial
---

# Glossary

## Tutorial
A hands-on learning artifact structured as a GitHub repository. The commit history is the narrative — each phase is a discrete set of commits showing the codebase evolve as SDLC habits are applied. A `TUTORIAL.md` at the root guides learners through the phases.

## Harness Engineering
The broader practice of disciplined, habit-driven AI-assisted development. Not tied to a specific tool. The tutorial demonstrates harness engineering using Claude Code as the concrete reference implementation.

## Phase 0 (Chaos Baseline)
The starting state of the tutorial repo. Two developers have each used AI with zero shared context to build features across both services. The divergent output — inconsistent error handling, naming, package layout, struct conventions — is committed to the repo as the baseline that every subsequent habit phase improves upon.

## Habit Phase
One of the five habit changes from the One2N AI-Enabled SDLC talk, each represented as a phase in the tutorial commit history. Phases are numbered 1–5 and produce concrete artifacts in the `context/` directory.

## Developer 1 / Developer 2
The two personas in the tutorial collaboration scenario. Both developers touch both services (`orders` and `fulfillment`). The "stepping on each other's toes" friction — conflicting patterns in shared files — is the pain the habit phases resolve.

## Orders Service
The `orders` service in the sample system. Exposes two HTTP endpoints:
- `POST /orders` — accepts an order (item, quantity, price), validates it, persists to an in-memory store, publishes an `order.placed` event to RabbitMQ.
- `GET /orders/{id}` — returns the current status of an order.

## Fulfillment Service
The `fulfillment` service in the sample system. Consumes `order.placed` events from RabbitMQ, updates order status through a state machine (placed → processing → fulfilled), and publishes an `order.fulfilled` event back to RabbitMQ.

## order.placed
The RabbitMQ message published by the `orders` service when an order is accepted and persisted. Consumed by the `fulfillment` service.

## order.fulfilled
The RabbitMQ message published by the `fulfillment` service when an order has been processed. Represents the terminal happy-path state.

## Context Artifacts
The files produced in the `context/` directory during each habit phase. Numbered to match habit order and sort correctly in the repo.

| Phase | Habit | Artifacts |
|---|---|---|
| 0 | Chaos | Divergent Go code, no shared conventions |
| 1 | Brief your AI | `CLAUDE.md`, `context/01-base-instructions.md` |
| 2 | Design before you code | `context/02-design.md` |
| 3 | Write your team's rules | `context/03-generation-instructions.md`, `context/03-review-checklist.md` |
| 4 | Keep a decision log | `context/04-implementation-guide.md`, `context/04-execution-report.md` |
| 5 | Capture what you learn | `context/05-retro.md`, updated `CLAUDE.md`, `context/05-prompt-templates.md` |

# Decisions

## Target audience: experienced engineers, not beginners
The tutorial assumes Go and distributed systems familiarity. It skips "what is RabbitMQ" explanations and focuses on the SDLC habits. The target reader is someone already using AI to code who is frustrated by inconsistent results.

## Claude Code as the concrete reference tool
The tutorial is Claude Code specific — it uses `CLAUDE.md`, `.claude/` conventions, and `context/` directory patterns from the Claude Code harness. This makes artifacts copy-pasteable. Tool-agnostic framing is deferred to avoid complexity.

## Both developers touch both services
Developer 1 and Developer 2 both contribute to `orders` and `fulfillment`. This is intentional — the stepping-on-toes friction (conflicting patterns in shared code) is the demonstration vehicle for why team rules matter.

## TUTORIAL.md narrative voice: descriptive, not prescriptive
TUTORIAL.md tells the story of Developer 1 and Developer 2 — what they built, what diverged, and how each habit changed their output. The primary use case is a conference attendee studying the reference repo after the talk. It is not a step-by-step lab exercise.

## Two tutorial documents: post-talk study and live audience
The repo has two tutorial files serving distinct primary readers. `TUTORIAL.md` is the post-talk narrative for someone who has heard the talk and is studying the repo at their own pace. A second file (`TALK.md`) targets a live projector follow-along: the file is displayed on screen while the speaker talks. Format constraints: sections must fit one projector scroll, large visual landmarks, minimal prose, one habit per section. Audience reads from their seats at projector resolution — not on a personal device.

## Minimal infrastructure: in-memory store, no auth, no DLQ
The `orders` service uses an in-memory store. No separate database, no authentication on the HTTP API, no dead-letter queue for fulfillment failures. Infrastructure complexity is kept low so the habits are the story.

## Tech stack: Go + RabbitMQ + Docker Compose
Go for both services. RabbitMQ for async messaging between services. Docker Compose to run the full system locally. The async boundary between services is where design discipline pays off most visibly.
