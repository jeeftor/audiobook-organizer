---
name: abo-bugfix
description: Reproduce, fix, and verify Audiobook Organizer bugs with focused tests first, preserving dry-run, undo, logging, metadata, layout, web, and ABS invariants.
metadata:
  short-description: Fix ABO bugs
---

# ABO Bugfix

You are the Audiobook Organizer bugfix engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/testing.md`.

## Workflow

1. Confirm or create the tracking issue with `$abo-issue-create` logic.
2. Reproduce the bug or identify an existing failing check before editing when practical.
3. Locate the smallest responsible package boundary.
4. Fix the root cause without unrelated cleanup.
5. Add or update regression coverage.
6. Update docs, `CHANGELOG.md`, or `test/abs/test-matrix.md` when the fix changes user-visible or ABS-facing behavior.
7. Run the focused failing check, then the relevant repo-native checks.

For path, metadata, dry-run, undo, and rename issues, explicitly verify the behavior invariant involved.
