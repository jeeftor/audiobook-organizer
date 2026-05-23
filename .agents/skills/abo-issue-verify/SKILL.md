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
2. Classify the issue as maintainer-created, user-originated, or unclear using `references/abo-assistant/common.md`.
3. Extract acceptance criteria and any follow-up decisions recorded during implementation.
4. Inspect the branch diff against `master`.
5. Verify each criterion against code, tests, docs, and behavior.
6. Check whether `CHANGELOG.md` is required and present.
7. For functionality or workflow changes, verify the docs impact check covered both `README.md` and the static docs site source under `web/src/content/docs/`, plus mirrored `docs/` pages when present.
8. Check whether `test/abs/test-matrix.md` is required and present for ABS-facing changes.
9. For user-facing workflow changes, verify that real E2E acceptance coverage exists and was run; classify mocked/stubbed UI or API tests as supplemental only.
10. For user-originated or unclear issues, determine whether repo-native verification fully proves the fix. If reporter confirmation or manual interaction is needed, mark closeout blocked until confirmation, maintainer approval, or documented obsolescence/duplication.
11. Run or confirm the narrowest relevant verification commands, widening when practical.
12. Verify whether the work is only locally complete or fully closed out through a PR merge back into `master`.
13. Report a checklist of pass/fail/blocked items and the exact remaining work.

Do not mark the issue done just because tests pass. Tie completion back to acceptance criteria and repo workflow obligations.
