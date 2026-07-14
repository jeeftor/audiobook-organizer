# Installation Guide

Audiobook Organizer ships as one binary named `audiobook-organizer`. Install the binary first, then make sure your audiobook folders contain metadata that the organizer can read.

## Pre-Requirement: Configure Audiobookshelf

<section class="media-callout">
  <a class="media-callout-image" href="https://github.com/jeeftor/audiobook-organizer/blob/master/docs/store_metadata.jpg" target="_blank" rel="noopener">
    <img src="store_metadata.jpg" alt="Audiobookshelf setting for storing metadata.json files">
  </a>
  <div class="media-callout-copy">
    <p>If you use Audiobookshelf, do this before the first organize run. Configure ABS to write <code>metadata.json</code> files into the same directories as your books.</p>
    <p>In the Audiobookshelf library settings, enable <strong>Store metadata with item</strong>.</p>
    <p>When that setting is enabled, Audiobookshelf writes a <code>metadata.json</code> file beside each book when metadata is generated or updated.</p>
  </div>
</section>

After Audiobookshelf saves metadata with each book, a folder should look like this:

```text
/audiobooks/The Case of Charles Dexter Ward/
  metadata.json
  01 - Chapter 1.mp3
  02 - Chapter 2.mp3
```

Audiobook Organizer reads that `metadata.json` file to get the book title, author, series, narrator, and year before it plans folder names. If your library does not have `metadata.json` files, use embedded metadata mode instead. See [Audiobookshelf](audiobookshelf.md) for the full ABS setup and cleanup flow, or [Explore Metadata](explore-metadata.md) to inspect what metadata the organizer can read.

After a real organize run, Audiobookshelf may show old paths as missing until it scans and reconciles moved files. The **Enable folder watcher for library** setting may help, but you should still trigger a scan and clean up stale missing-book entries when needed. See [After Organizing: Scan And Clean Up Missing Items](audiobookshelf.md#after-organizing-scan-and-clean-up-missing-items).

## Install Audiobook Organizer

Choose the installation method that matches how you want to run the tool.

### Direct Download From GitHub

Download the latest release for your platform from:

<https://github.com/jeeftor/audiobook-organizer/releases/latest>

Use this option when you want a release archive or package file without Homebrew, Go, or Docker. After downloading, extract the archive or install the package, then verify the binary:

```bash
audiobook-organizer version
```

### macOS

```bash
brew tap jeeftor/tap
brew install audiobook-organizer
```

You can also download the macOS archive from GitHub Releases and place the `audiobook-organizer` binary somewhere on your `PATH`.

### Linux

Download the package or archive for your platform from [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases). The examples below assume you downloaded a release package file into the current directory. The project does not currently document an APT, Yum/DNF, or APK repository that makes `apt install audiobook-organizer` work without first downloading or configuring a package source.

```bash
# Debian/Ubuntu .deb package
sudo apt install ./audiobook-organizer_*_linux_amd64.deb

# RedHat/Fedora .rpm package
sudo dnf install ./audiobook-organizer-*.x86_64.rpm

# Alpine .apk package
sudo apk add --allow-untrusted ./audiobook-organizer-*.apk
```

The web UI runs in your existing browser and does not require native desktop runtime packages.

### Windows

Download `audiobook-organizer-windows-amd64.zip` from GitHub Releases, extract it, and run:

```powershell
.\audiobook-organizer.exe version
```

### Docker

```bash
docker pull jeffsui/audiobook-organizer:latest

docker run --rm \
  -v /path/to/audiobooks:/books \
  -v /path/to/output:/output \
  jeffsui/audiobook-organizer --dir=/books --out=/output --dry-run
```

To run the local web UI in Docker, bind the server to all container interfaces
and publish the container port:

```bash
docker run --rm -p 8080:8080 \
  -v /path/to/audiobooks:/books \
  -v /path/to/output:/output \
  jeffsui/audiobook-organizer:latest \
  web --host=0.0.0.0 --port=8080 --no-open
```

Open the tokenized URL printed in the container logs at
`http://localhost:8080/`. Keep that URL private. When using Traefik or another
reverse proxy, route to container port `8080`, not the host-side published
port.

### Go Install

```bash
go install github.com/jeeftor/audiobook-organizer@latest

audiobook-organizer version
```

### Build From Source

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
```

## Choose A First Workflow

After installation and metadata setup:

- Use [Getting Started](getting-started.md) for a safe dry-run first pass.
- Use [Choose An Interface](interfaces.md) to pick web UI, CLI, TUI, rename, metadata, or ABS workflows.
- Use [Audiobookshelf](audiobookshelf.md) when ABS metadata, path mapping, scans, or missing-item cleanup are involved.

## Troubleshooting

### Command Not Found

Make sure the directory containing `audiobook-organizer` is on your `PATH`.

### Browser Does Not Open

If the browser UI does not open automatically, run it with `--no-open`, copy the printed URL, and open it manually:

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

- [Getting Started](getting-started.md) - Safe first run
- [Audiobookshelf](audiobookshelf.md) - ABS metadata setup, scans, and cleanup
- [Choose An Interface](interfaces.md) - Web UI, CLI, TUI, rename, metadata, and ABS workflows
- [Configuration](CONFIGURATION.md) - Configuration file setup
- [Main README](../README.md) - Project overview
