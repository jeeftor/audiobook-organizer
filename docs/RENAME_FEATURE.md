# Rename Feature Documentation

## Overview

The rename feature provides both CLI and TUI interfaces for renaming audiobook files based on their metadata. It supports both `metadata.json` files and embedded metadata from audio files (MP3, M4B, M4A, OGG, FLAC) and EPUB files.

## Commands

### `rename` - CLI Command

Non-interactive command-line interface for batch renaming files.

```bash
# Basic usage
audiobook-organizer rename --dir=/path/to/audiobooks

# Custom template
audiobook-organizer rename --dir=/path/to/audiobooks \
  --template="{author} - {title}"

# Preview changes (dry-run)
audiobook-organizer rename --dir=/path/to/audiobooks --dry-run

# Use Last, First author format
audiobook-organizer rename --dir=/path/to/audiobooks \
  --author-format=last-first

# Interactive prompts
audiobook-organizer rename --dir=/path/to/audiobooks --prompt

# Undo previous renames
audiobook-organizer rename --dir=/path/to/audiobooks --undo
```

#### Flags

- `--template` - Filename template with placeholders (default: `{author} - {series} {series_number} - {title}`)
- `--author-format` - Author name format: `first-last`, `last-first`, `preserve` (default: `first-last`)
- `--recursive` - Recursively process subdirectories (default: `true`)
- `--preserve-path` - Only rename filename, preserve directory structure (default: `true`)
- `--strict` - Error on missing template fields
- `--prompt` - Prompt before renaming each file
- `--dry-run` - Preview changes without executing

### `rename-explore` - TUI Command

Interactive terminal UI for building and testing rename templates with live preview.

```bash
audiobook-organizer rename-explore --dir=/path/to/audiobooks
```

#### Features

1. **Scan Screen** - Automatically scans directory and extracts metadata
2. **Template Builder** - Interactive template editor with:
   - Live preview of first 10 files
   - Available fields help (F1)
   - Sample metadata display (F2)
   - Author format cycling (Tab)
3. **Preview Screen** - Review all proposed changes with scrolling
4. **Process Screen** - Execute renames with progress tracking

#### Keyboard Controls

**Template Builder:**
- `Enter` - Confirm template and continue
- `Tab` - Cycle through author formats (First Last → Last, First → Preserve)
- `F1` - Toggle available fields help
- `F2` - Toggle sample metadata display
- `Q/Esc` - Go back to previous screen

**Preview Screen:**
- `Enter/Y` - Proceed with renames
- `↑↓/j/k` - Scroll through file list
- `PgUp/PgDn` - Page up/down
- `Q/Esc` - Go back to template editor

**Process Screen:**
- `Q` - Exit after completion

### `GUI Rename` - Desktop Application

Visual desktop interface for exploring metadata and renaming files with real-time preview.

```bash
# Launch GUI
audiobook-organizer-gui
```

#### Features

1. **Directory Picker** - Native file dialogs for selecting audiobook directory
2. **Book List Screen** - Visual metadata preview with:
   - Live path preview with color-coded components
   - Three metadata scanning modes (metadata.json, embedded directory, embedded file)
   - Interactive field mapping dialog for custom metadata structures
3. **Preview Screen** - File renaming options:
   - **"Keep Original Names"** toggle - Preserves current filenames
   - **"Rename Files"** toggle - Enables custom templates
   - **Template Builder Dialog** - Visual interface for building filename templates:
     - 4 template slots for field assignment
     - Field assignment buttons (author, series, track, title)
     - Separator selection (-, /, space, ., _)
     - Live preview showing template format
   - Before/After preview with color-coded paths
   - Conflict detection with visual highlighting
4. **Execution** - Process files with progress tracking

#### Mouse Controls

- **Click** - Select directories, toggle options, assign template fields
- **Scroll** - Navigate long file lists in preview
- **Drag** - Resize windows (where supported)

**See also:** [GUI.md](GUI.md) for complete GUI documentation with screenshots

## Template System

### Available Fields

- `{author}` - First author from the authors list
- `{authors}` - All authors joined with commas
- `{title}` - Book title
- `{series}` - Series name (cleaned of numbers)
- `{series_number}` - Series number with zero-padding
- `{track}` - Track number with zero-padding
- `{album}` - Album name (from audio metadata)
- `{year}` - Publication year
- `{narrator}` - Narrator name (if available)

### Fallback Support

Templates support fallback values using `||`:

```
{series||album} - {title}
```

If series is empty, uses album instead.

### Examples

```bash
# Standard audiobook format
{author} - {series} {series_number} - {title}
# Output: Brandon Sanderson - Mistborn 01 - The Final Empire.m4b

# Simple format
{author} - {title}
# Output: Brandon Sanderson - The Final Empire.m4b

# Track-based format for multi-file books
{track} - {title}
# Output: 01 - Chapter One.mp3

# With fallback
{series||album} - {track} - {title}
# Output: Mistborn - 01 - The Final Empire.mp3
```

## Metadata Sources

### Default Behavior

By default, the rename command uses this priority:

1. **metadata.json** - If present in the same directory as the file
2. **Embedded metadata** - Extracted from the file itself (fallback)

### Forcing Metadata Source

You can override the default behavior:

```bash
# Force embedded metadata (ignore metadata.json)
audiobook-organizer rename --dir=/path --use-embedded-metadata

# Flat mode (implies embedded metadata)
audiobook-organizer rename --dir=/path --flat
```

**Use Cases:**

- `--use-embedded-metadata`: When you have metadata.json but want to use per-file embedded tags instead
- `--flat`: When working with flat directory structures where each file has its own metadata
- Default (no flags): When you have metadata.json files and want consistent naming per directory

### Supported File Types

- **Audio**: MP3, M4B, M4A, OGG, FLAC
- **EPUB**: .epub files
- **JSON**: metadata.json files

### Metadata Extraction

**From metadata.json:**
```json
{
  "title": "The Final Empire",
  "authors": ["Brandon Sanderson"],
  "series": ["Mistborn #1"],
  "track_number": 1
}
```

**From embedded audio tags:**
- Uses ID3 tags (MP3) or similar metadata
- Extracts: title, artist, album, track, year, narrator
- Series info from custom tags (TXXX:SERIES)

**From EPUB files:**
- Extracts: title, authors, series, publisher
- Supports Calibre series metadata
- Handles EPUB3 collection metadata

## Author Format Options

### first-last (default)
Converts "Last, First" to "First Last"
- Input: `Sanderson, Brandon`
- Output: `Brandon Sanderson`

### last-first
Converts "First Last" to "Last, First"
- Input: `Brandon Sanderson`
- Output: `Sanderson, Brandon`

### preserve
Keeps original format unchanged

## Conflict Resolution

When multiple files would have the same target name, the system automatically resolves conflicts:

```
book.m4b → Brandon Sanderson - Mistborn 01 - The Final Empire.m4b
book2.m4b → Brandon Sanderson - Mistborn 01 - The Final Empire (2).m4b
book3.m4b → Brandon Sanderson - Mistborn 01 - The Final Empire (3).m4b
```

## Undo Functionality

All rename operations are logged to `.abook-rename.log` in the target directory.

```bash
# Undo last rename operation
audiobook-organizer rename --dir=/path/to/audiobooks --undo
```

The undo operation:
1. Reads the log file
2. Reverses all renames in reverse order
3. Removes the log file on success

## Environment Variables

All flags can be set via environment variables:

```bash
# Short form
export AO_RENAME_TEMPLATE="{author} - {title}"
export AO_RENAME_AUTHOR_FORMAT="last-first"

# Long form
export AUDIOBOOK_ORGANIZER_RENAME_TEMPLATE="{author} - {title}"
export AUDIOBOOK_ORGANIZER_RENAME_AUTHOR_FORMAT="last-first"
```

## Implementation Details

### Core Components

1. **Template Parser** (`internal/organizer/template.go`)
   - Parses `{field}` placeholders
   - Supports fallback syntax `{field1||field2}`
   - Validates template syntax

2. **Author Formatter** (`internal/organizer/author_formatter.go`)
   - Converts between name formats
   - Handles "Last, First" ↔ "First Last"

3. **Renamer Engine** (`internal/organizer/renamer.go`)
   - Scans directories for files
   - Extracts metadata (JSON or embedded)
   - Generates new filenames
   - Detects and resolves conflicts
   - Executes renames with logging

4. **TUI Models** (`internal/tui/models/rename_*.go`)
   - Scan model - Directory scanning
   - Template model - Interactive template builder
   - Preview model - Change preview with scrolling
   - Process model - Rename execution

### Metadata Handling

The renamer automatically detects and uses the appropriate metadata source:

```go
// Check for metadata.json first
if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err == nil {
    provider = NewJSONMetadataProvider(metadataJsonPath)
} else {
    // Fall back to embedded metadata
    provider = NewMetadataProvider(filePath)
}
```

This ensures:
- Consistent metadata for all files in a directory (when using metadata.json)
- Per-file metadata for individual audio files (when using embedded tags)
- Seamless handling of mixed scenarios

## Testing

```bash
# Run rename tests
go test ./internal/organizer/renamer*.go -v

# Run template tests
go test ./internal/organizer/template*.go -v

# Run author formatter tests
go test ./internal/organizer/author_formatter*.go -v

# Test with sample data
./bin/audiobook-organizer rename-explore --dir=./testdata/m4b --dry-run
```

## Best Practices

1. **Always test with --dry-run first**
   ```bash
   audiobook-organizer rename --dir=/path --dry-run
   ```

2. **Use rename-explore for complex templates**
   - See live preview as you type
   - View actual metadata from your files
   - Test before committing

3. **Keep backups**
   - The undo feature helps, but backups are safer
   - Test on a copy first

4. **Use appropriate templates for your use case**
   - Single-file audiobooks: `{author} - {series} {series_number} - {title}`
   - Multi-file audiobooks: `{track} - {title}`
   - Simple collections: `{author} - {title}`

5. **Check metadata quality**
   - Use F2 in rename-explore to view metadata
   - Fix source metadata if needed
   - Use field mapping for non-standard fields
