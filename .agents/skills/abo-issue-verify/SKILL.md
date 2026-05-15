---
name: abo-issue-verify
description: Verify that an Audiobook Organizer issue is done by checking acceptance criteria, code changes, tests, docs, changelog, and ABS matrix obligations.
metadata:
  short-description: Verify ABO issue completion
---

# ABO Issue Verify

You are the Audiobook Organizer acceptance verifier.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/testing.md`.

## Workflow

1. Identify the issue and read its body/comments.
2. Extract acceptance criteria and any follow-up decisions recorded during implementation.
3. Inspect the branch diff against `master`.
4. Verify each criterion against code, tests, docs, and behavior.
5. Check whether `CHANGELOG.md` is required and present.
6. Check whether `test/abs/test-matrix.md` is required and present for ABS-facing changes.
7. Run or confirm the narrowest relevant verification commands, widening when practical.
8. Verify whether the work is only locally complete or fully closed out through a PR merge back into `master`.
9. Report a checklist of pass/fail/blocked items and the exact remaining work.

Do not mark the issue done just because tests pass. Tie completion back to acceptance criteria and repo workflow obligations.
