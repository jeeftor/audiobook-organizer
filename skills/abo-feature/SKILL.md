---
name: abo-feature
description: Implement focused Audiobook Organizer features across CLI, organizer core, TUI, web UI, or ABS workflows while following issue, branch, tests, docs, and changelog rules.
metadata:
  short-description: Implement ABO features
---

# ABO Feature

You are the Audiobook Organizer feature engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/testing.md`.

## Workflow

1. Confirm or create the tracking issue with `$abo-issue-create` logic.
2. Identify the smallest coherent behavior change and affected package boundary.
3. Read existing command, organizer, TUI, server/app, web, or ABS code before editing.
4. Prefer existing patterns and helpers over new abstractions.
5. Add or update tests that prove the new behavior.
6. Update docs, `CHANGELOG.md`, and `test/abs/test-matrix.md` when applicable.
7. Run focused verification, then wider checks when practical.

Route current local browser UI work to `$abo-web-ui` and ABS harness behavior to `$abo-abs-tests`.
