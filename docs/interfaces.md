# Choose An Interface

Audiobook Organizer ships one binary with browser, terminal, scriptable CLI, and Audiobookshelf workflows. Pick the interface based on how much review and automation you need.

| Interface | Command | Best For | Notes |
| --- | --- | --- | --- |
| Local web UI | `audiobook-organizer web` | Visual setup, preview review, rename review, Audiobookshelf connection checks | Current GUI direction |
| Organize | `audiobook-organizer --dir=/books` | Main workflow for repeatable scripts and batch runs | Use `--dry-run` first |
| Organization TUI | `audiobook-organizer tui` | Keyboard-first interactive organization | Good when you want terminal review without browser setup |
| Rename Files | `audiobook-organizer rename --dir=/books` | Scripted rename previews and execution | Template-driven |
| Rename TUI | `audiobook-organizer rename-tui --dir=/books` | Interactive rename review | Supports field mapping and template preview |
| Explore Metadata | `audiobook-organizer metadata --dir=/books` | Text-only metadata inspection | Use `--json` for machine-readable output |
| Explore Metadata TUI | `audiobook-organizer metadata-tui --dir=/books` | Interactive metadata exploration | Useful before writing templates |
| Audiobookshelf CLI | `audiobook-organizer abs ...` | ABS library discovery, path mapping, metadata organization, scans | Requires ABS URL/token |

## Web UI

Use the web UI when you want forms, staged review, selected moves/candidates, and visible backend summaries:

```bash
audiobook-organizer web
```

See [Local Web UI](GUI.md).

## CLI

Use organize for automation and repeatable runs:

```bash
audiobook-organizer --dir=/books/source --out=/books/organized --dry-run
```

See [Organize](organize.md) and [CLI](CLI.md).

## TUI

Use the TUI when you want an interactive terminal flow:

```bash
audiobook-organizer tui --dir=/books/source --out=/books/organized
```

See [TUI](TUI.md).

## Audiobookshelf

Use ABS workflows when the server already has the metadata you trust:

```bash
audiobook-organizer abs scan --abs-url=http://localhost:13378 --abs-token="$ABS_TOKEN"
```

See [Audiobookshelf](audiobookshelf.md).
