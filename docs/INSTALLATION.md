# Installation Guide

Audiobook Organizer ships as one binary named `audiobook-organizer`. The same binary provides the CLI, TUI, Audiobookshelf commands, and local browser-based web UI.

## Commands

| Command | Mode | Description |
|---------|------|-------------|
| `audiobook-organizer web` | Local Web UI | Starts a localhost server and opens the browser UI |
| `audiobook-organizer gui` | Local Web UI alias | Compatibility alias for `web` |
| `audiobook-organizer tui` | Interactive Terminal | Keyboard-driven terminal workflow |
| `audiobook-organizer rename-tui` | Rename TUI | Interactive rename workflow |
| `audiobook-organizer metadata` | Metadata CLI | Text-only metadata inspection |
| `audiobook-organizer metadata-tui` | Metadata TUI | Interactive metadata exploration |
| `audiobook-organizer abs` | Audiobookshelf CLI | ABS libraries, mappings, item loading, and scan triggers |
| `audiobook-organizer --dir=/path` | CLI | Scriptable organization |

## macOS

```bash
brew tap jeeftor/tap
brew install audiobook-organizer

audiobook-organizer web
```

You can also download the macOS archive from GitHub Releases and place the `audiobook-organizer` binary somewhere on your `PATH`.

## Linux

Install the release package for your distribution, or download the archive from GitHub Releases.

```bash
# Debian/Ubuntu package, if published for the release
sudo apt install audiobook-organizer

# RedHat/Fedora package, if published for the release
sudo yum install audiobook-organizer

# Alpine package, if published for the release
sudo apk add audiobook-organizer

audiobook-organizer web
```

The web UI runs in your existing browser and does not require native desktop runtime packages.

## Windows

Download `audiobook-organizer-windows-amd64.zip` from GitHub Releases, extract it, and run:

```powershell
.\audiobook-organizer.exe web
```

For terminal workflows:

```powershell
.\audiobook-organizer.exe tui
.\audiobook-organizer.exe --dir C:\Audiobooks --out C:\Organized --dry-run
```

## Docker

```bash
docker pull jeffsui/audiobook-organizer:latest

docker run --rm \
  -v /path/to/audiobooks:/books \
  -v /path/to/output:/output \
  jeffsui/audiobook-organizer --dir=/books --out=/output --dry-run
```

The local web UI is primarily intended for host installs. For Docker, prefer CLI or TUI workflows unless you explicitly publish and secure a local port.

## Go Install

```bash
go install github.com/jeeftor/audiobook-organizer@latest

audiobook-organizer version
audiobook-organizer web
```

## Build From Source

```bash
git clone https://github.com/jeeftor/audiobook-organizer.git
cd audiobook-organizer

# Build the Go binary for local development
make dev

# Install and build embedded web assets
make web-install
make web-build

# Run tests
make test
```

## Verify Installation

```bash
audiobook-organizer version
audiobook-organizer --help
audiobook-organizer web --help
audiobook-organizer tui --help
audiobook-organizer abs --help
```

## Troubleshooting

### Command Not Found

Make sure the directory containing `audiobook-organizer` is on your `PATH`.

### Browser Does Not Open

Run with `--no-open`, copy the printed URL, and open it manually:

```bash
audiobook-organizer web --no-open
```

### Port Already In Use

Use a different port or let the app choose one:

```bash
audiobook-organizer web --port=0
```

### Permission Problems Moving Files

Run a dry-run first and check read/write access to both directories:

```bash
audiobook-organizer --dir=/books --out=/organized --dry-run
```

## See Also

- [CLI.md](CLI.md) - Command-line usage
- [GUI.md](GUI.md) - Local Web UI guide
- [TUI.md](TUI.md) - Terminal UI guide
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file setup
- [Main README](../README.md) - Project overview
