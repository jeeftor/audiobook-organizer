# Audiobook Organizer for Audiobookshelf (and More)

Audiobook Organizer is a single Go binary for previewing, organizing, renaming, and inspecting audiobook libraries. It works with local folders, Audiobookshelf libraries, `metadata.json`, embedded EPUB/MP3/M4B metadata, and metadata field mappings.

[![codecov](https://codecov.io/gh/jeeftor/audiobook-organizer/branch/master/graph/badge.svg)](https://codecov.io/gh/jeeftor/audiobook-organizer)
[![Coverage Status](https://coveralls.io/repos/github/jeeftor/audiobook-organizer/badge.svg?branch=master)](https://coveralls.io/github/jeeftor/audiobook-organizer?branch=master)

## [Open Documentation](https://jeeftor.github.io/audiobook-organizer/)

Use this README as the quick project overview. The generated documentation site is the primary home for installation details, first-run guidance, workflow pages, troubleshooting, and generated screenshots/GIFs.

[Getting Started](https://jeeftor.github.io/audiobook-organizer/getting-started.html) |
[Installation](https://jeeftor.github.io/audiobook-organizer/installation.html) |
[Choose An Interface](https://jeeftor.github.io/audiobook-organizer/interfaces.html) |
[Audiobookshelf](https://jeeftor.github.io/audiobook-organizer/audiobookshelf.html)

![Audiobook Organizer web UI preview](https://jeeftor.github.io/audiobook-organizer/assets/generated/web-ui/web-ui-metadata-json-preview.png)

![Animated CLI organize run](https://jeeftor.github.io/audiobook-organizer/assets/generated/cli/cli-organize-run.gif)

![Animated TUI organize preview](https://jeeftor.github.io/audiobook-organizer/assets/generated/tui/tui-organize-preview.gif)

Use it when you want to:

- preview filesystem changes before moving or renaming audiobook files;
- organize messy folders into predictable layouts such as `Author/Series/Title`;
- rename files from metadata templates;
- inspect and map metadata from `metadata.json`, embedded tags, and Audiobookshelf;
- trigger Audiobookshelf scans after filesystem changes;
- undo organization and rename operations from generated logs.

## Audiobookshelf `metadata.json` Workflow

If Audiobookshelf is your metadata source, configure ABS to store metadata beside each book before organizing. In the Audiobookshelf library settings, enable **Store metadata with item**. When ABS metadata is generated or updated, Audiobookshelf writes a `metadata.json` file into each book directory.

![Audiobookshelf setting for storing metadata.json files](docs/store_metadata.jpg)

That `metadata.json` file is the safest first metadata source for local organization because it keeps the book-level title, author, series, narrator, and year data next to the audio files:

```text
/audiobooks/The Case of Charles Dexter Ward/
  metadata.json
  01 - Chapter 1.mp3
  02 - Chapter 2.mp3
```

Preview the organizer against those `metadata.json` files before moving files:

```bash
audiobook-organizer \
  --dir=/audiobooks \
  --out=/organized-audiobooks \
  --dry-run \
  --verbose
```

When `metadata.json` exists beside MP3 or M4B files, Audiobook Organizer can use hybrid metadata: book-level fields come from `metadata.json`, while track-level fields can come from embedded audio tags. If your library does not have `metadata.json` files, use embedded metadata mode instead:

```bash
audiobook-organizer \
  --dir=/audiobooks \
  --out=/organized-audiobooks \
  --use-embedded-metadata \
  --dry-run
```

Audiobook Organizer moves files on disk; it does not rewrite Audiobookshelf database rows directly. After a real organization run, Audiobookshelf may briefly show old paths as missing until the library scans and reconciles the moved files. If that happens, open the ABS **Issues** view:

![Audiobookshelf issues view showing missing books](docs/issues.jpg)

Then use the missing-books cleanup action:

![Audiobookshelf remove missing books action](docs/remove_books.jpg)

The **Enable folder watcher for library** setting may help ABS detect some moved files, but a deliberate scan after filesystem changes is still the safer habit. See [Audiobookshelf](docs/audiobookshelf.md) for the full setup, cleanup, path mapping checks, and scan workflow.

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

To run the local web UI in Docker, bind the server to all container interfaces,
publish the container port, and keep the session-token URL private:

```bash
docker run --rm -p 8080:8080 \
  -v /path/to/audiobooks:/books \
  -v /path/to/output:/output \
  jeffsui/audiobook-organizer:latest \
  web --host=0.0.0.0 --port=8080 --no-open
```

Open the tokenized URL printed in the container logs at
`http://localhost:8080/`. When using Traefik or another reverse proxy, route
to container port `8080`, not the host-side published port.

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
| Explore Metadata | `audiobook-organizer metadata --dir=/books` | Text, pretty, or JSON metadata inspection |
| Audiobookshelf CLI | `audiobook-organizer abs ...` | ABS library discovery, path mapping, organization, and scan workflows |

Full comparison: [Choose An Interface](docs/interfaces.md).

## Common Tasks

| Task | Use |
| --- | --- |
| Organize books into `Author/Series/Title` | `audiobook-organizer --dir=/books --layout=author-series-title --dry-run` |
| Rename files from title, author, series, track, or disc fields | `audiobook-organizer rename --dir=/books --dry-run` |
| No `metadata.json`, but audio files have tags | `audiobook-organizer --dir=/books --use-embedded-metadata --dry-run` |
| Flat folder of individual audiobooks | `audiobook-organizer --dir=/books --flat --dry-run` |
| MP3 tags use non-standard fields | map fields with `--author-fields`, `--title-field`, `--series-field`, `--track-field`, or `--disc-field` |
| Previous organization needs to be reverted | `audiobook-organizer --dir=/books --undo` |

See [Organize](docs/organize.md), [Explore Metadata](docs/explore-metadata.md), [Metadata Sources](docs/METADATA.md), and [Safety And Undo](docs/safety-and-undo.md).

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
audiobook-organizer metadata --dir=/books/source --pretty
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

The generated documentation site is the canonical long-form guide:

<https://jeeftor.github.io/audiobook-organizer/>

The README should stay a compact GitHub landing page: what the tool does, how to install it, how to run safely, and where Audiobookshelf users must start. The docs site owns the full workflow pages, troubleshooting, detailed reference material, and generated visual demos.

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
