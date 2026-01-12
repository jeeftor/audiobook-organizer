# Configuration Guide

The Audiobook Organizer supports multiple configuration methods: config files, environment variables, and CLI flags. This guide covers all configuration options and their precedence.

## Configuration Precedence

Configuration values are loaded in this order (later sources override earlier ones):

```
Defaults → Config File → Environment Variables → CLI Flags
(lowest)                                          (highest)
```

**Example:**
```bash
# Config file sets: layout=author-title
# Environment sets: AO_LAYOUT=author-series-title
# CLI flag sets: --layout=author-only

# Result: author-only (CLI flag wins)
```

---

## Config File

### File Locations

The organizer searches for config files in this order:

1. **Custom path** (if specified): `--config /path/to/config.yaml`
2. **Current directory**: `./.audiobook-organizer.yaml`
3. **Home directory**: `~/.audiobook-organizer.yaml`

**First file found is used.** Subsequent files are ignored.

### File Format

Config files use YAML format:

```yaml
# Basic configuration
dir: "/path/to/audiobooks"
out: "/path/to/organized"

# Organization options
layout: "author-series-title"
replace_space: ""
remove-empty: true
use-embedded-metadata: false
flat: false
verbose: false
dry-run: false
prompt: false

# Field mapping
author-fields: "authors,narrators,album_artist,artist"
series-field: "series"
title-field: "album,title"
track-field: "track,track_number"

# Rename options (for rename command)
template: "{author} - {series} {series_number} - {title}"
author-format: "first-last"
recursive: true
preserve-path: true
strict: false
```

### Creating a Config File

**Recommended location:** Home directory

```bash
# Create config file
cat > ~/.audiobook-organizer.yaml <<'EOF'
# Audiobook Organizer Configuration

# Default directories
dir: "/media/audiobooks"
out: "/media/organized"

# Organization settings
layout: "author-series-title"
use-embedded-metadata: true
remove-empty: true
verbose: true

# Field mapping for MP3 files
author-fields: "narrators,authors,album_artist,artist"
title-field: "album"
EOF
```

---

## All Configuration Options

### Directory Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dir` / `input` | string | (required) | Input directory containing audiobooks |
| `out` / `output` | string | Same as `dir` | Output directory for organized files |

**Note:** Both `dir`/`input` and `out`/`output` are interchangeable aliases.

### Organization Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `layout` | string | `author-series-title` | Directory structure pattern (see [LAYOUTS.md](LAYOUTS.md)) |
| `replace_space` | string | (empty) | Character to replace spaces in paths |
| `remove-empty` | boolean | `false` | Remove empty directories after organizing |
| `use-embedded-metadata` | boolean | `false` | Extract metadata from audio file tags |
| `flat` | boolean | `false` | Process files individually (auto-enables `use-embedded-metadata`) |
| `dry-run` | boolean | `false` | Preview changes without executing |
| `verbose` | boolean | `false` | Show detailed progress output |
| `prompt` | boolean | `false` | Review and confirm each book move |
| `undo` | boolean | `false` | Restore files to original locations |

### Field Mapping Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `author-fields` | string | `authors` | Comma-separated fields to try for author (priority order) |
| `series-field` | string | `series` | Field to use for series |
| `title-field` | string | `title` | Field to use for title |
| `track-field` | string | `track` | Field to use for track number |

### Rename Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `template` | string | `{author} - {series} {series_number} - {title}` | Filename template with placeholders |
| `author-format` | string | `first-last` | Author name format: `first-last`, `last-first`, `preserve` |
| `recursive` | boolean | `true` | Recursively process subdirectories |
| `preserve-path` | boolean | `true` | Only rename filename, keep directory structure |
| `strict` | boolean | `false` | Error on missing template fields |

### Other Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `config` | string | `~/.audiobook-organizer.yaml` | Path to config file |

---

## Environment Variables

All options can be set via environment variables using either `AO_` or `AUDIOBOOK_ORGANIZER_` prefix.

### Prefix Options

Both prefixes work identically:
- **Short:** `AO_*` (e.g., `AO_VERBOSE`)
- **Long:** `AUDIOBOOK_ORGANIZER_*` (e.g., `AUDIOBOOK_ORGANIZER_VERBOSE`)

Choose based on your preference. Short prefix is more concise.

### Variable Names

Environment variable names are uppercase with underscores. Hyphens in config options become underscores.

**Config option** → **Environment variable**:
- `dir` → `AO_DIR` or `AUDIOBOOK_ORGANIZER_DIR`
- `use-embedded-metadata` → `AO_USE_EMBEDDED_METADATA`
- `remove-empty` → `AO_REMOVE_EMPTY`

### Complete Environment Variable Reference

#### Directory Variables

```bash
# Input directory (all are equivalent)
export AO_DIR="/path/to/audiobooks"
export AO_INPUT="/path/to/audiobooks"
export AUDIOBOOK_ORGANIZER_DIR="/path/to/audiobooks"
export AUDIOBOOK_ORGANIZER_INPUT="/path/to/audiobooks"

# Output directory (all are equivalent)
export AO_OUT="/path/to/output"
export AO_OUTPUT="/path/to/output"
export AUDIOBOOK_ORGANIZER_OUT="/path/to/output"
export AUDIOBOOK_ORGANIZER_OUTPUT="/path/to/output"
```

#### Organization Variables (Short Prefix)

```bash
export AO_LAYOUT="author-series-title"
export AO_REPLACE_SPACE="_"
export AO_REMOVE_EMPTY=true
export AO_USE_EMBEDDED_METADATA=true
export AO_FLAT=false
export AO_DRY_RUN=false
export AO_VERBOSE=true
export AO_PROMPT=false
```

#### Field Mapping Variables (Short Prefix)

```bash
export AO_AUTHOR_FIELDS="authors,narrators,album_artist,artist"
export AO_SERIES_FIELD="series"
export AO_TITLE_FIELD="album,title"
export AO_TRACK_FIELD="track,track_number"
```

#### Rename Variables (Short Prefix)

```bash
export AO_TEMPLATE="{author} - {series} {series_number} - {title}"
export AO_AUTHOR_FORMAT="first-last"
export AO_RECURSIVE=true
export AO_PRESERVE_PATH=true
export AO_STRICT=false
```

#### Organization Variables (Long Prefix)

```bash
export AUDIOBOOK_ORGANIZER_LAYOUT="author-series-title"
export AUDIOBOOK_ORGANIZER_REPLACE_SPACE="_"
export AUDIOBOOK_ORGANIZER_REMOVE_EMPTY=true
export AUDIOBOOK_ORGANIZER_USE_EMBEDDED_METADATA=true
export AUDIOBOOK_ORGANIZER_FLAT=false
export AUDIOBOOK_ORGANIZER_DRY_RUN=false
export AUDIOBOOK_ORGANIZER_VERBOSE=true
export AUDIOBOOK_ORGANIZER_PROMPT=false
```

### Boolean Values

Environment variables accept multiple boolean formats:
- **True:** `true`, `True`, `TRUE`, `1`, `yes`, `Yes`, `YES`
- **False:** `false`, `False`, `FALSE`, `0`, `no`, `No`, `NO`, (empty string)

```bash
export AO_VERBOSE=true    # ✓ Works
export AO_VERBOSE=TRUE    # ✓ Works
export AO_VERBOSE=1       # ✓ Works
export AO_VERBOSE=yes     # ✓ Works
```

---

## Example Configurations

### Scenario 1: Default Setup (Audiobookshelf)

**Use case:** Organize audiobooks with metadata.json files from Audiobookshelf

```yaml
# ~/.audiobook-organizer.yaml
dir: "/media/audiobooks"
out: "/media/organized"
layout: "author-series-title"
remove-empty: true
verbose: true
```

**Environment variables:**
```bash
export AO_DIR="/media/audiobooks"
export AO_OUT="/media/organized"
export AO_LAYOUT="author-series-title"
export AO_REMOVE_EMPTY=true
export AO_VERBOSE=true
```

**CLI:**
```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --out=/media/organized \
  --layout=author-series-title \
  --remove-empty \
  --verbose
```

### Scenario 2: Flat Mode for EPUBs

**Use case:** Organize single-file EPUB audiobooks

```yaml
# .audiobook-organizer.yaml
dir: "/downloads/epubs"
out: "/library/epubs"
layout: "author-title"
flat: true  # Auto-enables use-embedded-metadata
verbose: true
```

**CLI:**
```bash
audiobook-organizer \
  --dir=/downloads/epubs \
  --out=/library/epubs \
  --layout=author-title \
  --flat
```

### Scenario 3: MP3 with Custom Field Mapping

**Use case:** Organize MP3 audiobooks where "album" contains book title and "artist" contains author

```yaml
# ~/.audiobook-organizer.yaml
dir: "/media/mp3-audiobooks"
out: "/media/organized"
use-embedded-metadata: true
layout: "author-series-title"

# Field mapping
author-fields: "artist,album_artist,authors"
title-field: "album"
series-field: "series"
track-field: "track"

# Options
remove-empty: true
verbose: true
```

**CLI equivalent:**
```bash
audiobook-organizer \
  --dir=/media/mp3-audiobooks \
  --out=/media/organized \
  --use-embedded-metadata \
  --layout=author-series-title \
  --author-fields="artist,album_artist,authors" \
  --title-field="album" \
  --remove-empty \
  --verbose
```

### Scenario 4: Docker Environment

**Use case:** Run in Docker with environment variables

```bash
# .env file
AO_LAYOUT=author-title
AO_VERBOSE=true
AO_REMOVE_EMPTY=true
AO_USE_EMBEDDED_METADATA=true
```

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  organizer:
    image: jeffsui/audiobook-organizer:latest
    volumes:
      - /media/source:/input:ro
      - /media/output:/output
    env_file:
      - .env
    command: --dir=/input --out=/output
```

### Scenario 5: CI/CD Automation

**Use case:** Automated organization in GitHub Actions

```yaml
# .github/workflows/organize.yml
name: Organize Audiobooks
on:
  push:
    paths:
      - 'audiobooks/**'
jobs:
  organize:
    runs-on: ubuntu-latest
    env:
      AO_LAYOUT: author-series-title
      AO_VERBOSE: true
      AO_DRY_RUN: false
    steps:
      - uses: actions/checkout@v3
      - name: Install organizer
        run: |
          wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Linux_x86_64.tar.gz
          tar xzf audiobook-organizer_Linux_x86_64.tar.gz
      - name: Organize
        run: ./audiobook-organizer --dir=./audiobooks --out=./organized
```

### Scenario 6: Rename with Template

**Use case:** Rename files using custom template

```yaml
# ~/.audiobook-organizer.yaml
# For rename command
dir: "/media/audiobooks"
template: "{author} - {series} {series_number} - {track} - {title}"
author-format: "last-first"
author-fields: "narrators,authors"
title-field: "title"
verbose: true
dry-run: false  # Set to true for testing
```

**CLI:**
```bash
audiobook-organizer rename \
  --dir=/media/audiobooks \
  --template="{author} - {series} {series_number} - {track} - {title}" \
  --author-format=last-first \
  --author-fields="narrators,authors"
```

---

## Configuration Tips

### Start with Defaults

Begin with minimal config and add options as needed:

```yaml
# Minimal config
dir: "/media/audiobooks"
verbose: true
```

Run with `--dry-run` to preview results, then add more options:

```yaml
# After testing
dir: "/media/audiobooks"
out: "/media/organized"
layout: "author-title"
remove-empty: true
verbose: true
```

### Use Config Files for Complex Setups

For complicated field mappings or multiple options, config files are easier to manage than CLI flags:

```yaml
# Complex MP3 field mapping
author-fields: "narrators,authors,album_artist,artist,composer"
title-field: "album,title,work"
series-field: "series,album,grouping"
track-field: "track,track_number,trck"
```

### Environment Variables for Scripting

Use environment variables in scripts to avoid repetition:

```bash
#!/bin/bash
export AO_LAYOUT="author-series-title"
export AO_VERBOSE=true
export AO_REMOVE_EMPTY=true

audiobook-organizer --dir=/media/sci-fi --out=/library/sci-fi
audiobook-organizer --dir=/media/fantasy --out=/library/fantasy
audiobook-organizer --dir=/media/mystery --out=/library/mystery
```

### CLI Flags for One-Time Changes

Override config/environment for specific runs:

```bash
# Usually use config file settings, but this time:
audiobook-organizer \
  --dir=/media/audiobooks \
  --layout=author-only  # Override config file
```

### Separate Configs for Different Libraries

Use different config files for different audiobook collections:

```bash
# Config for Audiobookshelf collection
audiobook-organizer --config=~/configs/audiobookshelf.yaml

# Config for MP3 collection
audiobook-organizer --config=~/configs/mp3-audiobooks.yaml
```

---

## Troubleshooting Configuration

### Config file not found

**Check file location:**
```bash
# List possible config file locations
ls -la ~/.audiobook-organizer.yaml
ls -la ./.audiobook-organizer.yaml
```

**Verify YAML syntax:**
```bash
# Test YAML syntax (if you have python)
python -c "import yaml; yaml.safe_load(open('~/.audiobook-organizer.yaml'))"
```

### Environment variables not working

**Check variable names:**
```bash
# List all AO_* variables
env | grep AO_

# List all AUDIOBOOK_ORGANIZER_* variables
env | grep AUDIOBOOK_ORGANIZER_
```

**Verify values:**
```bash
echo $AO_DIR
echo $AO_LAYOUT
```

### CLI flags not overriding config

**Verify precedence:**
```bash
# Use verbose to see what's being used
audiobook-organizer --dir=/path --verbose
```

Look for log output showing which config source is being used.

### Unknown configuration options

**Symptoms:** Config file or environment variable seems to be ignored

**Solution:** Verify option names match exactly (case-sensitive for config file, uppercase for environment):

```yaml
# ✗ Wrong
Layout: "author-title"  # Capitalized
use_embedded_metadata: true  # Underscores instead of hyphens

# ✓ Correct
layout: "author-title"
use-embedded-metadata: true
```

---

## See Also

- [CLI.md](CLI.md) - Command-line reference with all flags
- [METADATA.md](METADATA.md) - Field mapping deep dive
- [LAYOUTS.md](LAYOUTS.md) - Directory layout options
- [Main README](../README.md) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Questions:** [GitHub Discussions Q&A](https://github.com/jeeftor/audiobook-organizer/discussions/categories/q-a)
