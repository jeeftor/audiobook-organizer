# Audiobook Organizer Web UI Reference

Use this reference only for the current new local browser UI design.

## Current Design

- `cmd/web.go`: starts the loopback HTTP server, creates a session token, and opens the browser.
- `cmd/gui.go`: compatibility alias for the local browser UI.
- `internal/server/`: routing, request validation, API errors, static assets, and token checks.
- `internal/app/`: service layer that adapts web requests to organizer, rename, and Audiobookshelf services.
- `web/`: Vue/Vite frontend.

If a checkout does not contain `web/`, report that the current new web UI is absent from the branch before editing web UI code.

## Work Rules

- Preserve backend/frontend boundaries: `internal/app` owns workflow orchestration, `internal/server` owns HTTP, and `web/` owns presentation.
- Preserve loopback-only and token-check behavior.
- Do not bypass dry-run, undo, log, or path-safety invariants in web flows.
- Keep web API request/response changes reflected in both backend tests and frontend callers.
- For ABS-facing UI changes, update `test/abs/test-matrix.md` if behavior changes.
- Do not change global npm settings.

## Verification

- API-only changes: targeted Go tests for `internal/app` or `internal/server`, then `make test-unit` when practical.
- Frontend changes: `make web-build`.
- User interaction changes: start `make web-dev` and run browser/Playwright checks when dependencies are available.
- Embedded asset changes: `make web-build`, then `make build` when packaging behavior changed.
