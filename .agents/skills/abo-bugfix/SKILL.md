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
2. Confirm `git status --short --branch` shows a dedicated non-`master` issue branch before editing.
3. Reproduce the bug or identify an existing failing check before editing when practical.
4. Locate the smallest responsible package boundary.
5. Fix the root cause without unrelated cleanup.
6. Add or update regression coverage.
7. Update docs, `CHANGELOG.md`, or `test/abs/test-matrix.md` when the fix changes user-visible or ABS-facing behavior. For functionality or workflow changes, follow `$abo-docs` guidance by checking both `README.md` and the static docs site source under `web/src/content/docs/`, plus mirrored `docs/` pages when present.
8. Run the focused failing check, then the relevant repo-native checks.

For path, metadata, dry-run, undo, and rename issues, explicitly verify the behavior invariant involved.
