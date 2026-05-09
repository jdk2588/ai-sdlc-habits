# Bootstrap - Go Codebase Generation Prompt

**Run this once on your existing Go codebase to generate `01-base-instructions` and `02-file-patterns`.**

Paste the prompt below into a Claude Code session with your project open.
Review the output, correct anything wrong, then paste into the actual files.

---

## The Prompt

```
Analyze this Go codebase and produce two priming documents that will be loaded
into every future Claude session. These replace generic training-data defaults
with this project's specific patterns.

---
DOCUMENT 1: context/01-base-instructions.md

Produce a concise reference covering:

1. Module and stack
   - Go module path (from go.mod)
   - Go version
   - HTTP router/framework (chi, gin, echo, stdlib, etc.)
   - Database driver (pgx, database/sql, sqlx, etc.)
   - Queue or background job system (asynq, machinery, etc.) if any
   - Config loading approach (envconfig, viper, godotenv, etc.)
   - Logging library (slog, zap, logrus, etc.)
   - Testing libraries (testify, gomock, etc.)
   - Migration tool if any

2. Go idioms in use — which of these patterns this codebase applies:
   - Error wrapping style (fmt.Errorf + %w, pkg/errors, custom types)
   - Context propagation (is ctx first param everywhere?)
   - Interface placement (consumer-side vs implementation-side)
   - Constructor patterns (NewXxx functions vs struct literals)
   - Dependency injection approach (constructor injection, wire, fx, etc.)
   - Where does application setup/wiring happen?

3. Architecture rules
   - Package layout and what belongs where
   - Repository pattern or direct DB access?
   - Where does validation happen?
   - How are domain errors modelled?
   - Any middleware patterns in use?

4. Non-negotiables — correctness rules that can never be skipped
   (signature verification, idempotency, auth checks, context deadlines, etc.)

5. What NOT to do — the 5–8 most common ways code gets generated that
   contradicts this codebase (wrong error style, wrong package layout,
   global state, missing context, etc.)

Format: concise markdown, headers, under 400 words. No code examples.

---
DOCUMENT 2: context/02-file-patterns.md

Produce a concise reference covering:

1. Package and directory layout — full tree with one-line description per directory

2. Naming conventions table:
   - File naming by type (handler, service, repository, worker, domain type)
   - Function/method naming (constructors, handlers, service methods, repo methods)
   - Interface naming convention
   - Error type naming

3. Code patterns — one short code snippet (8–12 lines) for each:
   - An HTTP handler function
   - A service method
   - A repository method (with pgx or your DB driver)
   - A domain error type
   - A table-driven test skeleton

Format: concise markdown, under 600 words. Snippets only — no prose explanation.

---
After generating both documents:
- List 3–5 things you couldn't confidently infer (deployment target, SLA, team size,
  external dependencies)
- Note any conventions that seem inconsistent across the codebase
```

---

## After Running This

1. Review both documents — correct anything Claude inferred wrong
2. Paste into `context/01-base-instructions.md` and `context/02-file-patterns.md`
3. Merge key rules into `CLAUDE.md` (Stack, Go Idioms, What NOT to Do sections)
4. Fill in manually the items Claude couldn't infer

## When to Re-Run

- Major Go version upgrade or router/framework change
- Significant architectural refactor (e.g. switching from direct DB to repository pattern)
- Onboarding this system into an evolved codebase

