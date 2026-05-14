# Audiobook Organizer Common Reference

Use this shared reference for Audiobook Organizer repo-local skills.

## Repository Rules

- Read `AGENTS.md` first.
- Default branch is `master`.
- Non-trivial code or documentation work needs a GitHub issue before edits.
- Work on a dedicated branch from `master`; do not push directly to `master`.
- Preserve unrelated dirty worktree changes.
- Keep edits focused on the issue.
- Keep the issue updated with decisions, blockers, test results, and follow-up work.
- User-visible features, fixes, behavior changes, Docker/runtime changes, and documentation changes need a `CHANGELOG.md` entry under `Unreleased`.
- ABS-facing behavior changes must update `test/abs/test-matrix.md` unless explicitly not applicable.
- PRs target `master`; issues normally close through PR merge.

## Architecture Map

- `main.go`: CLI entrypoint.
- `cmd/`: Cobra commands and top-level flag/config wiring.
- `internal/organizer/`: core organize and rename behavior.
- `internal/tui/`: Bubble Tea flows and screen models.
- `internal/app/`: app service layer used by the web API.
- `internal/server/`: local HTTP server, token checks, JSON routes, and embedded static assets.
- `web/`: current Vue/Vite local browser UI.
- `docs/`: user-facing docs.
- `testdata/`: audio and metadata fixtures.
- `test/abs/`: Audiobookshelf harness and E2E tests.

## Behavior Invariants

- Dry-run must not mutate the filesystem.
- Organization operations log to `.abook-org.log`.
- Rename operations log to `.abook-rename.log`.
- Undo compatibility must be preserved.
- Field mapping is first-class; prefer it over hard-coded metadata schema assumptions.
- `flat` mode implies embedded metadata and changes grouping behavior.

## High-Risk Areas

- `cmd/root.go`: flag aliases, Viper binding, and entrypoint behavior.
- `internal/organizer/path.go`: path generation and layout behavior.
- `internal/organizer/metadata_providers.go`: sidecar and embedded metadata behavior.
- `internal/organizer/types.go`: shared structs across CLI, TUI, web, and tests.
- `internal/tui/models/`: distributed Bubble Tea state.
- `internal/server/`: token checks and local API behavior.
- `internal/app/`: bridge from web requests into organizer, rename, and ABS operations.

## Useful Commands

- `gh issue list --state all --search "<query>"`
- `gh issue view <number> --comments`
- `gh issue create --title "<title>" --body-file <file>`
- `gh issue comment <number> --body-file <file>`
- `git fetch origin master`
- `git switch -c <branch> origin/master`
- `git status --short --branch`

Prefer `rg` for content search and `fd` or `find` for file discovery.
