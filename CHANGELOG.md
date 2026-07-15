# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **ABS metadata-source organize workflow**: The local web UI can now organize an Audiobookshelf-mounted library directly from validated ABS metadata, with the normal dry-run preview, selected-move review, run log, and ABS scan/cleanup follow-up.
- **Web metadata field mapping**: The organize and rename setup screens now offer
  source presets and editable title, author, series, track, and disc mappings;
  changes refresh the dry-run preview before any filesystem operation.

### Fixed

- **Custom metadata author mappings**: Arrays from `metadata.json` now apply correctly when selected as an author field in the web UI.
- **Web session recovery**: The browser UI now explains how to recover when opened without its required session-token URL parameter.
- **Docker web UI access**: Documented the required `--host=0.0.0.0` bind,
  container port publishing, token handling, and reverse-proxy backend port.

## [v0.13.1] — 2026-07-14

### Added

- **Repo-local AI skills**: Added `abo-*` agent skills for issue workflow, PR workflow, testing, ABS validation, web UI work, and dependency audit/update tasks.
- **Local web UI foundation**: `audiobook-organizer web` starts a browser-based UI from the same binary as the CLI and TUI. `audiobook-organizer gui` remains as a compatibility alias.
- **Embedded web assets**: Release builds now compile the Vue frontend into the Go binary so the app can serve its own UI.
- **Web API surface**: Added initial local REST endpoints for health, initial configuration, option lists, organize preview/run, rename preview/run, Audiobookshelf libraries, ABS path mapping tests, ABS item loading, and ABS scan triggers.
- **Audiobookshelf web workflow**: The new web foundation treats ABS as a first-class source alongside local metadata and embedded metadata.
- **Preview-oriented app service layer**: Added an internal application service that converts web requests into organizer, renamer, and ABS operations without coupling the HTTP layer to Cobra command handling.
- **REST execution coverage for ABS workflows**: Added Docker-backed REST tests for `metadata.json`, embedded metadata import, and flat import workflows against real Audiobookshelf containers.
- **Flat mode ABS coverage**: Added Docker-backed flat mechanics and flat ebook import matrix coverage for Audiobookshelf workflows.
- **Embedded EPUB ABS coverage**: Added Docker-backed already-indexed EPUB current-behavior coverage for embedded metadata workflows.
- **ABS browser UI coverage**: Added Docker-backed Playwright coverage for the web Audiobookshelf setup and operation controls against real ABS instances.
- **Organize browser coverage**: Expanded real Playwright coverage for embedded EPUB metadata, numbered layouts, remove-empty execution, dry-run behavior, undo-log creation, and real backend path validation errors.
- **Rename browser execution**: Added a local web UI rename run endpoint and real Playwright coverage for preview review, confirmed execution, conflict handling, skipped/error summaries, filesystem results, and `.abook-rename.log` guidance.
- **Web UI gating coverage**: Hardened Playwright checks for organize confirmation cancellation, failed preview lockouts, retry recovery, and ABS cleanup acknowledgement gating.
- **ABS metadata organization**: Added `audiobook-organizer abs organize` so already-indexed Audiobookshelf items can be reorganized with ABS API metadata while reusing the normal organizer move, dry-run, undo, layout, logging, and scan follow-up flow.
- **Custom organization layout templates**: Added `--layout-template` and web UI support for template-driven directory layouts such as `{author}/{series}/{series-count} - {title} ({narrator})`.
- **Composite layout template segments**: Custom templates now support optional literal+field composites such as `{Vol series_number:02 - }` and `{ [narrator]}`, numeric padding via `{series_number:02}`, and omission of empty slash-separated layout path segments.
- **Layout template CLI help**: Added `audiobook-organizer layout-template` for an in-terminal field reference, fallback syntax, examples, and path safety rules.
- **TUI custom layout templates**: The organize TUI settings screen now includes a `custom` layout option and editable layout template field.
- **Text-only metadata inspection**: `audiobook-organizer metadata` now prints non-interactive metadata scan results, `metadata --json` writes machine-readable output, and `metadata-tui` keeps the old interactive metadata exploration workflow explicit.
- **Verbose metadata inspection**: `audiobook-organizer metadata --verbose` now prints human-friendly visual labels for docs and manual metadata review.
- **Pretty metadata inspection**: `audiobook-organizer metadata --pretty` now reuses the organizer metadata formatter for source indicators, field mapping context, and hybrid metadata review while preserving the existing JSON output shape.
- **Web UI docs screenshots**: Added a local-only Playwright workflow for generating web UI metadata.json preview, review-plan, and embedded-metadata preview screenshots under ignored `output/docs-visuals/web-ui/`, plus a CI artifact workflow for docs visuals.
- **CLI docs captures**: Added a local-only docs workflow for generating terminal-style CLI help, dry-run organization, metadata inspection, and rename preview PNG captures, plus VHS animated GIFs for short CLI demos, under ignored `output/docs-visuals/cli/`.
- **TUI docs captures**: Added a VHS workflow for generating animated organize and rename TUI captures plus static final-frame PNGs under ignored `output/docs-visuals/tui/`.
- **Docs site publishing**: Added a lightweight generated documentation site, GitHub Pages publishing path for generated visuals, docs link verification, and a concise README landing page.

### Fixed

- **Per-file track prefixes during organize moves**: Multi-chapter book directories now use each audio file's embedded track number instead of applying the first file's track number to every moved file. Single-track audiobooks no longer receive a redundant `01 -` filename prefix.
- **Template narrator field behavior**: `{narrator}` now renders only the first narrator, while `{narrators}` continues to render the full comma-separated narrator list.
- Docker image publishing now builds with Go 1.25 so release tags match the toolchain required by `go.mod`.
- Beta release tags now derive from the latest stable SemVer tag instead of stacking beta suffixes from prior prereleases.
- Frontend embed path is now stable for goreleaser by building into `internal/server/static`.
- Release workflows build the web frontend before packaging the single binary.
- Stable release publishing now uses the supported GoReleaser release command.
- Organize summaries now report directories missing metadata even when verbose console logging is disabled, so web previews show real warning counts.
- Rename previews now report skipped files, extraction errors, and duplicate target conflicts in backend summaries used by the web UI.
- Web activity and review panels now show real request lifecycle events, completed backend summaries, persistent warnings/errors, and undo-log guidance only when a backend log path exists.

### Changed

- `audiobook-organizer layout-template` now points to the hosted GitHub layout guide so installed binaries show a usable docs link.
- Removed the deprecated GUI tree and release packaging. Releases now focus on one `audiobook-organizer` binary with CLI, TUI, ABS, and local web UI entrypoints.
- Improved the README overview, repository metadata, and web page metadata so search results describe the project as an audiobook organizer and renamer for Audiobookshelf with `metadata.json`, EPUB, MP3, and M4B support.
- Rewrote the root README for the single-binary web UI direction and current command surface.
- Split the Audiobookshelf E2E matrix into parallel GitHub Actions jobs for faster feedback.
- Added the local web UI Playwright suite to GitHub Actions so browser regressions run in CI with failure artifacts.
- Added a first-class Audiobookshelf smoke/reset matrix row for the reset, baseline restore, startup, scan, metadata setting, and initial item count contract.
- Clarified agent Gitflow rules for issue branches, worktree hook installation, PR merge strategy, branch verification, and closeout through merge back to `master`.
- Documented protected `master` workflow rules for required checks, auto-merge, and branch cleanup.
- Refined the `abo-workflow` skill to prompt between creating new tracked work and selecting from existing GitHub issues.
- Documented browser binary setup for local web UI E2E checks, including Playwright-managed Chromium and cached Chrome Headless Shell fallback behavior.
- Clarified maintainer workflow guidance so user-facing workflow issues require real E2E acceptance evidence, while mocked UI/API tests remain supplemental.
- Clarified repo-local agent closeout guidance so completed tracked work ends with a next-work recommendation based on open issues and dependency context.
- Clarified issue and PR closeout guidance for user-originated reports that require reporter confirmation or manual validation.
- Restructured the local web UI into explicit organize, rename, and Audiobookshelf workflow stages so configure, dry-run preview, run, and review states are separated.
- Improved local web UI startup handling so config/options loading and fallback states are visible and Audiobookshelf metadata mode stays scoped to the ABS workflow.
- Wired the local web UI organize workflow to real preview and run endpoints with review gating and backend result summaries.
- Wired the local web UI rename workflow to real preview candidates, reviewed execution, backend summaries, and undo-log guidance.
- Wired local web UI Audiobookshelf setup controls to load libraries and validate path mappings with explicit credential input.
- Improved the local web UI Audiobookshelf setup flow so users test the ABS URL/token first, then choose from discovered libraries instead of typing a library ID.
- Wired local web UI Audiobookshelf operation controls to load item metadata, inspect missing/invalid library state, trigger scans, and clean missing items behind destructive-action gating.
- Made the local web UI responsive on narrow browser widths, keeping workflow controls, details, and activity visible without document-level horizontal overflow.
- Added local web UI folder picker and drag-and-drop affordances for source and output path fields, with clear fallback messaging when the browser cannot expose a local path.
- Made local web UI manual testing clearer by adding a configure-step preview action and consolidating embedded metadata selection into the metadata source menu.
- Added local web UI path validation on the configure step so invalid source/output directories are caught before preview requests run.
- Show the reviewed organize and rename plans on the local web UI run step before mutating actions execute.
- Improved local web UI manual testing by replacing the metadata source dropdown with icon buttons, showing preview inputs with an edit setup affordance, and letting users select which organize moves or rename candidates execute from a reviewed preview.
- Combined local web UI setup and dry-run preview into an iterative first stage with stale-preview handling, a focused planned-change review stage, and run results on the final stage.
- Refined the local web UI iterative flow so setup option changes auto-refresh previews and review/run selection share one final workflow stage.
- Improved local web UI manual testing with a reusable template builder, colored layout/rename fields, relative local preview paths, completed move lists after execution, and automatic fallback from missing `metadata.json` previews to embedded file metadata.
- Updated the layout guide and docs workflow to use public-domain book examples and added more table-based layout guidance.
- Surfaced the Audiobookshelf lifecycle and missing-item cleanup screenshots on Getting Started and made the ABS metadata settings screenshot a compact side-by-side callout.
- Clarified the README as a compact project landing page and added a prominent link to the generated documentation site.
- Added an Astro Starlight documentation site proof of concept with local preview and static build scripts.
- Improved generated docs media delivery with homepage image sizing/lazy loading, WebP companions for static captures, and optional `gifsicle` optimization for animated GIF captures.

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
