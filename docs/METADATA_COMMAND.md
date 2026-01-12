# Metadata Command

## Overview

The `metadata` command provides an interactive TUI for exploring audiobook metadata. It helps you understand what metadata is available in your files before organizing or renaming them.

## Usage

```bash
# Basic usage
audiobook-organizer metadata --dir=/path/to/audiobooks

# Short form
audiobook-organizer metadata -d /path/to/audiobooks

# Force embedded metadata
audiobook-organizer metadata --dir=/path --use-embedded-metadata

# Flat mode
audiobook-organizer metadata --dir=/path --flat

# Verbose output
audiobook-organizer metadata --dir=/path -v
```

## Features

The metadata command provides an interactive interface with multiple screens:

### 1. Scan Screen
- Automatically scans the directory for audiobook files
- Extracts metadata from files
- Shows progress during scanning

### 2. Template Builder
- Interactive template editor with live preview
- Shows first 10 files with their proposed names
- **?** - Toggle available fields help
- **m** - Toggle sample metadata display
- **Tab** - Cycle through author formats (First Last → Last, First → Preserve)
- **Enter** - Confirm template and continue to preview

### 3. Preview Screen
- Shows all proposed renames
- Scrollable list (↑↓ or j/k)
- Conflict detection and resolution
- **Enter/Y** - Proceed with renames
- **Q** - Go back to template editor

### 4. Process Screen
- Executes renames with progress
- Shows summary statistics
- Provides undo command

## Metadata Display (m)

When you press **m** in the template builder, you'll see sample metadata from the first file:

```
Sample Metadata (First File):
  Title: The Final Empire
  Authors: [Brandon Sanderson]
  Series: [Mistborn #1]
  Track: 1
  Album: Mistborn
  Year: 2006
  Narrator: Michael Kramer
  Source: audio
```

The **Source** field shows where the metadata came from:
- `json` - From metadata.json file
- `audio` - From embedded audio tags (MP3, M4B, M4A, etc.)
- `epub` - From EPUB file metadata

## GUI Metadata Viewer

The desktop GUI provides visual metadata exploration on the **Book List Screen**.

```bash
# Launch GUI
audiobook-organizer-gui
```

### Features

1. **Metadata Preview Panel** - Bottom section showing up to 3 sample audiobooks:
   - Complete metadata display with all extracted fields
   - **Color-coded field indicators**:
     - 🟢 Green = TITLE field
     - 🟠 Orange = AUTHOR field
     - 🔵 Cyan = SERIES field
     - 🔵 Blue = TRACK field
     - 🟣 Purple = DISC field
   - Source type display (metadata.json, MP3, M4B, EPUB)
   - Scrollable with `≪ < > ≫` navigation buttons
   - Font size: 10px for compact display

2. **Metadata Scanning Mode** - Top bar with three buttons:
   - **metadata.json** (default) - Audiobookshelf format
   - **embedded (directory)** - Extract from audio, group by directory
   - **embedded (file)** - Extract from audio, process individually
   - Current mode displayed: e.g., "Using metadata.json (3 books found)"

3. **Field Mapping Dialog** - Interactive configuration:
   - **Title Field** dropdown (title, album, series, etc.)
   - **Series Field** dropdown (series, album, title, etc.)
   - **Author Fields** multi-select with priority order
   - **Track Field** dropdown (track, track_number, etc.)
   - **Disc Field** dropdown (disc, disc_number, etc.)
   - **Live Preview** - Shows 3 sample audiobooks with color-coded indicators
   - **Refresh Button** - Reload preview with updated mapping

4. **Hybrid Metadata Display** - When metadata.json exists alongside audio files:
   - Shows fields from both sources
   - Visual indicators: 📁 for JSON fields, 🎵 for embedded fields
   - Demonstrates automatic merging of book-level and track-level metadata

**See also:** [GUI.md](GUI.md) for complete GUI documentation with screenshots

## Available Template Fields

Press **?** in the template builder to see available fields:

- `{author}` - First author (formatted)
- `{authors}` - All authors (comma-separated)
- `{title}` - Book title
- `{series}` - Series name (without number)
- `{series_number}` - Series number only
- `{track}` - Track number (zero-padded)
- `{album}` - Album field
- `{year}` - Publication year
- `{narrator}` - Narrator (if available)

## Metadata Source Selection

### Default Behavior
By default, the command prefers metadata.json if present, falling back to embedded metadata.

### Force Embedded Metadata
```bash
audiobook-organizer metadata --dir=/path --use-embedded-metadata
```
Ignores metadata.json files and always uses embedded tags from each file.

### Flat Mode
```bash
audiobook-organizer metadata --dir=/path --flat
```
Implies `--use-embedded-metadata`. Useful for flat directory structures.

## Keyboard Controls

### Template Builder
- **Enter** - Confirm template and continue
- **Tab** - Cycle author formats
- **?** - Toggle help (show available fields)
- **m** - Toggle metadata display (show sample file metadata)
- **Q/Esc** - Go back
- **Ctrl+C** - Quit

### Preview Screen
- **Enter/Y** - Proceed with renames
- **↑↓** or **j/k** - Scroll through files
- **PgUp/PgDn** - Page up/down
- **Q/Esc** - Go back to template editor
- **Ctrl+C** - Quit

### Process Screen
- **Q** - Exit after completion
- **Ctrl+C** - Quit

## Use Cases

### 1. Explore Available Metadata
Use the metadata command to see what fields are available in your files before deciding on an organization or rename strategy:

```bash
audiobook-organizer metadata --dir=/path/to/books
```

Press F2 to see actual metadata from your files.

### 2. Test Rename Templates
Build and test rename templates interactively:

```bash
audiobook-organizer metadata --dir=/path/to/books
```

Type different templates and see live previews of how files would be renamed.

### 3. Compare Metadata Sources
See the difference between metadata.json and embedded metadata:

```bash
# View metadata.json data
audiobook-organizer metadata --dir=/path/to/books

# View embedded metadata
audiobook-organizer metadata --dir=/path/to/books --use-embedded-metadata
```

Press F2 to see which source is being used.

### 4. Interactive Rename Workflow
The metadata command doubles as an interactive rename tool:

1. Scan files
2. Build template with live preview
3. Review all changes
4. Execute renames
5. Get undo command if needed

## Examples

### Example 1: Explore Metadata
```bash
audiobook-organizer metadata --dir=./audiobooks
```

1. Scans the directory
2. Shows template builder
3. Press **m** to see sample metadata
4. Press **?** to see available fields
5. Press Q to exit without making changes

### Example 2: Test Different Templates
```bash
audiobook-organizer metadata --dir=./audiobooks
```

Try different templates:
- `{author} - {title}`
- `{author} - {series} {series_number} - {title}`
- `{track} - {title}`

See live preview for each template.

### Example 3: Rename with Preview
```bash
audiobook-organizer metadata --dir=./audiobooks
```

1. Build your template
2. Press Enter to see full preview
3. Review all changes
4. Press Y to execute or Q to go back

## Comparison with Other Commands

### `metadata` vs `rename`
- **metadata**: Interactive TUI with live preview and metadata display
- **rename**: Non-interactive CLI for batch operations

### `metadata` vs `gui`
- **metadata**: Focused on metadata exploration and renaming
- **gui**: Full organization workflow (scan → organize → move files)

## Tips

1. **Always use m** to see actual metadata from your files
2. **Test templates** with live preview before committing
3. **Use --dry-run** is not needed - the preview screen shows all changes
4. **Check metadata source** (press **m** to see if using json/audio/epub)
5. **Use Tab** to try different author formats
6. **Scroll through preview** to check all files before proceeding

## Technical Details

The metadata command:
- Reuses the rename TUI infrastructure
- Supports all metadata sources (JSON, audio, EPUB)
- Respects `--use-embedded-metadata` and `--flat` flags
- Provides the same functionality as `rename` but with interactive UI
- Shows metadata source in metadata display (press **m**) for transparency
