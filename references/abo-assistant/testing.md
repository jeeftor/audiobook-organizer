# Audiobook Organizer Testing Reference

Use repo-native checks first. Start narrow, then widen before PR closeout when practical.

## Core Checks

- Format Go code: `make fmt`
- Check formatting: `make fmt-check`
- Vet and format check: `make lint`
- Unit tests: `make test-unit`
- Default tests: `make test`
- Integration tests: `make test-integration`
- All tests: `make test-all`
- Coverage: `make coverage`
- Direct package tests: `go test ./internal/organizer/...`
- Focused test: `go test -run TestName ./path/to/package`
- Pre-commit hooks, when configured: `prek run --all-files`
- New worktree hook setup, when configured: `prek install --hook-type pre-commit --hook-type commit-msg`

Prefer `prek` over `pre-commit`. If no `.pre-commit-config.yaml` or `.pre-commit-config.yml` exists on the branch, report that pre-commit hooks are not configured rather than forcing the command.

## Web UI Checks

The current web UI is the local browser UI:

- Backend entrypoints: `cmd/web.go` and `cmd/gui.go`.
- Backend services: `internal/server/` and `internal/app/`.
- Frontend root: `web/`.
- Install dependencies: `make web-install`.
- Build embedded assets: `make web-build`.
- Run frontend dev server: `make web-dev`.
- Full distribution build: `make build`.

If `web/` is absent on an older checkout, stop and report that the checkout does not contain the current new web UI design.

## ABS Checks

ABS-facing changes include Audiobookshelf discovery, path mapping, metadata mode, scan triggering, import/organize behavior, mounted-library behavior, and ABS-facing web/API flows.

- Update `test/abs/test-matrix.md` for new or changed ABS behavior.
- CI smoke: `make abs-ci-smoke`
- Implemented matrix: `make abs-test-matrix`
- Metadata mode: `make abs-test-metadata`
- Full E2E: `make abs-test-e2e`
- Focused ABS test: `go test -tags=abs_e2e ./test/abs/e2e -run TestName -count=1 -v`

ABS tests require Docker and can download public-domain fixtures. If Docker, network, or corporate proxy/certificates block the run, report the exact blocker and command attempted.

## Test Selection

- Organizer path/layout behavior: focused `internal/organizer` tests, then `make test-unit`.
- CLI flag/config behavior: focused `cmd` tests, then `make test-unit`.
- TUI behavior: focused `internal/tui` tests when present, then `make test-unit`.
- Web API behavior: focused `internal/app` or `internal/server` tests, plus `make web-build` when contracts or assets changed.
- Web frontend behavior: `make web-build`, plus browser/Playwright checks when interaction changes and dependencies are available.
- ABS behavior: focused ABS E2E, then `make abs-test-matrix` when practical.
- Docs-only skill changes: validate Markdown/frontmatter; Go tests are not required unless repo code changed.

## Reporting

Always record commands run, pass/fail status, the first actionable failure, and any checks not run with the reason.
