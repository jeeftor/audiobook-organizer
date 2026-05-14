# AGENTS.md

This file is the single source of truth for AI coding agents working in this repository. Claude Code should read this file through the `CLAUDE.md` compatibility pointer.

For task-specific workflow details, use the repo-local skills in `skills/abo-*`.

## Project Summary

Audiobook Organizer is a Go application for organizing and renaming audiobook libraries using metadata from `metadata.json` files and embedded metadata in EPUB, MP3, and M4B files.

The repository supports these user-facing entrypoints:

- `audiobook-organizer` for non-interactive CLI organization
- `audiobook-organizer tui` for interactive terminal organization
- `audiobook-organizer rename` for CLI renaming
- `audiobook-organizer rename-tui` for interactive terminal renaming
- `audiobook-organizer web` for the local browser UI
- `audiobook-organizer gui` as a compatibility alias for the local browser UI

The project ships one `audiobook-organizer` binary with CLI, TUI, rename, Audiobookshelf, and local web UI support.

## Repository Shape

- `main.go`: CLI entrypoint
- `cmd/`: Cobra commands and top-level flag/config wiring
- `internal/organizer/`: core organization and rename logic
- `internal/tui/`: Bubble Tea TUI flows and models
- `internal/app/`: application service layer used by the web API
- `internal/server/`: local HTTP server, token checks, JSON routes, and embedded static assets
- `web/`: Vue/Vite frontend for the local browser UI
- `docs/`: user-facing documentation
- `testdata/`: test fixtures for audio and metadata scenarios
- `internal/organizer/integration/`: integration tests
- `test/abs/`: Audiobookshelf test harness and E2E tests

The supported UI is the local browser UI through `audiobook-organizer web`, `audiobook-organizer gui`, `cmd/web.go`, `cmd/gui.go`, `internal/server/`, `internal/app/`, and `web/`.

## Repo-Local Skills

Use these skills for repeatable Audiobook Organizer workflows:

- `$abo-workflow`: route broad maintainer requests to the right specialist skill.
- `$abo-feature`: implement focused features across CLI, core, TUI, web, or ABS boundaries.
- `$abo-bugfix`: reproduce, fix, and verify regressions with focused tests.
- `$abo-issue-create`: create or reuse an issue and prepare the issue branch.
- `$abo-issue-watcher`: inspect issue status, comments, linked PRs, and next steps.
- `$abo-issue-verify`: verify acceptance criteria, tests, docs, changelog, and ABS matrix obligations.
- `$abo-issue-closeout`: finish issue hygiene and close only when appropriate.
- `$abo-tests`: select, write, and run repo-native Go, TUI, server/app, web, and docs checks.
- `$abo-abs-tests`: handle Audiobookshelf harness, ABS E2E, and `test/abs/test-matrix.md` work.
- `$abo-web-ui`: work only on the current local browser UI design in `web/`, `internal/server/`, `internal/app`, `cmd/web.go`, and `cmd/gui.go`.
- `$abo-audit`: audit Go and current web UI dependencies without changing files.
- `$abo-updater`: update Go and current web UI dependencies, then verify.
- `$abo-docs`: maintain docs, AGENTS.md, changelog, and repo-local skill references.
- `$abo-pr`: route PR drafting, creation, watching, and closeout.
- `$abo-pr-writer`: draft or update PR descriptions.
- `$abo-pr-create`: commit, push, and create a draft PR into `master`.
- `$abo-pr-watcher`: watch PR CI, review comments, issue comments, and branch freshness.

Shared skill references live in `references/abo-assistant/`. Keep AGENTS.md focused on durable repo rules; put detailed repeatable procedures in the relevant skill or shared reference.

## GitHub Workflow

- Track non-trivial code and documentation changes with a GitHub issue before editing files.
- If an issue already exists, use it. If not, create one with the goal, motivation, and acceptance criteria.
- Create a dedicated branch from `master` for each issue. Use a descriptive branch name such as `feature/download-search-results` or `fix/series-metadata`.
- Keep the issue updated while working. Add comments for scope changes, important implementation decisions, blockers, test results, and follow-up work discovered during implementation.
- Keep commits focused on the issue. Do not mix unrelated cleanup, refactors, or separate features into the same branch.
- As part of each feature or fix, decide whether tests, docs, and `CHANGELOG.md` need updates. If they do, include them in the same branch. If they do not, note why in the PR.
- For new or changed ABS-facing features, update `test/abs/test-matrix.md` before implementation is considered complete, then add or update the corresponding automated coverage in the ABS test matrix workflow.
- Maintain the root `CHANGELOG.md`. User-visible features, fixes, behavior changes, Docker/runtime changes, and documentation changes should add a concise changelog entry under `Unreleased` before the PR is merged.
- Before opening a PR, run the relevant repo-native checks. If a check cannot be run or has known unrelated failures, document that in the PR.
- When pre-commit hooks are configured, prefer `prek run --all-files` over `pre-commit run --all-files`. If hooks are installed locally but no config exists on the branch, report that instead of treating hook execution as required.
- Open a pull request into `master` when the branch is ready. The PR body must include the issue it resolves, a short summary, tests run, docs/changelog status, and any follow-up issues created.
- Issues should close through the PR merge, not through direct commits to `master`.
- After a PR is merged, delete the remote feature branch and remove the local branch or worktree.
- Do not push directly to `master` for normal feature or fix work.
- If work is paused or deferred, leave the issue open and comment with the current state and next step.

Tiny explicitly requested edits may proceed without creating an issue, but do not mix unrelated work.

## Architecture Notes

### Command Layer

`cmd/root.go` is the main orchestration point for CLI organization:

- Cobra defines flags and subcommands.
- Viper handles config, environment variables, and defaults.
- `--dir` and `--input` are interchangeable.
- `--out` and `--output` are interchangeable.
- `flat` mode automatically enables embedded metadata.

Additional command files include:

- `cmd/tui.go`: organization TUI
- `cmd/rename.go`: rename CLI and template-driven rename flow
- `cmd/rename_tui.go`: rename TUI
- `cmd/web.go`: local browser UI server
- `cmd/gui.go`: compatibility alias for the web UI
- `cmd/version.go`, `cmd/update.go`, `cmd/metadata.go`: auxiliary commands

### Core Organizer Logic

Core logic lives in `internal/organizer/`.

Important files:

- `organizer.go`: main organizer config and execution setup
- `organize.go`: move/copy organization flow
- `renamer.go`: file rename flow
- `metadata_providers.go`: metadata extraction from JSON and embedded sources
- `types.go`: shared types including `Metadata`, `FieldMapping`, logs, and summaries
- `path.go`: path construction and sanitization
- `logging.go`: undo log support
- `album_detection.go` and `album_handler.go`: multi-file audiobook grouping
- `template.go`: rename template support
- `author_formatter.go`: author formatting logic for renames

### TUI Structure

`internal/tui/` uses Bubble Tea. Most screen state lives under `internal/tui/models/`.

Notable flows:

- scan -> book list -> preview -> settings -> process for organization
- scan -> metadata/field-mapping/template preview -> process for rename

When changing TUI behavior, inspect both the screen model and any shared view/style helpers.

### Web UI Structure

The local browser UI uses a Go backend and Vue/Vite frontend:

- `cmd/web.go` starts the loopback HTTP server, creates a session token, and opens the browser.
- `internal/server/` owns routing, request validation, API errors, static assets, and token checks.
- `internal/app/` adapts web requests to organizer, rename, and Audiobookshelf services without depending on Cobra.
- `web/` contains the Vue/Vite frontend.
- `make web-install` installs frontend dependencies.
- `make web-build` builds assets into `internal/server/static`.
- `make build` packages the embedded frontend into the single binary.

## Behavior That Matters

### Metadata Sources

The application can use:

- `metadata.json`
- embedded EPUB metadata
- embedded MP3 metadata
- embedded M4B metadata

`flat` mode implies embedded metadata and changes grouping behavior.

### Field Mapping

Field mapping is a first-class feature. Before changing metadata extraction behavior, inspect:

- `internal/organizer/types.go`
- `internal/organizer/metadata_providers.go`
- related field mapping tests

Avoid hard-coding one metadata schema if existing field mapping can solve the problem.

### Layouts and Naming

Organization layout handling is central behavior. Changes here can cascade into path generation, preview behavior, logging, and tests. Inspect layout tests before modifying path logic.

Rename behavior is template-driven. Validate both scan/preview output and final rename execution when changing rename logic.

### Undo and Dry-Run

Preserve these invariants:

- Dry-run must not mutate the filesystem.
- Organization operations log to `.abook-org.log`.
- Rename operations log to `.abook-rename.log`.
- Undo must remain compatible with the log format.

## Build And Test Commands

Use repo-native commands first:

```bash
make dev
make test
make test-unit
make test-integration
make test-all
make coverage
make lint
make fmt
make fmt-check
prek run --all-files
```

Useful direct commands:

```bash
go test ./...
go test -short ./...
go test ./internal/organizer/...
go test ./internal/tui/...
go test -run TestName ./path/to/package
```

Web-specific commands:

```bash
make web-install
make web-build
make web-dev
```

ABS-specific commands:

```bash
make abs-ci-smoke
make abs-test-metadata
make abs-test-e2e
```

ABS feature validation:

- Any change that affects Audiobookshelf discovery, path mapping, metadata mode, scan triggering, import/organize behavior, mounted-library behavior, or ABS-facing web/API flows must be reflected in `test/abs/test-matrix.md`.
- Add a matrix row for new behavior, or update the existing row when behavior changes.
- Promote implemented matrix rows into automated coverage through the ABS test matrix workflow in `.github/workflows/test.yml` and the related `make abs-*` target when needed.
- If an ABS-facing change does not need matrix coverage, document the reason in the PR.

## Agent Working Rules

- Prefer `rg` for content searches and `fd` or `find` for file discovery.
- Prefer focused changes that match the existing package boundaries.
- Do not refactor across CLI, TUI, organizer core, web UI, and ABS services simultaneously unless the task requires it.
- Check for tests near the code you are changing and update them with the behavior change.
- For bug fixes, prefer first creating or identifying a failing check that demonstrates the problem, then make that check pass.
- For refactors, verify behavior before and after when practical.
- Match existing project style even when you would design it differently.
- Do not clean up unrelated dead code, comments, formatting, or adjacent abstractions. Mention unrelated issues instead.
- Expect a dirty worktree. Do not revert unrelated changes.
- Verify with the narrowest relevant repo-native command first, then widen if needed.

## High-Risk Areas

Be careful when editing:

- `cmd/root.go` because flag aliasing and Viper binding affect many entrypoints
- `internal/organizer/path.go` because path formatting changes can cause broad regressions
- `internal/organizer/metadata_providers.go` because multiple file formats and fallback rules converge here
- `internal/organizer/types.go` because shared structs are used across CLI, TUI, tests, and web bindings
- `internal/tui/models/` because user flow is spread across multiple state models
- `internal/server/` because token checks and local API behavior affect the browser UI security model
- `internal/app/` because it bridges web requests into organizer, rename, and Audiobookshelf operations

## Recommended Workflow

1. Read the relevant `abo-*` skill and shared reference for the task.
2. Read the command layer and relevant package before editing.
3. Find existing tests for the same behavior.
4. Make the smallest coherent change.
5. Run formatting if needed.
6. Run the most relevant tests or lint target.
7. Update docs and `CHANGELOG.md` when the change is user-visible.
8. Summarize behavior changes and any verification gaps clearly.
