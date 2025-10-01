# Audiobook Organizer
[![codecov](https://codecov.io/gh/jeeftor/audiobook-organizer/branch/main/graph/badge.svg)](https://codecov.io/gh/jeeftor/audiobook-organizer)
[![Coverage Status](https://coveralls.io/repos/github/jeeftor/audiobook-organizer/badge.svg?branch=main)](https://coveralls.io/github/jeeftor/audiobook-organizer?branch=main)

![docs/logo.png](docs/logo.png)

CLI tool to organize audiobooks based on **EITHER** `metadata.json` files **OR** embedded metadata in `.epub`, `.mp3`, and `.m4b` files.


## Features

- Organizes audiobooks by author/series/title structure
- Handles multiple authors
- Preserves spaces by default
- Optional space replacement with custom character
- Dry-run mode to preview changes
- Interactive prompt mode for reviewing moves
- Undo functionality
- Colored output
- Operation logs for recovery
- Separate input/output directory support
- **NEW**: Extract metadata directly from EPUB, MP3, and M4B files
- **NEW**: Process files in a flat directory structure
- **NEW**: Flexible directory layout options (author/series/title, author/title, author-only)
- **NEW**: Special handling for MP3 files with non-standard metadata structure

## Flat Mode vs. Non-Flat Mode

The organizer supports two modes of operation, which affect how input files are processed:

### Flat Mode (`--flat` flag)
- **Use Case**: When all your audiobook files are in a single directory, each potentially with different metadata.
- **Behavior**:
  - Each file is processed independently based on its embedded metadata.
  - Files are organized into `{author}/{title}/` structure based on their individual metadata.
  - **Important**: Files in the same directory will be grouped ONLY if they share identical metadata (author, title, series).
  - Ideal for:
    - Collections of single-file audiobooks (e.g., `.epub`, `.m4b` files)
    - Mixed collections where multiple books might be in the same directory

### Non-Flat Mode (Default)
- **Use Case**: When your files are already organized in a directory structure where each directory represents a single book.
- **Behavior**:
  - Assumes all files in the same directory belong to the same book.
  - Processes all files in a directory as a single unit with shared metadata.
  - Ideal for:
    - Multi-file audiobooks (e.g., split `.mp3` files)
    - Pre-organized collections where files are already grouped by book

#### Examples

**Flat Mode with Mixed Content:**
```
Input:
  /books/
    book1_chapter1.mp3   # Metadata: Author A, Book 1
    book1_chapter2.mp3   # Metadata: Author A, Book 1
    book2_chapter1.mp3   # Metadata: Author B, Book 2
    book3.epub           # Metadata: Author C, Book 3

Output (with --layout=author-title):
  /books/
    Author A/
      Book 1/
        book1_chapter1.mp3
        book1_chapter2.mp3
    Author B/
      Book 2/
        book2_chapter1.mp3
    Author C/
      Book 3/
        book3.epub
```

**Non-Flat Mode:**
```
Input:
  /books/
    Author A - Book 1/
      chapter1.mp3
      chapter2.mp3
    Author B - Book 2/
      part1.m4b
      part2.m4b

Output (with --layout=author-title):
  /books/
    Author A/
      Book 1/
        chapter1.mp3
        chapter2.mp3
    Author B/
      Book 2/
        part1.m4b
        part2.m4b
```

#### Important Notes:
- In flat mode, files are ONLY grouped if they have identical metadata. Two MP3s from different books in the same directory will be treated as separate books.
- For multi-file books in flat mode, ensure all files have consistent metadata to be grouped correctly.
- Non-flat mode is recommended when you have pre-organized collections or multi-file books.
- The `--flat` flag only affects how files are discovered, not the output structure. The output structure is always determined by the `--layout` flag.

## Pre-requirements

In order for this tool to operate you need to configure audiobookshelf to store `metadata.json` files in the same directories as your books. When this setting is toggled whenver metadata is generated a copy will be stored inside the directory - this is what will be used to rename the books.

![Settings - metadata.json](docs/store_metadata.jpg)

## Post-requirements

Because this software is not modifying the internal databse (due to time constraints) upon running the software you *may* end up with a good nubmer of **Missing** books as audiobookshelf. I believe the setting **Enable folder watcher for library** in your library config may inhibit this from happening - but - if it does occur you will see an error like this:

![issue](docs/issues.jpg)

To resolve these issues simply click on the **Issues** button

![remove button](docs/remove_books.jpg)

next use the **Remove All x Books** button to clean up the errors.



## Installation

*There are various ways to install this - I actually haven't tested the Docker install - but it should work :)*

### Ubuntu/Debian

```bash
# Install required dependencies
sudo apt-get install -y jq curl wget

# Download and install the latest release
LATEST_RELEASE=$(curl -s https://api.github.com/repos/jeeftor/audiobook-organizer/releases/latest | jq -r '(.assets[].browser_download_url | select(. | contains("_amd64.deb")))')
wget "$LATEST_RELEASE" -O audiobook-organizer.deb
sudo dpkg -i audiobook-organizer.deb

# Install any missing dependencies
sudo apt-get install -f -y

# Clean up
rm audiobook-organizer.deb
```

### Redhat

```bash
curl -s https://api.github.com/repos/jeeftor/audiobook-organizer/releases/latest | \
grep "browser_download_url.*rpm" | \
cut -d : -f 2,3 | \
tr -d \" | \
wget -qi -
```


### Alpine

```bash
# Download the latest .apk package
curl -s https://api.github.com/repos/jeeftor/audiobook-organizer/releases/latest | \
  grep "browser_download_url.*apk" | \
  cut -d : -f 2,3 | \
  tr -d \" | \
  wget -qi -

# Install the package
sudo apk add --allow-untrusted ./audiobook-organizer_*.apk

# Clean up
rm audiobook-organizer_*.apk
```

# Install the package
sudo rpm -i audiobook-organizer_*.rpm

# Clean up
rm audiobook-organizer_*.rpm

### Homebrew (macOS/Linux)

```bash
# Add the tap repository
brew tap jeeftor/tap

# Install the application
brew install audiobook-organizer
```

### Go Install

```bash
go install github.com/jeeftor/audiobook-organizer@latest
```

### Docker

```bash
docker pull jeffsui/audiobook-organizer:latest
```

## Usage

Basic organization:

```bash

# Organize in place
audiobook-organizer --dir=/path/to/audiobooks

# Organize to separate output directory
audiobook-organizer --dir=/path/to/source/audiobooks --out=/path/to/organized/audiobooks
```

Options:

- `--dir` / `--input`: Base directory to scan (required)
- `--out` / `--output`: Output directory for organized files (optional, defaults to --dir if not specified)
- `--config`: Config file (default is $HOME/.audiobook-organizer.yaml)
- `--dry-run`: Preview changes without moving files
- `--verbose`: Show detailed progress
- `--undo`: Restore files to original locations
- `--prompt`: Review and confirm each book move interactively
- `--remove-empty`: Remove empty directories after moving files and during initial scan
- `--replace_space`: Character to replace spaces (optional)
- `--use-embedded-metadata`: Use metadata embedded in EPUB, MP3, and M4B files if metadata.json is not found
- `--flat`: Process files in a flat directory structure (automatically enables --use-embedded-metadata)
- `--layout`: Directory structure layout (options: author-series-title, author-series-title-number, author-title, author-only)
- `--author-fields`: Comma-separated list of fields to try for author (e.g., 'authors,narrators,album_artist,artist')
- `--series-field`: Field to use as series (e.g., 'series', 'album')
- `--title-field`: Field to use as title (e.g., 'album', 'title', 'track_title')
- `--track-field`: Field to use for track number (e.g., 'track', 'track_number')
### Directory Layout Options

#### `--layout` Flag

Controls the directory structure of the organized audiobooks. Available options:

- `author-series-title` (default): Organizes as `Author/Series/Book Title/`
- `author-series-title-number`: Organizes as `Author/Series/#1 - Book Title/` (includes series number in title directory)
- `author-title`: Organizes as `Author/Book Title/`
- `author-only`: Organizes as `Author/` with all files directly in the author directory

#### `--use-series-as-title` Flag

When enabled, this flag modifies the directory structure to use the Series field as the main title directory. This is particularly useful for MP3 files where the Series field contains the actual book title.

**Behavior:**
- When `--use-series-as-title` is set to `true` and a Series is present in the metadata, the Series field will be used as the title in the directory structure.
- If no Series is present, the flag is ignored and the Title field is used.
- This flag works in conjunction with the `--layout` flag, modifying its behavior.

#### Flag Interaction Examples

1. **Default Behavior** (`--layout=author-series-title --use-series-as-title=false`):
   - With series: `Author/Series/Book Title/`
   - Without series: `Author/Book Title/`

2. **Using Series as Title** (`--layout=author-series-title --use-series-as-title=true`):
   - With series: `Author/Book Title/` (uses Series as the title)
   - Without series: `Author/Book Title/` (falls back to Title field)

3. **Author-Title Layout** (`--layout=author-title --use-series-as-title=true`):
   - With series: `Author/Book Title/` (uses Series as the title)
   - Without series: `Author/Book Title/` (uses Title field)

4. **Flat Mode** (`--flat --use-series-as-title=true`):
   - With series: `Author - Book Title.mp3` (uses Series as the title)
   - Without series: `Author - Book Title.mp3` (uses Title field)

### Docker Usage Examples

Basic usage with single directory:

```bash
# Process current directory
docker run -v $(pwd):/books \
  jeffsui/audiobook-organizer --dir=/books

# Process specific directory
docker run -v /path/to/audiobooks:/books \
  jeffsui/audiobook-organizer --dir=/books
```

Separate input and output directories:

```bash
# Mount source and destination directories
docker run \
  -v /path/to/source:/input:ro \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output

# Use current directory as source, output to specific directory
docker run \
  -v $(pwd):/input:ro \
  -v /path/to/organized:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output
```

Interactive mode with input/output:
```bash
# Interactive prompt mode with separate directories
docker run -it \
  -v /path/to/source:/input:ro \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output --prompt
```

Dry run with verbose output:

```bash
# Preview changes without moving files
docker run \
  -v /path/to/source:/input:ro \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output --dry-run --verbose
```

### Docker Volume Mounting Notes

- Use `:ro` for read-only access to source directories
- The container paths must match the `--dir` and `--out` parameters
- Use `-it` flag when running with `--prompt` for interactive mode
- Multiple directories can be mounted for source/destination separation
- Source and destination can be the same directory if desired
- Log files will be written to the output directory

### Interactive Mode

Using the `--prompt` flag will show each book's details and proposed move location:

```
Book found:
  Title: The Book Title
  Authors: Author One, Author Two
  Series: Amazing Series #1

Proposed move:
  From: /input/original/path/book
  To: /output/Author One,Author Two/Amazing Series #1/The Book Title

Proceed with move? [y/N]
```

## Metadata Sources

The tool can obtain metadata from two sources:

### 1. metadata.json Files

The tool primarily looks for `metadata.json` files in the same directory as your audiobook files. These files should have the following structure:

```json
{
  "authors": ["Author Name"],
  "title": "Book Title",
  "series": ["Series Name #1"]
}
```

### 2. Embedded EPUB, MP3, and M4B Metadata

When using the `--use-embedded-metadata` flag (which is automatically enabled with `--flat`), the tool can extract metadata directly from EPUB, MP3, and M4B files. This is useful when:

- No metadata.json file exists
- Processing a flat directory of EPUB, MP3, or M4B files
- Working with EPUBs, MP3s, or M4Bs that contain their own metadata

The tool will extract author, title, and series information from the EPUB's, MP3's, or M4B's internal metadata structure.

## Directory Structure

### Standard Layout (--layout=author-series-title)

Without series:

```
/output/Author Name/Book Title/
```

With series:

```
/output/Author Name/Series Name #1/Book Title/
```

### Author-Title Layout (--layout=author-title)

```
/output/Author Name/Book Title/
```

### Author-Only Layout (--layout=author-only)

```
/output/Author Name/
```

### Using Series as Title (--use-series-as-title)

For MP3 files where the Series field contains the actual book title and Title contains chapter info:

```
/output/Author Name/Series Name/
```

### Multiple Authors

```
/output/Author One,Author Two/Book Title/
```

### Space Replacement (--replace_space=".")

```
/output/Author.Name/Series.Name.#1/Book.Title/
```

## Recovery

Operations are logged to `.abook-org.log` in the output directory. Use `--undo` to restore files to their original locations:

```bash
# Undo with same input/output configuration
docker run \
  -v /path/to/source:/input \
  -v /path/to/destination:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output --undo
```

<!--
## FileFlows Docker Mod

If you want to include this in FileFlows you can add the following docker-mod script:

```bash
#!/bin/bash

# Function to handle errors
function handle_error {
    echo "An error occurred. Exiting..."
    exit 1
}

# Check if the --uninstall option is provided
if [ "$1" == "--uninstall" ]; then
    echo "Uninstalling audiobook-organizer..."
    if apt-get remove -y audiobook-organizer; then
        # Clean up repository files
        rm -f /usr/local/share/keyrings/audiobook-organizer.gpg
        rm -f /etc/apt/sources.list.d/audiobook-organizer.list
        apt-get update
        echo "audiobook-organizer successfully uninstalled."
        exit 0
    else
        handle_error
    fi
fi

# Check if audiobook-organizer is already installed
if command -v audiobook-organizer &>/dev/null; then
    echo "audiobook-organizer is already installed."
    exit 0
fi

echo "audiobook-organizer is not installed. Installing..."

# Install required dependencies
apt-get update
apt-get install -y curl gpg

# Create keyrings directory if it doesn't exist
mkdir -p /usr/local/share/keyrings

# Add the repository GPG key
if ! curl -fsSL https://github.com/jeeftor/audiobook-organizer/raw/main/key.gpg | gpg --dearmor -o /usr/local/share/keyrings/audiobook-organizer.gpg; then
    handle_error
fi

# Add repository
if ! echo "deb [signed-by=/usr/local/share/keyrings/audiobook-organizer.gpg] https://github.com/yourusername/audiobook-organizer/releases/latest/download/ /" > /etc/apt/sources.list.d/audiobook-organizer.list; then
    handle_error
fi

# Update package lists and install audiobook-organizer
if ! apt-get update || ! apt-get install -y audiobook-organizer; then
    handle_error
fi

# Verify installation
if command -v audiobook-organizer &>/dev/null; then
    echo "audiobook-organizer successfully installed."
    exit 0
fi

echo "Failed to install audiobook-organizer."
exit 1
```-->





## Configuration

The audiobook organizer supports multiple ways to configure its behavior:

### Configuration File

You can create a YAML configuration file in either:
- Your home directory: `~/.audiobook-organizer.yaml`
- The current directory: `.audiobook-organizer.yaml`
- Or specify a custom location: `--config /path/to/config.yaml`

Example configuration file:

```yaml
# Input directory (use either dir/input)
dir: "/path/to/audiobooks"
# or
input: "/path/to/audiobooks"

# Output directory (use either out/output)
out: "/path/to/organized/audiobooks"
# or
output: "/path/to/organized/audiobooks"

replace_space: "_"
verbose: true
dry-run: false
prompt: true
remove-empty: true  # Remove empty directories
use-embedded-metadata: true # Use metadata embedded in EPUB, MP3, and M4B files
flat: false  # Process files in a flat directory structure
layout: "author-series-title"  # Directory structure layout options: author-series-title, author-series-title-number, author-title, author-only
use-series-as-title: false  # Use Series field as the main title directory for MP3 files

# Metadata field mapping
author-fields: "authors,narrators,album_artist,artist"
series-field: "series"
title-field: "album,title,track_title"
track-field: "track,track_number"
```

### Environment Variables

All options can be set using environment variables with either the prefix `AO_` or `AUDIOBOOK_ORGANIZER_`:

```bash
# Input directory (use any)
export AO_DIR="/path/to/audiobooks"
export AO_INPUT="/path/to/audiobooks"
export AUDIOBOOK_ORGANIZER_DIR="/path/to/audiobooks"
export AUDIOBOOK_ORGANIZER_INPUT="/path/to/audiobooks"


# Output directory (use any)
export AO_OUT="/path/to/output"
export AO_OUTPUT="/path/to/output"
export AUDIOBOOK_ORGANIZER_OUT="/path/to/output"
export AUDIOBOOK_ORGANIZER_OUTPUT="/path/to/output"

# Other settings (use either prefix)
export AO_REPLACE_SPACE="_"
export AO_VERBOSE=true
export AO_REMOVE_EMPTY=true
export AO_USE_EMBEDDED_METADATA=true
export AO_LAYOUT="author-series-title"  # Options: author-series-title, author-series-title-number, author-title, author-only
export AO_USE_SERIES_AS_TITLE=false
export AO_AUTHOR_FIELDS="authors,narrators,album_artist,artist"
export AO_SERIES_FIELD="series"
export AO_TITLE_FIELD="album,title,track_title"
export AO_TRACK_FIELD="track,track_number"

# or
export AUDIOBOOK_ORGANIZER_REPLACE_SPACE="_"
export AUDIOBOOK_ORGANIZER_VERBOSE=true
export AUDIOBOOK_ORGANIZER_REMOVE_EMPTY=true
export AUDIOBOOK_ORGANIZER_USE_EMBEDDED_METADATA=true
export AUDIOBOOK_ORGANIZER_LAYOUT="author-series-title"
export AUDIOBOOK_ORGANIZER_USE_SERIES_AS_TITLE=false
export AUDIOBOOK_ORGANIZER_AUTHOR_FIELDS="authors,narrators,album_artist,artist"
export AUDIOBOOK_ORGANIZER_SERIES_FIELD="series"
export AUDIOBOOK_ORGANIZER_TITLE_FIELD="album,title,track_title"
export AUDIOBOOK_ORGANIZER_TRACK_FIELD="track,track_number"
```

### Command Line Flags

Command line flags take precedence over configuration files and environment variables. The input and output directories can be specified using either of their respective aliases:

```bash
# Using --dir and --out
audiobook-organizer \
  --dir=/path/to/audiobooks \
  --out=/path/to/output \
  --replace_space=_ \
  --verbose \
  --use-embedded-metadata

# Or using --input and --output
audiobook-organizer \
  --input=/path/to/audiobooks \
  --output=/path/to/output \
  --replace_space=_ \
  --verbose \
  --use-embedded-metadata
```

# Configuration Precedence

The configuration values are loaded in the following order (later sources override earlier ones):

1. Default values
2. Configuration file (`~/.audiobook-organizer.yaml` or specified with `--config`)
3. Environment variables
4. Command line flags

For the input and output directories, both aliases (`--dir`/`--input` and `--out`/`--output`) are treated equally, with the last specified value taking precedence.
