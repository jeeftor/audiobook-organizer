# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] — v0.12.0

### Added

- **Unified binary**: `audiobook-organizer gui` now launches the full desktop GUI window from the same binary as the CLI and TUI. No separate download required.
- **Native macOS directory dialog**: File picker uses a native `NSOpenPanel` dispatched on the main thread, preventing the dialog from being immediately dismissed when the app is not frontmost.
- **WebKit developer tools**: Pass `--devtools` flag to enable right-click → Inspect and a _Developer_ menu with _Open Inspector_.
- **Drag-and-drop**: Files and folders can be dragged onto the GUI window.
- **CLI command in preview panel**: The Preview screen now shows the equivalent bash command for the planned operation, making it easy to script or audit changes.
- **Book-grouped preview**: Preview screen groups file moves by book/album for easier review.
- **Skipped-file report**: Files that were excluded from an operation are now reported separately so nothing is silently dropped.
- **Undo support in GUI**: The Results screen includes an Undo button that reverses the last organization operation using the `.abook-org.log` file.
- **Selection filtering**: Only checked/selected files are passed to the organizer — previously all scanned files could be processed regardless of selection state.
- **macOS DMG artifact**: Beta releases now include a `.dmg` with a `README.txt` covering Gatekeeper and quarantine instructions.

### Fixed

- Linux CI: `webkit2gtk-4.0` pkg-config shims and library symlinks added so Wails builds succeed on Ubuntu 24.04 (which ships only `webkit2gtk-4.1`).
- macOS CI: `extldflags '-framework UniformTypeIdentifiers'` added to the darwin linker flags for the unified binary.
- Frontend embed path resolved before goreleaser runs to prevent `all:frontend/dist` embed failures.
- Various CI workflow failures across build, test, and release pipelines.

### Changed

- Beta releases now distribute the **unified CLI+GUI binary** (`audiobook-organizer`) instead of the standalone Wails binary (`audiobook-organizer-gui`).

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
