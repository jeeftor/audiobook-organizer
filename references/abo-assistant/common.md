# Audiobook Organizer Common Reference

Use this shared reference for Audiobook Organizer repo-local skills.

## Repository Rules

- Read `AGENTS.md` first.
- Default branch is `master`.
- Non-trivial code or documentation work needs a GitHub issue before edits.
- Work on a dedicated branch from `origin/master` before editing; do not push directly to `master`.
- Use branch prefixes by work type: `feature/<short-name>`, `fix/<short-name>`, `docs/<short-name>`, or `chore/<short-name>`.
- Verify `git status --short --branch` before editing, committing, or pushing.
- If a separate Git worktree is created and hook config exists, run `prek install --hook-type pre-commit --hook-type commit-msg` inside that worktree.
- Preserve unrelated dirty worktree changes.
- Keep edits focused on the issue.
- Keep the issue updated with decisions, blockers, test results, and follow-up work.
- User-visible features, fixes, behavior changes, Docker/runtime changes, and documentation changes need a `CHANGELOG.md` entry under `Unreleased`.
- ABS-facing behavior changes must update `test/abs/test-matrix.md` unless explicitly not applicable.
- PRs target `master`; prefer Squash and merge; issues normally close through PR merge.
- Do not treat a local commit or draft PR as done. Finish the cycle by getting the PR ready, passing required checks, merging back into `master`, confirming the linked issue closed, and cleaning up the branch or worktree.

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
- `prek install --hook-type pre-commit --hook-type commit-msg`

Prefer `rg` for content search and `fd` or `find` for file discovery.
