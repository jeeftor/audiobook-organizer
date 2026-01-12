# Audiobook Organizer GUI

The **Audiobook Organizer GUI** is a modern desktop application built with Wails that provides a visual, point-and-click interface for organizing your audiobook library.

## Overview

The GUI offers several advantages over the CLI and TUI modes:

- **Visual workflow** - Step-by-step wizard through 4 screens
- **Native file dialogs** - OS-native directory pickers
- **Live metadata preview** - See extracted metadata with color-coded field indicators
- **Interactive field mapping** - Visual configuration dialog for metadata sources
- **Real-time path preview** - See exactly where files will be organized as you configure
- **Conflict detection** - Visual highlighting of naming conflicts
- **Filename template builder** - Intuitive UI for customizing file names
- **Mouse-driven** - No need to remember keyboard shortcuts

**Best for:** First-time users, exploring metadata, visual configuration, one-time organization tasks

**Not ideal for:** Automation, batch processing, SSH/remote sessions (use CLI instead)

---

## Installation

### macOS

**Download from releases:**
```bash
# Download the .dmg file from GitHub releases
# Or install via Homebrew Cask (coming soon)
brew install --cask audiobook-organizer-gui
```

**Run the app:**
```bash
audiobook-organizer-gui
```

### Linux

**Download from releases:**
- **Debian/Ubuntu**: Download `.deb` file
  ```bash
  sudo dpkg -i audiobook-organizer-gui_*.deb
  ```
- **RedHat/Fedora**: Download `.rpm` file
  ```bash
  sudo rpm -i audiobook-organizer-gui_*.rpm
  ```
- **Universal**: Download `.AppImage` file
  ```bash
  chmod +x audiobook-organizer-gui_*.AppImage
  ./audiobook-organizer-gui_*.AppImage
  ```

### Windows

**Download from releases:**
```
Download: audiobook-organizer-gui-setup.exe
Double-click to install
```

### Beta Releases

To try pre-release features from `feature/*` branches:
1. Go to [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases)
2. Look for releases tagged with `-beta`
3. Download platform-specific binaries

**Note:** Beta releases may be unstable. Use at your own risk.

---

## Getting Started

### First Launch

When you launch the GUI for the first time:

```bash
audiobook-organizer-gui
```

You'll see the **Directory Picker** screen where you select input and output directories.

### CLI Integration

You can pre-populate directories via command-line flags to skip the picker screen:

```bash
# Auto-advance to book list with directories set
audiobook-organizer-gui --dir=/path/to/audiobooks --out=/path/to/organized
```

This is useful for:
- Creating desktop shortcuts with specific directories
- Launching from file managers
- Scripting GUI launches

---

## Workflow Screens

The GUI guides you through a 4-step workflow:

### Screen 1: Directory Picker

**Purpose:** Select input directory (containing audiobooks) and output directory (where they'll be organized).

![Directory Picker](screenshots/gui-directory-picker.png)

**Features:**
- **Native file browser dialogs** - Click "Browse..." to open OS-native directory picker
- **Path display** - Selected paths shown in read-only input fields
- **Validation** - "Next" button disabled until both directories selected
- **Error messaging** - Clear feedback if directories are invalid

**Navigation:**
- Click "Browse..." buttons to select directories
- Click "Next: Scan for Audiobooks" to proceed (enabled when both directories selected)

**Tips:**
- Input and output can be the same directory (organize in-place)
- Use separate output directory to preserve original structure

---

### Screen 2: Book List & Configuration

**Purpose:** Review scanned audiobooks, configure organization settings, and select which books to organize.

![Book List Screen](screenshots/gui-main-screen.png)

This is the most feature-rich screen with multiple panels and configuration options.

#### Top Bar: Metadata Scanning Mode

Three buttons to select metadata source:

1. **metadata.json** (default)
   - Looks for `metadata.json` files created by Audiobookshelf
   - **Auto-fallback:** If no metadata.json files found, automatically switches to embedded mode

2. **embedded (directory)**
   - Extracts metadata from audio file tags (MP3, M4B)
   - Treats all files in a directory as one book/album
   - Good for multi-file audiobooks

3. **embedded (file)**
   - Extracts metadata from audio file tags
   - Treats each file separately
   - Good for single-file audiobooks or mixed collections

**Current mode displayed:** e.g., "Using metadata.json (3 books found)" or "Using embedded metadata (5 books found)"

#### Layout Options

Dropdown selector with 5 layout options:

| Layout | Example Path | Use Case |
|--------|--------------|----------|
| `author-series-title` | `Author/Series/Title/` | Standard (default) |
| `author-series-title-number` | `Author/Series/#1 - Title/` | Numbered series |
| `author-series` | `Author/Series/` | Multi-file books in series folder |
| `author-title` | `Author/Title/` | No series level |
| `author-only` | `Author/` | Flatten all books per author |

**See also:** [LAYOUTS.md](LAYOUTS.md) for detailed layout comparison

#### Field Mapping Configuration

**Button:** "Configure Field Mapping"

Opens a dialog with interactive field mapping controls:

![Field Mapping Dialog](screenshots/gui-field-mapping.png)

**Purpose:** Tell the organizer which metadata fields to use when extracting from audio files. Essential for MP3 files with non-standard tag structures.

**Configuration options:**

1. **Title Field** (dropdown)
   - Options: `title`, `album`, `series`, `name`, `book`, `work`
   - Default: `title`
   - Example: If your MP3s store book title in the "album" tag, select "album"

2. **Series Field** (dropdown)
   - Options: `series`, `album`, `title`, `name`, `book`, `work`
   - Default: `series`
   - Example: Some files store series in "album" tag

3. **Author Fields** (multi-select, priority order)
   - Options: `authors`, `artist`, `album_artist`, `narrator`, `narrators`, `creator`, `author`, `writer`, `composer`
   - Default: `authors`, `artist`, `album_artist`
   - Example: Try "artist" first, fallback to "album_artist" if not found

4. **Track Field** (dropdown)
   - Options: `track`, `track_number`, `trck`, `trk`, `tracknumber`, `disc`, `discnumber`, `disk`, `tpos`, `disc_number`
   - Default: `track`

5. **Disc Field** (dropdown)
   - Same options as Track Field
   - Default: `disc`

**Live Preview:**
Shows up to 3 sample audiobooks with:
- Raw metadata fields
- Color-coded indicators showing which fields map to which categories:
  - 🟢 Green = TITLE
  - 🟠 Orange = AUTHOR
  - 🔵 Cyan = SERIES
  - 🔵 Blue = TRACK
  - 🟣 Purple = DISC

**Refresh button:** Reload preview with updated field mapping

**See also:** [METADATA.md](METADATA.md) for field mapping deep dive

#### Metadata Preview Panel

**Location:** Bottom section, scrollable area

Shows up to 3 sample audiobooks with complete metadata:
- Filename
- Source type (metadata.json, MP3, M4B, EPUB)
- All extracted metadata fields
- Color-coded field indicators

**Navigation:** Use `≪ < > ≫` buttons to scroll through previews

**Font size:** 10px for compact display

**Purpose:** Verify metadata extraction is working correctly before organizing

#### Input Files Panel

**Location:** Left panel

Lists all scanned audiobooks:
- Checkbox for each file (select which to organize)
- Full file paths
- Scrollable list (max-height: 64)

**Selection controls:**
- Individual checkboxes per file
- "Select All" / "Deselect All" toggle button
- Selection count displayed: "X files selected"

#### Output Files Panel

**Location:** Right panel

**Live path preview** showing where each selected file will be organized:

**Color-coded path components:**
- **Gray** - Path prefix
- **🟠 Orange** - Author
- **🔵 Cyan** - Series
- **🟢 Green** - Title
- **Gray** - Filename

**Real-time updates:** Changes instantly when you:
- Switch layout options
- Modify field mapping
- Change metadata scanning mode

**Purpose:** See exactly where files will go before executing

#### Navigation

- **Back** - Return to directory picker
- **Next** - Proceed to preview screen
  - Shows selection count: "Next (X files selected)"
  - Disabled if no files selected

---

### Screen 3: Preview Changes

**Purpose:** Review all file operations before execution, configure filename templates, and detect conflicts.

![Preview Changes Screen](screenshots/gui-preview-changes.png)

#### Preview Display

Shows all file operations in a scrollable list:
- **From path** → **To path** for each operation
- **Color-coded paths** matching the book list screen
- **Conflict highlighting** - Yellow background for conflicting operations
- **Conflict count badge** - Shows number of affected operations

**Scrollable:** Max-height 70vh to fit many operations

#### File Naming Options

Two toggles to control filename handling:

1. **Keep Original Names** (toggle)
   - When enabled: Preserves original filenames
   - When disabled: Allows custom filename templates

2. **Rename Files** (toggle)
   - When enabled: Shows "Configure Template" button
   - When disabled: Uses original filenames or default pattern

#### Filename Template Builder

**Button:** "Configure Template" (visible when "Rename Files" is enabled)

Opens a dialog to customize filename patterns:

![Template Builder Dialog](screenshots/gui-template-builder.png)

**Available fields:**
- `author` - First author (formatted)
- `series` - Series name
- `track` - Track number (zero-padded)
- `title` - Book/chapter title

**Template configuration:**
- **4 template slots** - Assign fields to positions 1-4
- **Field assignment buttons** - Click to assign field to slot
- **Clear buttons** - Remove field from slot
- **Separator selection** - Choose from: `-`, `/`, ` ` (space), `.`, `_`
- **Live preview** - Shows template format: `{author} - {series} - {title}.{ext}`

**Example template:**
```
Slot 1: author
Slot 2: series
Slot 3: title
Separator: " - "

Result: Brandon Sanderson - Mistborn - The Final Empire.m4b
```

#### Conflict Detection

**Conflict types:**
- Multiple books with identical target paths
- Filename collisions after renaming

**Visual indicators:**
- Yellow background highlight
- Conflict reason displayed
- Count badge showing affected operations

**Execution blocking:** "Execute" button disabled if conflicts detected

**Resolution:** Adjust field mapping, layout, or filename template to resolve conflicts

#### Execution Controls

- **Back** - Return to book list
- **Execute** - Perform file organization
  - Disabled if conflicts exist
  - Shows "Organizing..." state during execution
  - Triggers completion screen on success

---

### Screen 4: Completion

**Purpose:** Confirm successful organization.

**Display:**
- Success message with checkmark emoji
- Summary of operations performed

**Navigation:**
- **Organize More Files** - Resets workflow to directory picker

---

## Advanced Features

### CLI Integration

The GUI accepts all standard command-line flags:

```bash
# Pre-populate directories
audiobook-organizer-gui --dir=/books --out=/organized

# Set default layout
audiobook-organizer-gui --layout=author-title

# Enable verbose logging
audiobook-organizer-gui --verbose
```

**Note:** CLI flags set initial values but can be changed in the GUI.

### Configuration File Support

The GUI respects configuration files:
- `~/.audiobook-organizer.yaml`
- `./.audiobook-organizer.yaml`
- Custom path via `--config` flag

**See:** [CONFIGURATION.md](CONFIGURATION.md) for config file format

### Metadata Fallback

**Smart auto-fallback:** If you select "metadata.json" mode but no metadata.json files are found, the GUI automatically:
1. Switches to "embedded (directory)" mode
2. Rescans the directory
3. Displays notification of the switch

This prevents empty scan results and improves user experience.

---

## Keyboard Shortcuts

Currently, the GUI is primarily mouse-driven. Future versions may add:
- `Ctrl+O` - Open directory picker
- `Ctrl+Enter` - Execute organization
- `Ctrl+Z` - Undo (when implemented)
- `Esc` - Cancel/back

---

## Troubleshooting

### No books found

**Symptoms:** Scan completes but shows "0 books found"

**Solutions:**
1. **Check metadata mode** - Try switching between metadata.json and embedded modes
2. **Verify file types** - Ensure directory contains supported formats (MP3, M4B, EPUB, or metadata.json)
3. **Check directory structure** - For embedded (directory) mode, files should be grouped in subdirectories
4. **Review logs** - Enable verbose mode (`--verbose`) and check console output

### Field mapping not working

**Symptoms:** Metadata extracted incorrectly (wrong author, title, etc.)

**Solutions:**
1. **Open field mapping dialog** - Click "Configure Field Mapping"
2. **Check metadata preview** - Review raw metadata fields with color indicators
3. **Adjust field priorities** - Reorder author fields or change title/series field selection
4. **Refresh preview** - Click refresh button to see changes
5. **See also:** [METADATA.md](METADATA.md#field-mapping) for field mapping guide

### Conflicts detected

**Symptoms:** "Execute" button disabled, yellow highlighting in preview

**Solutions:**
1. **Review conflict reason** - Hover over highlighted items to see why
2. **Adjust layout** - Try different layout option (e.g., `author-series-title-number` instead of `author-series-title`)
3. **Configure field mapping** - Ensure unique metadata extraction
4. **Check filename template** - Verify template includes enough unique fields

### GUI doesn't launch

**Symptoms:** Application fails to start or crashes immediately

**Solutions:**

**macOS:**
- Right-click app → Open (bypass Gatekeeper on first launch)
- Check for WebView2 requirement (usually built-in)

**Linux:**
- Ensure `libwebkit2gtk-4.0` is installed
  ```bash
  sudo apt install libwebkit2gtk-4.0-37  # Debian/Ubuntu
  sudo yum install webkit2gtk3           # RedHat/Fedora
  ```

**Windows:**
- Install WebView2 runtime if on Windows 10
- Download from: https://developer.microsoft.com/en-us/microsoft-edge/webview2/

### File operations fail

**Symptoms:** Execution starts but some files fail to move

**Possible causes:**
- **Permissions** - Ensure write access to output directory
- **Disk space** - Verify sufficient space in output location
- **File locks** - Close any programs using the audiobook files
- **Path length** - Windows has 260-character path limit (use shorter paths)

**Check logs:** Operations are logged to `.abook-org.log` in output directory

---

## Comparison with Other Modes

| Feature | GUI | TUI | CLI |
|---------|-----|-----|-----|
| **Visual Interface** | ✓ Modern desktop UI | ✓ Terminal UI | ✗ Text output only |
| **Mouse Support** | ✓ Full | ✗ Keyboard only | ✗ N/A |
| **Native File Dialogs** | ✓ OS-native | ✗ Text-based | ✗ Flags only |
| **Live Metadata Preview** | ✓ With color coding | ✓ Text-based | ✗ Dry-run only |
| **Field Mapping UI** | ✓ Interactive dialog | ✓ Screen-based | ✗ Config/flags |
| **Template Builder** | ✓ Visual wizard | ✓ Interactive | ✗ String flags |
| **Conflict Detection** | ✓ Highlighted | ✓ Listed | ✗ Log output |
| **Scriptable** | ✗ Interactive only | ✗ Interactive only | ✓ Full |
| **SSH/Remote** | ✗ Requires X11 | ✓ Works | ✓ Works |
| **Batch Processing** | ✗ Manual | ✗ Manual | ✓ Automated |
| **Learning Curve** | Low | Medium | High |

**Use GUI when:**
- First time using the organizer
- Exploring metadata and configuration options
- One-time organization of a library
- You prefer visual interfaces

**Use TUI when:**
- Working over SSH
- Prefer keyboard navigation
- Want interactive workflow but no GUI

**Use CLI when:**
- Automating with scripts
- Batch processing multiple directories
- CI/CD integration
- Cron jobs

---

## See Also

- [TUI.md](TUI.md) - Terminal User Interface guide
- [CLI.md](CLI.md) - Command-line reference
- [METADATA.md](METADATA.md) - Metadata extraction deep dive
- [LAYOUTS.md](LAYOUTS.md) - Directory layout options
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file format
- [Main README](../README.md) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Feature requests:** [GitHub Discussions](https://github.com/jeeftor/audiobook-organizer/discussions)
- **Questions:** [GitHub Discussions Q&A](https://github.com/jeeftor/audiobook-organizer/discussions/categories/q-a)

When reporting GUI issues, please include:
- Operating system and version
- GUI version (`audiobook-organizer-gui --version`)
- Steps to reproduce
- Screenshots if applicable
- Console output if available (`--verbose` flag)
