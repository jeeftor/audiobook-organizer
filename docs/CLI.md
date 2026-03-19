# Audiobook Organizer CLI (Command Line Interface)

The **CLI (Command Line Interface)** provides direct, scriptable access to all audiobook organization features for automation and batch processing.

## Overview

The CLI is designed for:

- **Automation** - Scripts, cron jobs, scheduled tasks
- **Batch processing** - Process multiple directories
- **CI/CD integration** - Automated workflows
- **Docker containers** - Isolated execution environments
- **Power users** - Direct control without interactive prompts

**Best for:** Automation, scripts, batch processing, CI/CD pipelines, users who prefer direct command execution

**Not ideal for:** First-time users, exploring metadata, trial-and-error configuration (use GUI or TUI instead)

---

## Installation

```bash
# macOS/Linux via Homebrew
brew tap jeeftor/tap
brew install audiobook-organizer

# Debian/Ubuntu
sudo apt install audiobook-organizer

# RedHat/Fedora
sudo yum install audiobook-organizer

# Alpine
sudo apk add audiobook-organizer

# Go install
go install github.com/jeeftor/audiobook-organizer@latest

# Docker
docker pull jeffsui/audiobook-organizer:latest
```

**See also:** [INSTALLATION.md](INSTALLATION.md) for detailed platform-specific instructions

---

## Basic Usage

### Organize Audiobooks

```bash
# Organize in place (input = output)
audiobook-organizer --dir=/path/to/audiobooks

# Organize to separate output directory
audiobook-organizer --dir=/source --out=/organized

# Preview changes without moving files
audiobook-organizer --dir=/source --out=/organized --dry-run

# Verbose output
audiobook-organizer --dir=/source --out=/organized --verbose
```

### Rename Files

```bash
# Rename with default template
audiobook-organizer rename --dir=/path/to/audiobooks

# Rename with custom template
audiobook-organizer rename --dir=/path --template="{author} - {series} {series_number} - {title}"

# Preview renames
audiobook-organizer rename --dir=/path --dry-run

# Use Last, First author format
audiobook-organizer rename --dir=/path --author-format=last-first
```

### Undo Operations

```bash
# Undo previous organization (reads .abook-org.log)
audiobook-organizer --dir=/path --undo

# Undo previous rename
audiobook-organizer rename --dir=/path --undo
```

---

## Organization Commands

### Required Flags

| Flag | Aliases | Description |
|------|---------|-------------|
| `--dir` | `--input` | Input directory containing audiobooks (required) |

### Optional Flags

| Flag | Aliases | Default | Description |
|------|---------|---------|-------------|
| `--out` | `--output` | Same as `--dir` | Output directory for organized files |
| `--config` | - | `~/.audiobook-organizer.yaml` | Config file path |
| `--dry-run` | - | `false` | Preview changes without executing |
| `--verbose` | `-v` | `false` | Show detailed progress |
| `--prompt` | - | `false` | Review and confirm each book move |
| `--undo` | - | `false` | Restore files to original locations |
| `--remove-empty` | - | `false` | Remove empty directories |
| `--replace_space` | - | (none) | Character to replace spaces |
| `--use-embedded-metadata` | - | `false` | Extract metadata from audio files |
| `--flat` | - | `false` | Process files individually (auto-enables `--use-embedded-metadata`) |
| `--skip-errors` | - | `false` | Skip files with missing/invalid metadata instead of stopping |
| `--layout` | - | `author-series-title` | Directory structure pattern |
| `--author-fields` | - | `authors` | Comma-separated fields to try for author |
| `--series-field` | - | `series` | Field to use as series |
| `--title-field` | - | `title` | Field to use as title |
| `--track-field` | - | `track` | Field to use for track number |
| `--disc-field` | - | `disc` | Field to use for disc number (e.g., `disc`, `discnumber`, `disk`, `tpos`) |

### Layout Options

Seven directory structure patterns:

```bash
# Standard (default): Author/Series/Title/
--layout=author-series-title

# Numbered series: Author/Series/#1 - Title/
--layout=author-series-title-number

# Series-focused: Author/Series/ (multi-file books in series folder)
--layout=author-series

# No series: Author/Title/
--layout=author-title

# Flat per author: Author/
--layout=author-only

# Series without author: Series/Title/
--layout=series-title

# Series without author, numbered: Series/#1 - Title/
--layout=series-title-number
```

**See also:** [LAYOUTS.md](LAYOUTS.md) for detailed layout comparison

### Examples

**Basic organization:**
```bash
audiobook-organizer --dir=/media/audiobooks
```

**Organize to separate directory:**
```bash
audiobook-organizer \
  --dir=/media/unorganized \
  --out=/media/organized
```

**Preview changes first:**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --dry-run \
  --verbose
```

**Use embedded metadata (MP3/M4B files):**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --use-embedded-metadata
```

**Flat mode for single-file audiobooks:**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --flat
```

**Custom layout:**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --layout=author-title
```

**Replace spaces with underscores:**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --replace_space=_
```

**Interactive prompt mode:**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --prompt
```

---

## Rename Commands

### Required Flags

| Flag | Description |
|------|-------------|
| `--dir` | Directory containing audiobooks (required) |

### Optional Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--template` | `{author} - {series} {series_number} - {title}` | Filename template with placeholders |
| `--author-format` | `first-last` | Author name format: `first-last`, `last-first`, `preserve` |
| `--recursive` | `true` | Recursively process subdirectories |
| `--preserve-path` | `true` | Only rename filename, keep directory structure |
| `--prompt` | `false` | Prompt before renaming each file |
| `--strict` | `false` | Error on missing template fields |
| `--undo` | `false` | Undo previous rename operations |
| `--dry-run` | `false` | Preview renames without executing |
| `--author-fields` | `authors` | Comma-separated fields to try for author |
| `--title-field` | `title` | Field to use as title |
| `--series-field` | `series` | Field to use as series |
| `--track-field` | `track` | Field to use for track number |
| `--disc-field` | `disc` | Field to use for disc number (e.g., `disc`, `discnumber`, `tpos`) |

### Template Fields

Available placeholders for `--template`:

| Field | Description | Example |
|-------|-------------|---------|
| `{author}` | First author (formatted per `--author-format`) | `Brandon Sanderson` |
| `{authors}` | All authors (comma-separated) | `Brandon Sanderson, Dan Wells` |
| `{title}` | Book title | `The Final Empire` |
| `{series}` | Series name (without number) | `Mistborn` |
| `{series_number}` | Series number only | `1` |
| `{track}` | Track number (zero-padded) | `01` |
| `{album}` | Album field | `Mistborn Era 1` |
| `{year}` | Publication year | `2006` |
| `{narrator}` | Narrator (if available) | `Michael Kramer` |

### Examples

**Rename with custom template:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{author} - {track} - {title}"
```

**Use Last, First author format:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{author} - {title}" \
  --author-format=last-first
```

**Include series information:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{series} {series_number} - {track} - {title}"
```

**Preview renames first:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{author} - {title}" \
  --dry-run
```

**With field mapping:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{author} - {title}" \
  --author-fields="narrators,authors" \
  --title-field="album"
```

**Interactive prompt mode:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{author} - {title}" \
  --prompt
```

**Undo previous renames:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --undo
```

---

## Field Mapping

For files with non-standard metadata structures (especially MP3 files), use field mapping flags to specify which fields contain author, title, series, etc.

### Author Fields

Comma-separated list of fields to try in priority order:

```bash
--author-fields="authors,narrators,album_artist,artist"
```

**How it works:**
1. Try first field (`authors`)
2. If empty, try next field (`narrators`)
3. Continue until a non-empty field is found
4. Use first non-empty value

**Common configurations:**

```bash
# For Audiobookshelf metadata
--author-fields="authors"

# For standard MP3s
--author-fields="artist,album_artist"

# For audiobooks with narrator as artist
--author-fields="narrators,artist,album_artist,authors"
```

### Title Field

Single field to use for book title:

```bash
--title-field="album"  # Use album tag as title
--title-field="title"  # Use title tag (default)
```

**Common use cases:**
- MP3 files where album contains book title: `--title-field="album"`
- Standard metadata: `--title-field="title"`

### Series Field

Single field to use for series information:

```bash
--series-field="series"  # Use series tag (default)
--series-field="album"   # Use album tag as series
```

### Track Field

Single field to use for track numbers:

```bash
--track-field="track"         # Use track tag (default)
--track-field="track_number"  # Alternative field name
```

### Complete Field Mapping Example

```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --use-embedded-metadata \
  --author-fields="narrators,album_artist,artist" \
  --title-field="album" \
  --series-field="series" \
  --track-field="track" \
  --layout=author-series-title
```

**See also:** [METADATA.md](METADATA.md#field-mapping) for detailed field mapping guide

---

## Environment Variables

All flags can be set via environment variables with either `AO_` or `AUDIOBOOK_ORGANIZER_` prefix:

### Directory Flags

```bash
# Input directory (any of these work)
export AO_DIR="/path/to/audiobooks"
export AO_INPUT="/path/to/audiobooks"
export AUDIOBOOK_ORGANIZER_DIR="/path/to/audiobooks"
export AUDIOBOOK_ORGANIZER_INPUT="/path/to/audiobooks"

# Output directory (any of these work)
export AO_OUT="/path/to/output"
export AO_OUTPUT="/path/to/output"
export AUDIOBOOK_ORGANIZER_OUT="/path/to/output"
export AUDIOBOOK_ORGANIZER_OUTPUT="/path/to/output"
```

### Other Settings

```bash
# Short prefix (AO_)
export AO_REPLACE_SPACE="_"
export AO_VERBOSE=true
export AO_REMOVE_EMPTY=true
export AO_USE_EMBEDDED_METADATA=true
export AO_LAYOUT="author-series-title"
export AO_AUTHOR_FIELDS="authors,narrators,album_artist,artist"
export AO_SERIES_FIELD="series"
export AO_TITLE_FIELD="album,title"
export AO_TRACK_FIELD="track,track_number"

# Long prefix (AUDIOBOOK_ORGANIZER_)
export AUDIOBOOK_ORGANIZER_REPLACE_SPACE="_"
export AUDIOBOOK_ORGANIZER_VERBOSE=true
export AUDIOBOOK_ORGANIZER_REMOVE_EMPTY=true
export AUDIOBOOK_ORGANIZER_USE_EMBEDDED_METADATA=true
export AUDIOBOOK_ORGANIZER_LAYOUT="author-series-title"
export AUDIOBOOK_ORGANIZER_AUTHOR_FIELDS="authors,narrators,album_artist,artist"
```

**Precedence:** CLI flags > Environment variables > Config file > Defaults

**See also:** [CONFIGURATION.md](CONFIGURATION.md) for complete configuration guide

---

## Docker Usage

Run the CLI in a Docker container for isolated execution:

### Basic Usage

```bash
# Process current directory
docker run -v $(pwd):/books \
  jeffsui/audiobook-organizer --dir=/books

# Process specific directory
docker run -v /path/to/audiobooks:/books \
  jeffsui/audiobook-organizer --dir=/books
```

### Separate Input and Output

```bash
# Mount source (read-only) and destination directories
docker run \
  -v /path/to/source:/input:ro \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output
```

### Interactive Prompt Mode

```bash
# Use -it for interactive mode
docker run -it \
  -v /path/to/source:/input:ro \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output --prompt
```

### Dry Run

```bash
# Preview changes without moving files
docker run \
  -v /path/to/source:/input:ro \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output --dry-run --verbose
```

### With Configuration File

```bash
# Mount config file
docker run \
  -v $(pwd)/.audiobook-organizer.yaml:/config.yaml:ro \
  -v /path/to/books:/books \
  jeffsui/audiobook-organizer --dir=/books --config=/config.yaml
```

### Environment Variables

```bash
# Pass environment variables
docker run \
  -e AO_LAYOUT=author-title \
  -e AO_VERBOSE=true \
  -v /path/to/books:/books \
  jeffsui/audiobook-organizer --dir=/books
```

### Docker Compose

```yaml
version: '3.8'
services:
  audiobook-organizer:
    image: jeffsui/audiobook-organizer:latest
    volumes:
      - /path/to/source:/input:ro
      - /path/to/destination:/output
      - ./config.yaml:/config.yaml:ro
    environment:
      - AO_LAYOUT=author-series-title
      - AO_VERBOSE=true
    command: --dir=/input --out=/output --config=/config.yaml
```

### Volume Mounting Notes

- **`:ro` suffix** - Mount volume as read-only (recommended for source)
- **Container paths** - Must match `--dir` and `--out` parameters
- **Log files** - Written to output directory (`.abook-org.log`)
- **Multiple mounts** - Can mount multiple directories
- **Same directory** - Source and destination can be the same

---

## Scripting Examples

### Bash Script: Batch Processing

```bash
#!/bin/bash
# Process multiple audiobook directories

DIRS=(
  "/media/audiobooks/sci-fi"
  "/media/audiobooks/fantasy"
  "/media/audiobooks/non-fiction"
)

for dir in "${DIRS[@]}"; do
  echo "Processing: $dir"
  audiobook-organizer \
    --dir="$dir" \
    --layout=author-series-title \
    --verbose
done
```

### Cron Job: Scheduled Organization

```bash
# Add to crontab: crontab -e
# Run every night at 2 AM

0 2 * * * /usr/local/bin/audiobook-organizer \
  --dir=/media/audiobooks \
  --out=/media/organized \
  --remove-empty \
  >> /var/log/audiobook-organizer.log 2>&1
```

### CI/CD: GitHub Actions

```yaml
name: Organize Audiobooks

on:
  push:
    paths:
      - 'audiobooks/**'

jobs:
  organize:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install audiobook-organizer
        run: |
          wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Linux_x86_64.tar.gz
          tar xzf audiobook-organizer_Linux_x86_64.tar.gz
          chmod +x audiobook-organizer

      - name: Organize audiobooks
        run: |
          ./audiobook-organizer \
            --dir=./audiobooks \
            --out=./organized \
            --verbose

      - name: Upload organized files
        uses: actions/upload-artifact@v3
        with:
          name: organized-audiobooks
          path: organized/
```

### Python Script: API Wrapper

```python
#!/usr/bin/env python3
import subprocess
import json

def organize_audiobooks(input_dir, output_dir, layout="author-series-title", dry_run=False):
    """Organize audiobooks using CLI"""
    cmd = [
        "audiobook-organizer",
        f"--dir={input_dir}",
        f"--out={output_dir}",
        f"--layout={layout}",
        "--verbose"
    ]

    if dry_run:
        cmd.append("--dry-run")

    result = subprocess.run(cmd, capture_output=True, text=True)

    return {
        "success": result.returncode == 0,
        "stdout": result.stdout,
        "stderr": result.stderr
    }

# Usage
result = organize_audiobooks(
    input_dir="/media/audiobooks",
    output_dir="/media/organized",
    dry_run=True
)

print(json.dumps(result, indent=2))
```

### PowerShell Script: Windows Automation

```powershell
# Organize audiobooks on Windows
param(
    [string]$InputDir = "C:\Audiobooks",
    [string]$OutputDir = "C:\Organized",
    [string]$Layout = "author-series-title"
)

Write-Host "Organizing audiobooks from $InputDir to $OutputDir"

& audiobook-organizer `
    --dir=$InputDir `
    --out=$OutputDir `
    --layout=$Layout `
    --verbose

if ($LASTEXITCODE -eq 0) {
    Write-Host "Organization complete!" -ForegroundColor Green
} else {
    Write-Host "Organization failed!" -ForegroundColor Red
    exit $LASTEXITCODE
}
```

---

## Advanced Usage

### Flat vs Non-Flat Mode

**Flat Mode** (`--flat`):
- Processes each file independently based on its metadata
- Files grouped only if they have identical metadata
- Auto-enables `--use-embedded-metadata`
- Good for: Single-file audiobooks, mixed collections

```bash
audiobook-organizer --dir=/books --flat
```

**Skipping bad files** (`--skip-errors`):

Use with `--flat` when your collection has mixed metadata quality. Files with missing or invalid metadata (e.g., no author tag) are skipped with a warning instead of stopping the entire run. Files with good metadata are still organized.

```bash
audiobook-organizer \
  --dir=/books \
  --flat \
  --skip-errors \
  --dry-run
```

**Non-Flat Mode** (default):
- All files in a directory treated as one book
- Shared metadata per directory
- Good for: Multi-file audiobooks, pre-organized collections

```bash
audiobook-organizer --dir=/books
```

**See also:** [METADATA.md](METADATA.md#flat-vs-non-flat) for detailed comparison

### Metadata Source Selection

**metadata.json files** (default):
```bash
audiobook-organizer --dir=/books
```

**Embedded metadata** (EPUB, MP3, M4B):
```bash
audiobook-organizer --dir=/books --use-embedded-metadata
```

**Hybrid mode** (automatic):
- When metadata.json exists alongside audio files
- Book-level metadata from JSON
- Track-level metadata from audio files
- No special flag needed

**See also:** [METADATA.md](METADATA.md) for metadata extraction guide

### Undo Operations

All operations are logged to `.abook-org.log` in the output directory:

```bash
# Undo organization
audiobook-organizer --dir=/source --out=/dest --undo

# Undo rename
audiobook-organizer rename --dir=/books --undo
```

**Log file format:** JSON with source/target paths for each operation

**Limitations:**
- Only works if log file exists
- Can't undo if files were manually modified after organization
- Undo reverses operations in reverse order

---

## Troubleshooting

### No audiobooks found

**Solutions:**
- Verify directory contains supported formats (MP3, M4B, EPUB, metadata.json)
- Try `--use-embedded-metadata` if no metadata.json files
- Use `--flat` for single-file audiobooks
- Check file permissions

### Metadata extraction errors

**Solutions:**
- Use `--skip-errors` with `--flat` to skip files with bad metadata and organize the rest
- Use `--author-fields` to specify correct metadata fields
- Check file tags with metadata viewer: `audiobook-organizer metadata --dir=/path`
- Verify audio files aren't corrupted
- See [METADATA.md](METADATA.md) for field mapping guide

### Permission errors

**Solutions:**
- Ensure write access to output directory
- Run with appropriate user permissions
- Check directory ownership: `ls -la`
- On Windows, run as Administrator if needed

### Path length errors (Windows)

**Symptoms:** "path too long" errors on Windows

**Solutions:**
- Use shorter output paths
- Enable long path support: https://learn.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation
- Use `--layout=author-only` for shorter paths
- Use `--replace_space=_` to shorten paths

### Dry run shows unexpected results

**Solutions:**
- Review field mapping: use `audiobook-organizer metadata --dir=/path`
- Adjust `--layout` option
- Check `--author-fields`, `--title-field`, `--series-field` values
- Use `--verbose` for detailed logging

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | File system error |
| `4` | Metadata extraction error |

**Use in scripts:**
```bash
audiobook-organizer --dir=/books
if [ $? -eq 0 ]; then
    echo "Success"
else
    echo "Failed with exit code $?"
fi
```

---

## See Also

- [GUI.md](GUI.md) - Desktop GUI guide
- [TUI.md](TUI.md) - Terminal User Interface guide
- [METADATA.md](METADATA.md) - Metadata extraction guide
- [LAYOUTS.md](LAYOUTS.md) - Directory layout options
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file format
- [INSTALLATION.md](INSTALLATION.md) - Platform-specific installation
- [Main README](../README.md) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Feature requests:** [GitHub Discussions](https://github.com/jeeftor/audiobook-organizer/discussions)
- **Questions:** [GitHub Discussions Q&A](https://github.com/jeeftor/audiobook-organizer/discussions/categories/q-a)

When reporting CLI issues, please include:
- Operating system and version
- CLI version (`audiobook-organizer version`)
- Complete command with all flags
- Error output (use `--verbose`)
- Sample directory structure (if relevant)
