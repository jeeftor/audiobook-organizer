# Audiobook Organizer TUI (Terminal User Interface)

The **TUI (Text User Interface)** provides an interactive, keyboard-driven workflow for organizing and renaming audiobooks directly in your terminal.

## Overview

The TUI offers a middle ground between the visual GUI and the scriptable CLI:

- **Interactive keyboard navigation** - Visual feedback without requiring a mouse
- **Step-by-step workflow** - Guided screens like the GUI
- **Real-time filtering** - Type to filter directories and options
- **Visual progress** - Color-coded output and progress tracking
- **SSH-friendly** - Works over remote connections
- **No X11 required** - Pure terminal-based interface

**Best for:** Power users who prefer keyboard navigation, SSH/remote sessions, interactive workflow without GUI overhead

**Not ideal for:** Automation, batch processing, users unfamiliar with terminals (use GUI or CLI instead)

---

## Installation

The TUI is included with the main `audiobook-organizer` CLI binary:

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
```

**See also:** [INSTALLATION.md](INSTALLATION.md) for platform-specific instructions

---

## Organization TUI

The **Organization TUI** guides you through organizing audiobooks into structured directories.

### Launch

```bash
# Interactive directory picker
audiobook-organizer tui

# Pre-selected directories (skip picker)
audiobook-organizer tui --dir=/path/to/audiobooks --out=/path/to/organized

# With specific layout
audiobook-organizer tui --layout=author-title

# Verbose output
audiobook-organizer tui --verbose
```

### Workflow Screens

The organization TUI follows a 6-screen workflow:

#### 1. Output Path Configuration (Optional)

**Purpose:** Select or configure output directory if not provided via `--out` flag

**Navigation:**
- `Tab` - Switch between input field and buttons
- `Enter` - Confirm selection
- `Ctrl+Q` - Quit

**Tips:**
- Can be same as input directory (organize in-place)
- Leave empty to use input directory

#### 2. Directory Picker

**Purpose:** Browse and select input directory containing audiobooks

**Features:**
- **Hierarchical directory tree** - Visual directory structure
- **Real-time filtering** - Type to filter visible directories
- **Current directory indicator** - Highlighted path
- **Breadcrumb trail** - Shows current location

**Keyboard shortcuts:**
- `↑/↓` or `j/k` - Navigate directories
- `Enter` - Open/navigate into directory
- `Ctrl+S` - Select current directory
- **Type to filter** - Start typing to filter directories
- `ESC` - Clear filter
- `Ctrl+B` - Go up one level (parent directory)
- `Ctrl+H` - Jump to home directory
- `Ctrl+R` - Jump to root directory
- `Ctrl+Q` - Quit

**Tips:**
- Use filter to quickly find deeply nested directories
- Press `ESC` to clear filter and see all directories

#### 3. Scan Screen

**Purpose:** Scan selected directory for audiobooks

**Display:**
- Progress spinner
- Files/directories scanned count
- Audiobooks found count

**What happens:**
- Searches for `metadata.json` files (if using metadata.json mode)
- Extracts metadata from EPUB, MP3, M4B files (if using embedded mode)
- Groups files into books/albums

**Duration:** Depends on directory size and file count

#### 4. Book List Screen

**Purpose:** Review discovered audiobooks and select which to organize

**Display:**
- List of all discovered audiobooks
- Metadata for each book (title, author, series)
- Selection checkboxes
- Summary counts

**Navigation:**
- `↑/↓` - Navigate list
- `Space` - Toggle selection
- `a` - Select all
- `n` - Deselect all
- `Enter` - Continue to settings (with selected books)
- `q` - Back to directory picker

**Tips:**
- Review metadata before proceeding
- Use `Space` to deselect books with incorrect metadata

#### 5. Settings Screen

**Purpose:** Configure organization options

**Available settings:**
- **Layout** - Directory structure pattern
  - `author-series-title` (default)
  - `author-series-title-number`
  - `author-title`
  - `author-only`
  - `series-title`
  - `series-title-number`
- **Use Embedded Metadata** - Extract from audio files instead of metadata.json
- **Flat Mode** - Process files individually vs. by directory
- **Remove Empty Directories** - Clean up after moving
- **Replace Spaces** - Character to replace spaces (or leave blank)
- **Verbose Output** - Detailed logging

**Navigation:**
- `↑/↓` - Navigate settings
- `Enter` / `Space` - Toggle or edit setting
- `Tab` - Move between fields
- `Enter on "Continue"` - Proceed to preview
- `q` - Back to book list

**Tips:**
- Change layout to match your preferred organization style
- Enable verbose for troubleshooting

#### 6. Preview Screen

**Purpose:** Review all proposed file operations before executing

**Display:**
- **From** → **To** paths for each operation
- **Color-coded paths** (if terminal supports):
  - Author in orange
  - Series in cyan
  - Title in green
- **Summary** - Total operations, conflicts detected
- **Action choice** - Execute or cancel

**Navigation:**
- `Scroll` - Review all operations
- `Enter on "Execute"` - Perform organization
- `Enter on "Cancel"` - Return to settings
- `q` - Back to settings

**Tips:**
- Review carefully before executing
- Look for conflicts or incorrect paths
- Use `q` to go back and adjust settings

#### 7. Processing Screen

**Purpose:** Show real-time progress during file organization

**Display:**
- Current operation
- Files processed / total
- Progress bar
- Success/error messages

**What happens:**
- Creates target directories
- Moves/copies files
- Logs operations to `.abook-org.log`
- Shows final summary

**Duration:** Depends on file count and sizes

### After Organization

**Success:** Shows summary and exits
**Errors:** Displays error messages with details

**Undo:** Use CLI undo command if needed:
```bash
audiobook-organizer --dir=/path --undo
```

---

## Rename TUI

The **Rename TUI** provides an interactive workflow for renaming audiobook files based on metadata templates.

### Launch

```bash
# Interactive rename workflow
audiobook-organizer rename-tui --dir=/path/to/audiobooks

# With specific directory
audiobook-organizer rename-tui --dir=/media/audiobooks

# Preview mode (no changes)
audiobook-organizer rename-tui --dir=/path --dry-run
```

### Workflow Screens

The rename TUI follows a 6-screen workflow:

#### 1. Scan Screen

**Purpose:** Discover audiobook files and extract metadata

**What happens:**
- Scans directory recursively for audio files
- Extracts metadata from files
- Displays progress

#### 2. Field Mapping Screen

**Purpose:** Configure which metadata fields to use for different purposes

**Display:**
- **Hybrid Metadata Display** - Shows both metadata.json fields (📁) and embedded audio fields (🎵)
- **Visual indicators** - Icons show field source
- **Field selectors** - Configure author, title, series, track fields

**Available fields:**
- **Author Fields** - Priority list: `authors`, `narrators`, `album_artist`, `artist`, etc.
- **Title Field** - Single field: `album`, `title`, `track_title`, etc.
- **Series Field** - Single field: `series`, `album`, etc.
- **Track Field** - Single field: `track`, `track_number`, etc.

**Navigation:**
- `↑/↓` - Navigate field selectors
- `Enter` / `Space` - Edit field
- `Tab` - Move between sections
- `Enter on "Continue"` - Proceed to command screen
- `q` - Back to scan

**Tips:**
- Review sample metadata to see which fields are populated
- Order author fields by priority (tried first to last)

**See also:** [METADATA.md](METADATA.md#field-mapping) for field mapping guide

#### 3. Command Screen (Optional Exit Point)

**Purpose:** Generate CLI commands based on current configuration

**Display:**
- Complete CLI command with all flags
- Copy-pasteable for future use
- Option to continue or exit

**Navigation:**
- `c` - Copy command to clipboard (if supported)
- `Enter on "Continue"` - Proceed to template builder
- `Enter on "Exit"` - Quit TUI (command shown for manual use)
- `q` - Back to field mapping

**Example command:**
```bash
audiobook-organizer rename \
  --dir=/path/to/audiobooks \
  --template="{author} - {series} {series_number} - {title}" \
  --author-fields="narrators,authors" \
  --title-field="album"
```

#### 4. Template Screen

**Purpose:** Design filename template with live preview

**Features:**
- **Template builder** - Select fields and separators
- **Available fields**:
  - `{author}` - First author (formatted)
  - `{authors}` - All authors (comma-separated)
  - `{title}` - Book title
  - `{series}` - Series name
  - `{series_number}` - Series number only
  - `{track}` - Track number (zero-padded)
  - `{album}` - Album field
  - `{year}` - Publication year
  - `{narrator}` - Narrator (if available)
- **Live preview** - Shows how template will look with sample data
- **Separator configuration** - Choose between `-`, `/`, ` `, `.`, `_`

**Navigation:**
- `↑/↓` - Navigate field list
- `Enter` / `Space` - Add field to template
- `Backspace` - Remove last field
- `Tab` - Change separator
- `Enter on "Continue"` - Proceed to preview
- `q` - Back to command screen

**Example template:**
```
Template: {author} - {track} - {title}
Preview: Brandon Sanderson - 01 - The Final Empire.m4b
```

#### 5. Preview Screen

**Purpose:** Review all proposed renames before executing

**Display:**
- **Before** → **After** filenames
- **Color-coded indicators** - Show changed portions
- **Conflict detection** - Highlights duplicate filenames
- **Summary** - Total renames, conflicts found

**Navigation:**
- `Scroll` - Review all renames
- `Enter on "Execute"` - Perform renames
- `Enter on "Cancel"` - Return to template
- `q` - Back to template screen

**Tips:**
- Check for conflicts (duplicate target filenames)
- Verify template produces desired results
- Use `q` to adjust template if needed

#### 6. Process Screen

**Purpose:** Execute renames with progress tracking

**Display:**
- Current file being renamed
- Progress bar
- Success/error messages
- Operation log path

**What happens:**
- Renames files according to template
- Logs operations to `.abook-org.log`
- Shows final summary

**Undo support:** Operations logged for reversal:
```bash
audiobook-organizer rename --dir=/path --undo
```

---

## Metadata Viewer

The **Metadata Viewer** lets you explore metadata for audiobook files interactively.

### Launch

```bash
# Interactive metadata exploration
audiobook-organizer metadata --dir=/path/to/audiobooks
```

### Features

- **Browse files** - Navigate directory tree
- **View metadata** - See all extracted fields
- **Multiple formats** - Supports metadata.json, EPUB, MP3, M4B
- **Field indicators** - Shows which fields are populated
- **Export options** - Copy metadata to clipboard

### Use Cases

- **Understand metadata structure** - Before designing field mappings
- **Troubleshoot extraction** - Verify which fields are populated
- **Compare formats** - See differences between metadata.json and embedded tags
- **Plan templates** - Determine which fields to use in rename templates

### Navigation

- `↑/↓` - Navigate files
- `Enter` - View metadata details
- `q` - Back / Exit
- `Tab` - Switch between list and detail view

---

## Hybrid Metadata Extraction

**NEW Feature:** When `metadata.json` files exist alongside audio files (MP3, M4B), the TUI automatically merges:

- **Book-level metadata** from `metadata.json`:
  - Title, authors, series, description
  - Publisher, publication date, genres
- **File-level metadata** from embedded audio tags:
  - Track numbers, disc numbers
  - Individual chapter titles

**Benefits:**
- Complete picture with proper track numbering
- Best of both metadata sources
- Essential for multi-file audiobooks

**How it works:**
1. TUI reads `metadata.json` for book details
2. TUI reads audio files for track information
3. Merges data, preferring `metadata.json` for book fields
4. Uses embedded tags for track/disc numbers

---

## Keyboard Shortcuts Reference

### Global (All Screens)

| Key | Action |
|-----|--------|
| `q` | Back / Previous screen |
| `Ctrl+C` | Quit TUI immediately |
| `Ctrl+Q` | Quit TUI (some screens) |

### Directory Picker

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Open directory / Navigate into |
| `Ctrl+S` | Select current directory |
| **Type** | Filter directories |
| `ESC` | Clear filter |
| `Ctrl+B` | Go up one level (parent) |
| `Ctrl+H` | Jump to home directory |
| `Ctrl+R` | Jump to root directory |

### Lists (Book List, File List)

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Space` | Toggle selection |
| `a` | Select all |
| `n` | Deselect all |
| `Enter` | Continue with selected |

### Forms (Settings, Field Mapping)

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous field |
| `↓` / `j` | Next field |
| `Tab` | Move between sections |
| `Enter` | Edit field / Toggle option |
| `Space` | Toggle checkbox |
| `ESC` | Cancel edit |

### Preview Screens

| Key | Action |
|-----|--------|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `PgUp` | Page up |
| `PgDn` | Page down |
| `Home` | Top of list |
| `End` | Bottom of list |

---

## Tips & Tricks

### Filtering Directories

When in the directory picker, start typing to filter:
```
/home/user/media/audiobooks/
Type: "fantasy" → Shows only directories with "fantasy" in path
```

Press `ESC` to clear filter and see all directories again.

### Quick Navigation

Use shortcuts to jump to common locations:
- `Ctrl+H` - Home directory (`~`)
- `Ctrl+R` - Root directory (`/`)
- `Ctrl+B` - Parent directory (`..`)

### Review Before Executing

Always review the preview screen carefully:
- Check for incorrect paths
- Look for conflicts
- Verify metadata extracted correctly

Use `q` to go back and adjust settings if needed.

### Field Mapping Strategy

For non-standard MP3 files:
1. Use metadata viewer to explore available fields
2. Identify which fields contain author, title, series
3. Configure field mapping in rename TUI
4. Verify with template preview before executing

### Dry Run First

Test your configuration without making changes:
```bash
audiobook-organizer tui --dir=/path --dry-run
```

Or use the preview screen to review operations before executing.

---

## Troubleshooting

### TUI doesn't render correctly

**Symptoms:** Garbled output, broken boxes, missing colors

**Solutions:**
- Update terminal emulator to latest version
- Try different terminal (iTerm2, Alacritty, Windows Terminal)
- Check `$TERM` environment variable:
  ```bash
  echo $TERM
  # Should be: xterm-256color or similar
  ```
- Force 256-color mode:
  ```bash
  TERM=xterm-256color audiobook-organizer tui
  ```

### SSH rendering issues

**Symptoms:** TUI broken over SSH connection

**Solutions:**
- Ensure terminal supports 256 colors on both client and server
- Use `-t` flag with SSH to force TTY allocation:
  ```bash
  ssh -t user@server audiobook-organizer tui
  ```
- Consider using tmux/screen for better compatibility

### Keyboard shortcuts don't work

**Symptoms:** Keys like `Ctrl+H` trigger wrong actions

**Solutions:**
- Check terminal key bindings (may conflict)
- Use mouse if terminal supports it (for directory picker)
- Use alternative navigation (`↑/↓` instead of shortcuts)

### Filter not working

**Symptoms:** Typing doesn't filter directories

**Solutions:**
- Ensure you're on the directory picker screen
- Press `ESC` to clear any existing filter first
- Check keyboard layout (some keys may not register)

---

## Comparison with Other Modes

| Feature | GUI | TUI | CLI |
|---------|-----|-----|-----|
| **Visual Interface** | ✓ Modern desktop | ✓ Terminal UI | ✗ Text only |
| **Mouse Support** | ✓ Full | Limited | ✗ N/A |
| **Keyboard Navigation** | ✓ | ✓ Full | ✗ N/A |
| **SSH/Remote** | ✗ Requires X11 | ✓ Works great | ✓ Works |
| **Interactive Workflow** | ✓ | ✓ | ✗ Direct execution |
| **Scriptable** | ✗ | ✗ | ✓ Full |
| **Field Mapping UI** | ✓ Dialog | ✓ Screen | ✗ Config/flags |
| **Real-time Preview** | ✓ Live updates | ✓ Preview screen | ✗ Dry-run |
| **Learning Curve** | Low | Medium | Medium-High |

**Use TUI when:**
- Working over SSH or remote connections
- Prefer keyboard navigation over mouse
- Want interactive workflow without GUI overhead
- Need visual feedback but CLI is too bare

**Use GUI when:**
- First time using the organizer
- Prefer mouse-driven interface
- Want most visual feedback

**Use CLI when:**
- Automating with scripts
- Batch processing
- Non-interactive execution

---

## See Also

- [GUI.md](GUI.md) - Desktop GUI guide
- [CLI.md](CLI.md) - Command-line reference
- [METADATA.md](METADATA.md) - Metadata extraction guide
- [LAYOUTS.md](LAYOUTS.md) - Directory layout options
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file format
- [Main README](../README.md) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Feature requests:** [GitHub Discussions](https://github.com/jeeftor/audiobook-organizer/discussions)
- **Questions:** [GitHub Discussions Q&A](https://github.com/jeeftor/audiobook-organizer/discussions/categories/q-a)

When reporting TUI issues, please include:
- Operating system and terminal emulator
- Version (`audiobook-organizer version`)
- `$TERM` environment variable value
- Screenshot or text copy of broken rendering
- Steps to reproduce
