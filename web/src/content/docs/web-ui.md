---
title: "Local Web UI"
description: "Use the local browser interface for Audiobook Organizer."
---

The Audiobook Organizer web UI is served by the same `audiobook-organizer` binary as the CLI and TUI. It runs on localhost, opens in your browser, and reuses the existing organizer, rename, metadata, and Audiobookshelf code paths through an internal API layer.

## Launch

```bash
# Canonical command
audiobook-organizer web

# Compatibility alias
audiobook-organizer gui

# Start with directories pre-filled
audiobook-organizer web --input=/path/to/audiobooks --output=/path/to/organized

# Bind a specific local address
audiobook-organizer web --host=127.0.0.1 --port=8080

# Print the URL without opening the browser
audiobook-organizer web --no-open
```

The server generates a temporary token at startup. The browser URL includes that token, and API requests can also pass it with `X-Audiobook-Organizer-Token` or `Authorization: Bearer`. If you open the UI without the token, reopen the complete startup URL.

## Interface

The first web UI pass is an operational dashboard:

- Left workflow panel for source, output, scan mode, fixed or custom layout, and Audiobookshelf connection settings.
- Center table for scanned books, metadata health, destination preview, and conflicts.
- Right inspector for selected book metadata, field mapping, rename template, and ABS details.
- Bottom job console for scan, preview, and API activity.

The UI is intentionally browser-based instead of native-desktop-specific. That keeps releases to one binary and avoids platform-specific desktop runtime packaging.

## Local Screenshots

Generate local web UI screenshots from the repository root. The output files are local-only and ignored by git:

```bash
make docs-web-screenshots
```

The workflow builds the embedded web frontend, starts the real local Go web server with `--no-open`, drives the browser with Playwright, captures three populated states: a generated `metadata.json` preview, the matching Review & Run plan screen, and an embedded metadata preview, then writes PNG assets under the ignored `output/docs-visuals/web-ui/` directory. It copies committed LibriVox sample media from `testdata/mp3flat/` into generated local sample data under `output/docs-web-ui-sample/`, so the demo uses real public-domain audio while paths stay stable and do not include machine-specific absolute directories.

If Playwright-managed Chromium is not installed, run:

```bash
cd web && npm run install:browsers
```

You can also point the workflow at an existing Chrome or Chrome Headless Shell binary with `ABO_DOCS_BROWSER_EXECUTABLE_PATH=/path/to/chrome make docs-web-screenshots`. Containerized visual generation is tracked separately in #148.

## Audiobookshelf

The web API exposes Audiobookshelf workflow endpoints for:

- Listing libraries.
- Testing host path to container path mappings.
- Loading item metadata from ABS.
- Triggering library scans after organization.

To organize directly from ABS metadata, select **Audiobookshelf metadata** in the **Organize** workflow. Test the connection, choose a library, and validate its ABS-to-local path mapping before the dry-run preview can run. The review and selected-move run stages use the same validated connection. Use the separate **Audiobookshelf** workflow when you need inspection, scanning, or missing-item cleanup.

## Custom Layouts

The organize workflow includes a custom layout template field next to the fixed layout selector. When set, the custom template overrides the selected layout for preview and run requests.

Example template:

```text
{author}/{series}/{series-count} - {title} ({narrator})
```

## Metadata Field Mapping

The Organize and Rename setup screens include a **Metadata Field Mapping** panel.
Start with the `metadata.json`, audio, or EPUB preset, then update the title,
author, series, track, or disc field names when your metadata uses different
keys. Enter more than one author field as a comma-separated list.

Every mapping edit refreshes the dry-run preview. Review the updated destination
or filename candidates before continuing to the run stage.

## Development

```bash
# Install frontend dependencies
make web-install

# Build Vue assets into internal/server/static
make web-build

# Run the Go server with embedded assets
go run . web --host=127.0.0.1 --port=8080 --no-open

# Run the Vite dev server for frontend-only iteration
make web-dev
```

The production build embeds `internal/server/static` into the Go binary. The Vite dev server proxies `/api` to `http://127.0.0.1:8080`.

See [GUI_TESTING.md](/audiobook-organizer/development/web-ui-testing/) for the Playwright-based GUI test framework.

## Related Commands

- `audiobook-organizer tui` for keyboard-first terminal workflows.
- `audiobook-organizer rename-tui` for interactive rename workflows.
- `audiobook-organizer abs` for scriptable Audiobookshelf operations.
- `audiobook-organizer --dir=/path --dry-run` for CLI previews.
