# Audiobook Organizer for Audiobookshelf (and More)

Audiobook Organizer is a single Go binary for previewing, organizing, renaming, and inspecting audiobook libraries. It works with local folders, Audiobookshelf libraries, `metadata.json`, embedded EPUB/MP3/M4B metadata, and metadata field mappings.

[![codecov](https://codecov.io/gh/jeeftor/audiobook-organizer/branch/master/graph/badge.svg)](https://codecov.io/gh/jeeftor/audiobook-organizer)
[![Coverage Status](https://coveralls.io/repos/github/jeeftor/audiobook-organizer/badge.svg?branch=master)](https://coveralls.io/github/jeeftor/audiobook-organizer?branch=master)

![Audiobook Organizer web UI preview](https://jeeftor.github.io/audiobook-organizer/assets/generated/web-ui/web-ui-metadata-json-preview.png)

![Animated CLI organize run](https://jeeftor.github.io/audiobook-organizer/assets/generated/cli/cli-organize-run.gif)

![Animated TUI organize preview](https://jeeftor.github.io/audiobook-organizer/assets/generated/tui/tui-organize-preview.gif)

Use it when you want to:

- preview filesystem changes before moving or renaming audiobook files;
- organize messy folders into predictable layouts such as `Author/Series/Title`;
- rename files from metadata templates;
- inspect and map metadata from sidecar JSON, embedded tags, and Audiobookshelf;
- trigger Audiobookshelf scans after filesystem changes;
- undo organization and rename operations from generated logs.

## Audiobookshelf `metadata.json` Setup

If Audiobookshelf is your metadata source, enable **Store metadata with item** in the Audiobookshelf library settings before organizing. Audiobookshelf will write a `metadata.json` sidecar into each book directory when metadata is generated or updated, and Audiobook Organizer can use those sidecars as the default metadata source.

![Audiobookshelf setting for storing metadata.json files](docs/store_metadata.jpg)

After a real organization run, Audiobookshelf may briefly show old paths as missing until the library scans and reconciles the moved files. See [Audiobookshelf](docs/audiobookshelf.md) for the setup, cleanup screenshots, path mapping checks, and scan workflow.

## Install

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

Release archives and Linux packages are available from [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases). See the [installation guide](docs/INSTALLATION.md) for package details and platform notes.

## Safe First Run

Start with a dry run against a small source folder and a separate output directory:

```bash
audiobook-organizer \
  --dir=/path/to/books \
  --out=/path/to/organized \
  --dry-run \
  --verbose
```

Review the planned paths and warnings. Run for real only after the preview looks right:

```bash
audiobook-organizer \
  --dir=/path/to/books \
  --out=/path/to/organized
```

Organization writes `.abook-org.log`; rename writes `.abook-rename.log`. Keep those logs until you are satisfied with the result. See [Getting Started](docs/getting-started.md) and [Safety And Undo](docs/safety-and-undo.md).

## Choose A Workflow

| Interface | Command | Best For |
| --- | --- | --- |
| Local web UI | `audiobook-organizer web` | Browser setup, visual previews, rename review, Audiobookshelf connection checks |
| Organize | `audiobook-organizer --dir=/books` | Main workflow for repeatable scripts and batch organization |
| Organization TUI | `audiobook-organizer tui` | Keyboard-first interactive organization |
| Rename Files | `audiobook-organizer rename --dir=/books` | Scriptable filename cleanup |
| Rename TUI | `audiobook-organizer rename-tui --dir=/books` | Interactive rename previews and field mapping |
| Explore Metadata | `audiobook-organizer metadata --dir=/books` | Text or JSON metadata inspection |
| Audiobookshelf CLI | `audiobook-organizer abs ...` | ABS library discovery, path mapping, organization, and scan workflows |

Full comparison: [Choose An Interface](docs/interfaces.md).

## Common Commands

Start the local browser UI:

```bash
audiobook-organizer web
```

Preview an organization run:

```bash
audiobook-organizer --dir=/books/source --out=/books/organized --dry-run
```

Rename files with a template:

```bash
audiobook-organizer rename \
  --dir=/books/source \
  --template="{author} - {series} {series_number} - {title}" \
  --dry-run
```

Inspect metadata:

```bash
audiobook-organizer metadata --dir=/books/source
```

Check Audiobookshelf path mapping:

```bash
audiobook-organizer abs scan \
  --abs-url=http://localhost:13378 \
  --abs-token="$ABS_TOKEN" \
  --abs-path-map="/audiobooks:/mnt/media/audiobooks" \
  --dir=/mnt/media/audiobooks \
  --check-files
```

## Documentation

Published docs: <https://jeeftor.github.io/audiobook-organizer/>

Repo docs:

- [Installation](docs/INSTALLATION.md)
- [Getting Started](docs/getting-started.md)
- [Choose An Interface](docs/interfaces.md)
- [Organize](docs/organize.md)
- [Rename Files](docs/RENAME_FEATURE.md)
- [Explore Metadata](docs/explore-metadata.md)
- [Local Web UI](docs/GUI.md)
- [CLI](docs/CLI.md)
- [TUI](docs/TUI.md)
- [Audiobookshelf](docs/audiobookshelf.md)
- [Metadata Sources](docs/METADATA.md)
- [Layouts](docs/LAYOUTS.md)
- [Configuration](docs/CONFIGURATION.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Changelog](CHANGELOG.md)

## Generated Visuals

Pull requests upload generated screenshots and GIFs as short-lived Actions artifacts for review. `master` publishes the same generated visuals to stable GitHub Pages paths under `assets/generated/`.

Regenerate locally:

```bash
make docs-publish-site
```

See [Docs Visuals](docs/development/docs-visuals.md).

## Development

Common checks:

```bash
make test
make web-build
make docs-verify
```

See [AGENTS.md](AGENTS.md) for repository workflow rules and maintainer guidance.
