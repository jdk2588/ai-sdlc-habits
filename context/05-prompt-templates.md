# Prompt Templates — Order Processing Services

Reusable prompts for common AI-assisted development tasks on this repo.
Each template includes context for when to use it and why it worked.

---

## PT-1: Implementation guide prompt

**When to use:** Before implementing any change that introduces a new function
signature, modifies the state machine, or adds a new test case. Write this
guide first. Get it approved before touching code. This is the Phase 4 habit.

**Template:**

```
I need to implement [feature description] in [file or service].

What to build:
[1-2 sentences describing the feature, not the implementation approach]

Why:
[the observable problem this solves — what breaks or is invisible without it]

Out of scope:
- [explicit exclusion 1]
- [explicit exclusion 2]
- [add more as needed — the list should be non-obvious to be useful]

Interfaces involved:

[Function name] — [file]
Current signature: [current signature, or "new function"]
New signature: [exact new signature]

Data flow (happy path):
  [state/step 1] -> [state/step 2] -> [state/step 3]

Data flow (failure path, if applicable):
  [state/step 1] -> [state/step 2] -> [state/step 3 — error branch]

Tests required:
- [TestName]: injects [what], asserts [what]
- [TestName for existing test that needs updating]: [what changes in its setup]

Do not implement yet. Write only the guide. I will approve it before any
code is written.
```

**Why it worked (Phase 4 evidence):** The Phase 4 implementation guide
specified the exact `fulfill()` signature, both state flow paths as arrow
diagrams, and the assertion for `TestFulfill_ProcessingFails`. The result was
zero deviations between guide and implementation. The out-of-scope list
(retries, DLQ, alerting, failure reason storage) prevented each of those from
appearing in the generated code.

---

## PT-2: Checklist review prompt

**When to use:** After staging or completing a diff, before declaring a task
done. Paste the diff and ask Claude Code to check each BLOCK item from the
review checklist.

**Template:**

```
I am reviewing a change to the order processing services. Read the diff
below, then check each BLOCK item from context/03-review-checklist.md.
For each item, output: PASS, FAIL, or N/A. For each FAIL, show the exact
line and explain what rule it breaks.

<diff>
[paste git diff output here]
</diff>
```

Follow-up for each FAIL:

```
Fix the BLOCK failure for [item name]. Do not change any other behavior.
Show only the changed lines.
```

**Why it worked (Phase 3 evidence):** The structured PASS/FAIL/N/A output
caught two BLOCK failures that would have merged unnoticed: a silently dropped
error return in the recovery branch of `fulfill()` (B1) and a leaf validation
error built with `fmt.Errorf` instead of `errors.New` (B2). Referencing the
checklist file ensures AI applies the project's specific rules, not general Go
style guidance.
