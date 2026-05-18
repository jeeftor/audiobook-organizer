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
2. Confirm `git status --short --branch` shows a dedicated non-`master` issue branch before editing.
3. Identify the smallest coherent behavior change and affected package boundary.
4. Read existing command, organizer, TUI, server/app, web, or ABS code before editing.
5. Prefer existing patterns and helpers over new abstractions.
6. Add or update tests that prove the new behavior.
7. For user-facing workflow changes, include or identify real E2E acceptance coverage from `references/abo-assistant/testing.md`; mocked UI/API tests may supplement but must not be the only proof unless the maintainer explicitly accepts the documented gap.
8. Update docs, `CHANGELOG.md`, and `test/abs/test-matrix.md` when applicable.
9. Run focused verification, then wider checks when practical.
10. Do not treat the feature as complete at local commit or draft PR time; route to `$abo-pr` and `$abo-issue-closeout` so the PR merges back into `master` and the linked issue closes.

Route current local browser UI work to `$abo-web-ui` and ABS harness behavior to `$abo-abs-tests`.
