# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] — v0.12.0

### Added

- **Repo-local AI skills**: Added `abo-*` agent skills for issue workflow, PR workflow, testing, ABS validation, web UI work, and dependency audit/update tasks.
- **Local web UI foundation**: `audiobook-organizer web` starts a browser-based UI from the same binary as the CLI and TUI. `audiobook-organizer gui` remains as a compatibility alias.
- **Embedded web assets**: Release builds now compile the Vue frontend into the Go binary so the app can serve its own UI.
- **Web API surface**: Added initial local REST endpoints for health, initial configuration, option lists, organize preview, rename preview, Audiobookshelf libraries, ABS path mapping tests, ABS item loading, and ABS scan triggers.
- **Audiobookshelf web workflow**: The new web foundation treats ABS as a first-class source alongside local metadata and embedded metadata.
- **Preview-oriented app service layer**: Added an internal application service that converts web requests into organizer, renamer, and ABS operations without coupling the HTTP layer to Cobra command handling.
- **REST execution coverage for ABS workflows**: Added Docker-backed REST tests for `metadata.json`, embedded metadata import, and flat import workflows against real Audiobookshelf containers.
- **Text-only metadata inspection**: `audiobook-organizer metadata` now prints non-interactive metadata scan results, `metadata --json` writes machine-readable output, and `metadata-tui` keeps the old interactive metadata exploration workflow explicit.

### Fixed

- Beta release tags now derive from the latest stable SemVer tag instead of stacking beta suffixes from prior prereleases.
- Frontend embed path is now stable for goreleaser by building into `internal/server/static`.
- Release workflows build the web frontend before packaging the single binary.
- Organize summaries now report directories missing metadata even when verbose console logging is disabled, so web previews show real warning counts.
- Rename previews now report skipped files, extraction errors, and duplicate target conflicts in backend summaries used by the web UI.
- Web activity and review panels now show real request lifecycle events, completed backend summaries, persistent warnings/errors, and undo-log guidance only when a backend log path exists.

### Changed

- Removed the deprecated GUI tree and release packaging. Releases now focus on one `audiobook-organizer` binary with CLI, TUI, ABS, and local web UI entrypoints.
- Improved the README overview, repository metadata, and web page metadata so search results describe the project as an audiobook organizer and renamer for Audiobookshelf with `metadata.json`, EPUB, MP3, and M4B support.
- Rewrote the root README for the single-binary web UI direction and current command surface.
- Split the Audiobookshelf E2E matrix into parallel GitHub Actions jobs for faster feedback.
- Added a first-class Audiobookshelf smoke/reset matrix row for the reset, baseline restore, startup, scan, metadata setting, and initial item count contract.
- Clarified agent Gitflow rules for issue branches, worktree hook installation, PR merge strategy, branch verification, and closeout through merge back to `master`.
- Documented protected `master` workflow rules for required checks, auto-merge, and branch cleanup.
- Refined the `abo-workflow` skill to prompt between creating new tracked work and selecting from existing GitHub issues.
- Documented browser binary setup for local web UI E2E checks, including Playwright-managed Chromium and cached Chrome Headless Shell fallback behavior.
- Clarified maintainer workflow guidance so user-facing workflow issues require real E2E acceptance evidence, while mocked UI/API tests remain supplemental.
- Restructured the local web UI into explicit organize, rename, and Audiobookshelf workflow stages so configure, dry-run preview, run, and review states are separated.
- Improved local web UI startup handling so config/options loading and fallback states are visible and Audiobookshelf metadata mode stays scoped to the ABS workflow.
- Wired the local web UI organize workflow to real preview and run endpoints with review gating and backend result summaries.
- Wired the local web UI rename workflow to real preview candidates while keeping rename execution explicitly deferred.
- Wired local web UI Audiobookshelf setup controls to load libraries and validate path mappings with explicit credential input.
- Wired local web UI Audiobookshelf operation controls to load item metadata, inspect missing/invalid library state, trigger scans, and clean missing items behind destructive-action gating.

---

## [v0.11.0] — 2026-01-02

### Added

- New layout options (see `docs/LAYOUTS.md` for full list).
- Environment variables now properly recognized for all flags — fixes [#17](https://github.com/jeeftor/audiobook-organizer/issues/17).

### Fixed

- `metadata.json` parsing edge cases.
- Trailing underscore stripping in sanitized paths.

---

## [v0.10.0] and earlier

See [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases) for earlier release notes.
