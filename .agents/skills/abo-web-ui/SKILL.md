---
name: abo-web-ui
description: Build, debug, and verify the current Audiobook Organizer local browser UI in web/, cmd/web.go, cmd/gui.go, internal/server, and internal/app.
metadata:
  short-description: Work on new ABO web UI
---

# ABO Web UI

You are the Audiobook Organizer local browser UI engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/web-ui.md`.

## Workflow

1. Confirm the checkout contains the current `web/` frontend. If not, report that this branch does not contain the new web UI design and stop before web UI edits.
2. Confirm `git status --short --branch` shows a dedicated non-`master` issue branch before web UI edits.
3. Identify whether the change belongs in `web/`, `internal/server`, `internal/app`, `cmd/web.go`, or `cmd/gui.go`.
4. Preserve loopback-only server behavior and session token checks.
5. Preserve dry-run, undo, log, and path-safety invariants in UI-triggered workflows.
6. Keep API request/response changes synchronized across backend handlers, app services, frontend callers, and tests.
7. Verify backend changes with focused Go tests and frontend changes with `make web-build`.
8. For user interaction changes, run browser or Playwright checks when dependencies are available.

Keep this skill focused on the current local browser UI paths only.
