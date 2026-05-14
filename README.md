# Audiobook Organizer for AudiobookShelf

[![codecov](https://codecov.io/gh/jeeftor/audiobook-organizer/branch/master/graph/badge.svg)](https://codecov.io/gh/jeeftor/audiobook-organizer)
[![Coverage Status](https://coveralls.io/repos/github/jeeftor/audiobook-organizer/badge.svg?branch=master)](https://coveralls.io/github/jeeftor/audiobook-organizer?branch=master)

![Audiobook Organizer logo](docs/logo.png)

Audiobook Organizer is a single-binary tool for cleaning up audiobook libraries. It can preview and organize folders, rename files from metadata templates, inspect metadata, and connect to Audiobookshelf for library metadata and scan workflows.

The project now ships one `audiobook-organizer` binary with:

- a local browser UI: `audiobook-organizer web`
- terminal workflows: `audiobook-organizer tui`, `rename-tui`, and `metadata`
- scriptable CLI workflows: root organize command, `rename`, and `abs`

`audiobook-organizer gui` remains as a compatibility alias for `audiobook-organizer web`.

## What It Does

- Organizes audiobooks into predictable directory layouts such as `Author/Series/Title`.
- Reads metadata from `metadata.json`, EPUB, MP3, M4B, and Audiobookshelf.
- Renames files with templates such as `{author} - {series} {series_number} - {title}`.
- Supports dry-run previews before filesystem changes.
- Logs organization and rename operations for undo.
- Handles field mapping for non-standard metadata tags.
- Provides local browser, terminal, and command-line interfaces.
- Includes Audiobookshelf commands for library discovery, path mapping checks, item previews, scan triggers, and WebSocket scan event testing.

## Current Interfaces

| Interface | Command | Best For | Status |
| --- | --- | --- | --- |
| Local web UI | `audiobook-organizer web` | Browser-based preview, configuration, rename preview, ABS connection workflows | Current GUI direction |
| GUI alias | `audiobook-organizer gui` | Existing muscle memory/scripts | Alias for `web` |
| Organize CLI | `audiobook-organizer --dir=/books` | Batch organization and automation | Stable |
| Organization TUI | `audiobook-organizer tui` | Keyboard-first interactive organization | Stable |
| Rename CLI | `audiobook-organizer rename --dir=/books` | Scriptable file renaming | Stable |
| Rename TUI | `audiobook-organizer rename-tui --dir=/books` | Interactive rename previews | Stable |
| Metadata TUI | `audiobook-organizer metadata --dir=/books` | Metadata inspection and template exploration | Stable |
| Audiobookshelf CLI | `audiobook-organizer abs ...` | ABS discovery, path mapping, metadata previews, and scan triggers | Active development |

## Quick Start

### Install

Homebrew:

```bash
brew tap jeeftor/tap
brew install audiobook-organizer
```

Go:

```bash
go install github.com/jeeftor/audiobook-organizer@latest
```

Docker:

```bash
docker pull jeffsui/audiobook-organizer:latest
```

Release archives and Linux packages are available from [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases).

See [docs/INSTALLATION.md](docs/INSTALLATION.md) for platform-specific installation notes.

### Start The Browser UI

```bash
audiobook-organizer web
```

The web UI starts a local HTTP server, binds to `127.0.0.1` by default, generates a temporary session token, and opens your browser.

Useful variants:

```bash
# Pre-fill source and output directories
audiobook-organizer web --input=/path/to/books --output=/path/to/organized

# Choose a local port
audiobook-organizer web --host=127.0.0.1 --port=8080

# Print the URL without opening a browser
audiobook-organizer web --no-open

# Compatibility alias
audiobook-organizer gui
```

The current web API supports preview-oriented organization and rename flows plus Audiobookshelf library, path mapping, item loading, and scan-trigger endpoints. Use the CLI or TUI for full filesystem execution when needed.

See [docs/GUI.md](docs/GUI.md) for the local web UI guide.

### Preview And Organize From The CLI

```bash
# Preview without moving files
audiobook-organizer --dir=/path/to/books --out=/path/to/organized --dry-run

# Organize to a separate output directory
audiobook-organizer --dir=/path/to/books --out=/path/to/organized

# Organize in place
audiobook-organizer --dir=/path/to/books

# Undo the previous organization operation
audiobook-organizer --dir=/path/to/books --undo
```

Common organization flags:

| Flag | Purpose |
| --- | --- |
| `--dir`, `--input` | Source directory |
| `--out`, `--output` | Output directory; defaults to source |
| `--dry-run` | Preview without changing files |
| `--layout` | Directory layout |
| `--use-embedded-metadata` | Use embedded EPUB/MP3/M4B metadata |
| `--flat` | Process files individually; also enables embedded metadata |
| `--remove-empty` | Remove empty source directories after moving |
| `--skip-errors` | Continue past missing or invalid metadata |
| `--author-fields`, `--title-field`, `--series-field`, `--track-field`, `--disc-field` | Field mapping overrides |

See [docs/CLI.md](docs/CLI.md) for the full CLI reference.

### Rename Files

```bash
# Preview renames
audiobook-organizer rename --dir=/path/to/books --dry-run

# Rename using a template
audiobook-organizer rename \
  --dir=/path/to/books \
  --template="{author} - {series} {series_number} - {title}"

# Rename from embedded metadata
audiobook-organizer rename --dir=/path/to/books --use-embedded-metadata

# Undo the previous rename operation
audiobook-organizer rename --dir=/path/to/books --undo

# Open the interactive rename TUI
audiobook-organizer rename-tui --dir=/path/to/books
```

Template help:

```bash
audiobook-organizer rename help-template
```

See [docs/RENAME_FEATURE.md](docs/RENAME_FEATURE.md) for rename templates and examples.

### Use The TUI

```bash
# Interactive organization workflow
audiobook-organizer tui

# Start with paths pre-filled
audiobook-organizer tui --dir=/path/to/books --out=/path/to/organized

# Interactive rename workflow
audiobook-organizer rename-tui --dir=/path/to/books

# Metadata exploration workflow
audiobook-organizer metadata --dir=/path/to/books
```

See [docs/TUI.md](docs/TUI.md) and [docs/METADATA_COMMAND.md](docs/METADATA_COMMAND.md).

## Audiobookshelf Workflows

Audiobookshelf support is available in both the local web UI and the `abs` CLI command group.

Current ABS command group:

```bash
# List/discover libraries and item counts
audiobook-organizer abs scan \
  --abs-url=http://localhost:13378 \
  --abs-token="$ABS_TOKEN"

# Preview ABS metadata with manual container-to-host path mapping
audiobook-organizer abs scan \
  --abs-url=http://localhost:13378 \
  --abs-token="$ABS_TOKEN" \
  --abs-library=Audiobooks \
  --abs-path-map="/audiobooks:/mnt/media/audiobooks" \
  --dir=/mnt/media/audiobooks \
  --check-files

# Test SQLite-backed path mapping discovery
audiobook-organizer abs test-paths \
  --abs-sqlite=/path/to/absdatabase.sqlite \
  --dir=/mnt/media/audiobooks

# Trigger an ABS library scan after filesystem changes
audiobook-organizer abs scan-trigger \
  --abs-url=http://localhost:13378 \
  --abs-token="$ABS_TOKEN" \
  --abs-library=Audiobooks

# Listen for scan events over the ABS WebSocket API
audiobook-organizer abs websocket-test \
  --abs-url=http://localhost:13378 \
  --abs-token="$ABS_TOKEN"
```

ABS metadata-driven organization is still being built out and validated through `test/abs/test-matrix.md`. For already indexed ABS libraries, prefer previewing path mappings and metadata first, then trigger an ABS scan after any filesystem move so ABS can reconcile missing and moved items.

## Metadata Sources

Audiobook Organizer can use:

1. `metadata.json` sidecars, including Audiobookshelf-style metadata.
2. Embedded EPUB metadata.
3. Embedded MP3 ID3 tags.
4. Embedded M4B metadata.
5. Audiobookshelf API metadata for ABS workflows.

Processing modes:

- Non-flat mode, the default, treats each directory as one book or album.
- Flat mode processes each supported file independently and automatically enables embedded metadata.
- Hybrid behavior can combine `metadata.json` book-level data with embedded track-level data when both are present.
- Field mapping lets you choose which raw metadata fields should be treated as author, title, series, track, and disc.

See [docs/METADATA.md](docs/METADATA.md).

## Directory Layouts

Supported layout values:

| Layout | Example |
| --- | --- |
| `author-series-title` | `Author/Series/Title/` |
| `author-series-title-number` | `Author/Series/#1 - Title/` |
| `author-series` | `Author/Series/` |
| `author-title` | `Author/Title/` |
| `author-only` | `Author/` |
| `series-title` | `Series/Title/` |
| `series-title-number` | `Series/#1 - Title/` |

Example:

```bash
audiobook-organizer \
  --dir=/path/to/books \
  --out=/path/to/organized \
  --layout=author-title \
  --dry-run
```

See [docs/LAYOUTS.md](docs/LAYOUTS.md).

## Configuration

Configuration can come from flags, environment variables, or YAML config.

Config lookup:

1. `--config /custom/path.yaml`
2. `./.audiobook-organizer.yaml`
3. `~/.audiobook-organizer.yaml`

Example:

```yaml
dir: "/path/to/audiobooks"
out: "/path/to/organized"
layout: "author-series-title"
use-embedded-metadata: true
remove-empty: true
author-fields: "authors,narrators,album_artist,artist"
title-field: "album"
```

Environment examples:

```bash
export AO_DIR="/path/to/audiobooks"
export AO_OUTPUT="/path/to/organized"
export AO_LAYOUT="author-series-title"
export AO_USE_EMBEDDED_METADATA=true
```

Precedence is: defaults, config file, environment variables, CLI flags.

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md).

## Docker

```bash
# Read/write organization inside one mounted library
docker run --rm \
  -v /media/audiobooks:/books \
  jeffsui/audiobook-organizer --dir=/books --dry-run

# Separate read-only input and writable output
docker run --rm \
  -v /media/source:/input:ro \
  -v /media/organized:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output

# Configure with environment variables
docker run --rm \
  -v /media/audiobooks:/books \
  -e AO_LAYOUT=author-series-title \
  -e AO_VERBOSE=true \
  jeffsui/audiobook-organizer --dir=/books
```

See [docs/CLI.md#docker-usage](docs/CLI.md).

## Development

```bash
git clone https://github.com/jeeftor/audiobook-organizer.git
cd audiobook-organizer

# Build the Go binary
make dev

# Install and build embedded web assets
make web-install
make web-build

# Run unit tests
make test-unit

# Run lint checks
make lint
```

ABS harness commands:

```bash
make abs-ci-smoke
make abs-test-metadata
make abs-test-e2e
```

New or changed ABS-facing features should be reflected in [test/abs/test-matrix.md](test/abs/test-matrix.md) and promoted into the ABS test workflow when implemented.

## Documentation

- [Installation Guide](docs/INSTALLATION.md)
- [Local Web UI Guide](docs/GUI.md)
- [TUI Guide](docs/TUI.md)
- [CLI Reference](docs/CLI.md)
- [Configuration](docs/CONFIGURATION.md)
- [Metadata Guide](docs/METADATA.md)
- [Layout Guide](docs/LAYOUTS.md)
- [Rename Feature](docs/RENAME_FEATURE.md)
- [Metadata Command](docs/METADATA_COMMAND.md)

## Updates

```bash
# Check for updates
audiobook-organizer update --check

# Install the latest release when supported by the install method
audiobook-organizer update
```

## Support

- Bug reports: [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- Feature requests and questions: [GitHub Discussions](https://github.com/jeeftor/audiobook-organizer/discussions)
- Releases: [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases)
- Docker Hub: [jeffsui/audiobook-organizer](https://hub.docker.com/r/jeffsui/audiobook-organizer)
- Homebrew Tap: [jeeftor/tap](https://github.com/jeeftor/homebrew-tap)

## License

MIT License. See [LICENSE](LICENSE).
