# Local Web UI Guide

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

The server generates a temporary token at startup. The browser URL includes that token, and API requests can also pass it with `X-Audiobook-Organizer-Token` or `Authorization: Bearer`.

## Interface

The first web UI pass is an operational dashboard:

- Left workflow panel for source, output, scan mode, fixed or custom layout, and Audiobookshelf connection settings.
- Center table for scanned books, metadata health, destination preview, and conflicts.
- Right inspector for selected book metadata, field mapping, rename template, and ABS details.
- Bottom job console for scan, preview, and API activity.

The UI is intentionally browser-based instead of native-desktop-specific. That keeps releases to one binary and avoids platform-specific desktop runtime packaging.

## Audiobookshelf

The web API exposes Audiobookshelf workflow endpoints for:

- Listing libraries.
- Testing host path to container path mappings.
- Loading item metadata from ABS.
- Triggering library scans after organization.

Use the ABS controls in the web UI when you want the organizer to reconcile local filesystem paths with an Audiobookshelf server.

## Custom Layouts

The organize workflow includes a custom layout template field next to the fixed layout selector. When set, the custom template overrides the selected layout for preview and run requests.

Example template:

```text
{author}/{series}/{series-count} - {title} ({narrator})
```

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

See [GUI_TESTING.md](GUI_TESTING.md) for the Playwright-based GUI test framework.

## Related Commands

- `audiobook-organizer tui` for keyboard-first terminal workflows.
- `audiobook-organizer rename-tui` for interactive rename workflows.
- `audiobook-organizer abs` for scriptable Audiobookshelf operations.
- `audiobook-organizer --dir=/path --dry-run` for CLI previews.
