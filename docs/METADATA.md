# Metadata Extraction Guide

This guide explains how the Audiobook Organizer extracts metadata from various file formats and how to configure field mapping for non-standard metadata structures.

## Overview

The organizer can extract audiobook metadata from four sources:

1. **metadata.json files** - JSON files created by Audiobookshelf
2. **Embedded EPUB metadata** - Dublin Core metadata in EPUB files
3. **Embedded MP3 tags** - ID3v2 tags in MP3 audio files
4. **Embedded M4B tags** - iTunes-style metadata in M4B audio files

**Hybrid mode:** When metadata.json exists alongside audio files, the organizer automatically merges book-level metadata from JSON with track-level metadata from audio files.

---

## Metadata Sources

### 1. metadata.json Files (Audiobookshelf)

**Default mode.** The organizer looks for `metadata.json` files created by Audiobookshelf.

**File structure:**
```json
{
  "title": "The Final Empire",
  "authors": ["Brandon Sanderson"],
  "series": ["Mistborn #1"],
  "publisher": "Tor Books",
  "publishedYear": "2006",
  "description": "...",
  "genres": ["Fantasy", "Epic Fantasy"],
  "narrator": "Michael Kramer"
}
```

**How to use:**
```bash
# Automatic if metadata.json files exist
audiobook-organizer --dir=/path/to/audiobooks
```

**Pros:**
- Clean, structured metadata
- Book-level information (not per-file)
- Consistent format

**Cons:**
- Requires Audiobookshelf or manual JSON creation
- Doesn't include track numbers for multi-file books

### 2. Embedded EPUB Metadata

**Uses Dublin Core metadata** embedded in EPUB files.

**Extracted fields:**
- Title (`dc:title`)
- Authors (`dc:creator`)
- Publisher (`dc:publisher`)
- Publication date (`dc:date`)
- Description (`dc:description`)
- Language (`dc:language`)

**How to use:**
```bash
audiobook-organizer --dir=/path/to/epubs --use-embedded-metadata
```

**Pros:**
- Self-contained (no external metadata needed)
- Standardized Dublin Core format

**Cons:**
- No track/disc numbers (EPUB is single file)
- Limited series information

### 3. Embedded MP3 Tags (ID3v2)

**Extracts ID3v2 tags** from MP3 audio files.

**Common fields:**
- Title (`TIT2`)
- Artist (`TPE1`)
- Album (`TALB`)
- Album Artist (`TPE2`)
- Track Number (`TRCK`)
- Disc Number (`TPOS`)
- Year (`TYER` or `TDRC`)
- Comment (`COMM`)

**How to use:**
```bash
audiobook-organizer --dir=/path/to/mp3s --use-embedded-metadata
```

**Pros:**
- Widely supported
- Per-file metadata (track numbers)
- No external files needed

**Cons:**
- Inconsistent field usage (see Field Mapping below)
- May lack book-level metadata

### 4. Embedded M4B Tags (iTunes)

**Extracts iTunes-style metadata** from M4B audio files.

**Common fields:**
- Title (`©nam`)
- Artist (`©ART`)
- Album (`©alb`)
- Album Artist (`aART`)
- Track Number (`trkn`)
- Disc Number (`disk`)
- Year (`©day`)
- Comment (`©cmt`)

**How to use:**
```bash
audiobook-organizer --dir=/path/to/m4bs --use-embedded-metadata
```

**Pros:**
- Audiobook-specific format
- Good metadata support
- Per-file track information

**Cons:**
- Less common than MP3
- Field mapping still may be needed

---

## Hybrid Metadata Mode

**NEW Feature:** Automatic merging of metadata.json with embedded audio tags.

### How It Works

When both `metadata.json` and audio files exist in the same directory:

1. **Book-level metadata** comes from `metadata.json`:
   - Title, authors, series
   - Publisher, publication date
   - Description, genres
   - Narrator

2. **File-level metadata** comes from embedded tags:
   - Track numbers
   - Disc numbers
   - Individual chapter titles (if different from book title)

3. **Merging logic:**
   - `metadata.json` fields take priority for book info
   - Audio file tags used for track/disc numbers
   - Best of both sources

### Example

**Directory structure:**
```
/audiobooks/Mistborn/
├── metadata.json          # Contains: title, authors, series
├── track01.mp3            # Contains: track=1, disc=1
├── track02.mp3            # Contains: track=2, disc=1
└── track03.mp3            # Contains: track=3, disc=1
```

**Result:**
- **Title:** From metadata.json → "The Final Empire"
- **Author:** From metadata.json → "Brandon Sanderson"
- **Series:** From metadata.json → "Mistborn #1"
- **Track Numbers:** From MP3 files → 1, 2, 3
- **Disc Numbers:** From MP3 files → 1, 1, 1

**Benefits:**
- Complete metadata picture
- Proper track numbering for multi-file books
- No manual intervention needed

### Enabling Hybrid Mode

**Automatic** - No special flags needed. If both metadata.json and audio files exist, hybrid mode activates automatically.

```bash
# Just run normally
audiobook-organizer --dir=/path/to/audiobooks
```

---

## Flat vs Non-Flat Processing

The `--flat` flag changes how files are discovered and grouped.

### Non-Flat Mode (Default)

**Behavior:**
- All files in a directory treated as one book
- Shared metadata per directory
- Directory structure implies book grouping

**Good for:**
- Multi-file audiobooks
- Pre-organized collections
- Books split across multiple MP3/M4B files

**Example:**
```
Input:
  /books/
    Mistborn_The_Final_Empire/
      chapter1.mp3  # All belong to same book
      chapter2.mp3
      chapter3.mp3

Output:
  /organized/
    Brandon Sanderson/
      Mistborn/
        The Final Empire/
          chapter1.mp3
          chapter2.mp3
          chapter3.mp3
```

**Command:**
```bash
audiobook-organizer --dir=/books --out=/organized
```

### Flat Mode (`--flat`)

**Behavior:**
- Each file processed independently
- Files grouped ONLY if they have identical metadata
- Ignores directory structure

**Good for:**
- Single-file audiobooks (EPUB, single M4B)
- Mixed collections in one directory
- Files with complete per-file metadata

**Example:**
```
Input:
  /downloads/
    book1.epub         # Metadata: Author A, Book 1
    book2.epub         # Metadata: Author B, Book 2
    book3_ch1.mp3      # Metadata: Author C, Book 3
    book3_ch2.mp3      # Metadata: Author C, Book 3 (same metadata → grouped)

Output:
  /organized/
    Author A/
      Book 1/
        book1.epub
    Author B/
      Book 2/
        book2.epub
    Author C/
      Book 3/
        book3_ch1.mp3
        book3_ch2.mp3  # Grouped because identical metadata
```

**Command:**
```bash
audiobook-organizer --dir=/downloads --out=/organized --flat
```

**Important:** Flat mode auto-enables `--use-embedded-metadata`.

### When to Use Each Mode

| Scenario | Mode | Reason |
|----------|------|--------|
| Multi-file books in separate directories | Non-flat (default) | Directory structure groups files |
| Single-file audiobooks (EPUB, M4B) | Flat | Each file is complete book |
| Mixed collection in one directory | Flat | Files need independent processing |
| Pre-organized library | Non-flat (default) | Directory structure is meaningful |
| Files with complete metadata | Flat | Metadata determines grouping |
| Files with partial metadata | Non-flat (default) | Directory structure provides context |

---

## Field Mapping

**Problem:** Metadata fields aren't standardized. MP3 files might use "artist" for author, "album" for title, etc.

**Solution:** Field mapping tells the organizer which fields contain author, title, series, etc.

### Why Field Mapping is Needed

**Standard audiobook metadata:**
```
Title: The Final Empire
Authors: Brandon Sanderson
Series: Mistborn
Track: 1
```

**Real MP3 file metadata:**
```
title: Chapter 1 (track title, not book title!)
artist: Michael Kramer (narrator, not author!)
album: The Final Empire (book title in album field!)
album_artist: Brandon Sanderson (author in album_artist field!)
```

**Without field mapping:** Organizer uses wrong fields → incorrect organization

**With field mapping:** Organizer knows to use `album` for title, `album_artist` for author

### Field Mapping Options

#### Author Fields

**Comma-separated list** of fields to try in priority order:

```bash
--author-fields="authors,narrators,album_artist,artist"
```

**How it works:**
1. Try `authors` field first
2. If empty, try `narrators`
3. If empty, try `album_artist`
4. If empty, try `artist`
5. Use first non-empty value

**Common configurations:**

| Use Case | Configuration |
|----------|---------------|
| Standard Audiobookshelf | `--author-fields="authors"` |
| Standard MP3s | `--author-fields="artist,album_artist"` |
| Audiobooks with narrator as artist | `--author-fields="narrators,album_artist,artist"` |
| MP3s with author in album_artist | `--author-fields="album_artist,artist,authors"` |

#### Title Field

**Single field** to use for book title:

```bash
--title-field="album"  # Use album tag as title
--title-field="title"  # Use title tag (default)
```

**Common use cases:**

| Use Case | Configuration |
|----------|---------------|
| MP3s where album = book title | `--title-field="album"` |
| Standard metadata | `--title-field="title"` |
| EPUBs | `--title-field="title"` (default works) |

#### Series Field

**Single field** to use for series:

```bash
--series-field="series"  # Use series tag (default)
--series-field="album"   # Use album as series
```

#### Track Field

**Single field** to use for track numbers:

```bash
--track-field="track"         # Use track tag (default)
--track-field="track_number"  # Alternative field name
```

### Field Mapping in Different Modes

#### CLI

Use flags directly:

```bash
audiobook-organizer \
  --dir=/media/audiobooks \
  --use-embedded-metadata \
  --author-fields="album_artist,artist" \
  --title-field="album" \
  --series-field="series"
```

#### Config File

Set in config file:

```yaml
# ~/.audiobook-organizer.yaml
author-fields: "album_artist,artist,narrators"
title-field: "album"
series-field: "series"
track-field: "track"
```

#### Environment Variables

```bash
export AO_AUTHOR_FIELDS="album_artist,artist,narrators"
export AO_TITLE_FIELD="album"
export AO_SERIES_FIELD="series"
export AO_TRACK_FIELD="track"
```

#### GUI

Click "Configure Field Mapping" button → Interactive dialog with dropdowns

#### TUI

Navigate to Field Mapping screen → Select fields with keyboard

### Field Mapping Presets

Common configurations for different file types:

#### Preset 1: Audio (Default)

For well-structured MP3/M4B files:

```bash
--author-fields="authors,album_artist,artist"
--title-field="title"
--series-field="series"
--track-field="track"
```

#### Preset 2: Audio (Album as Title)

For MP3s where book title is in album field:

```bash
--author-fields="album_artist,artist,narrators"
--title-field="album"
--series-field="series"
--track-field="track"
```

#### Preset 3: Audiobookshelf + Audio

For hybrid mode (metadata.json + audio files):

```bash
# Book info from metadata.json, track info from audio
--author-fields="authors"  # From JSON
--title-field="title"      # From JSON
--series-field="series"    # From JSON
--track-field="track"      # From audio files
```

---

## Album Detection

**Multi-file audiobooks** are automatically detected and grouped.

### How It Works

Files are grouped into albums when they share:
1. **Same directory** (in non-flat mode)
2. **Same author**
3. **Same title/album**
4. **Same series** (if present)

**Result:** Files treated as chapters of one book

### Example

```
Input files:
  /books/mistborn/chapter01.mp3  # Author: Sanderson, Album: Mistborn
  /books/mistborn/chapter02.mp3  # Author: Sanderson, Album: Mistborn
  /books/mistborn/chapter03.mp3  # Author: Sanderson, Album: Mistborn

Detection:
  ✓ Same directory
  ✓ Same author (Sanderson)
  ✓ Same album (Mistborn)
  → Grouped as one book with 3 chapters

Output:
  /organized/Brandon Sanderson/Mistborn/
    chapter01.mp3
    chapter02.mp3
    chapter03.mp3
```

### Indicators

**GUI/TUI:** Shows album indicator (📀 icon) with file count

**CLI:** Logs "Processing album: Title (X files)"

---

## Metadata Viewer

Use the metadata viewer to explore extracted fields:

```bash
audiobook-organizer metadata --dir=/path/to/audiobooks
```

**Shows:**
- All extracted metadata fields
- Field source (metadata.json, MP3, M4B, EPUB)
- Which fields are populated
- Color-coded indicators for mapped fields

**Use cases:**
- Understand metadata structure before organizing
- Design field mapping configuration
- Troubleshoot extraction issues
- Compare different file formats

---

## Troubleshooting Metadata Issues

### No metadata found

**Symptoms:** Scan completes but no audiobooks found, or books have empty metadata

**Solutions:**
1. Verify files contain metadata:
   ```bash
   audiobook-organizer metadata --dir=/path
   ```
2. Check file formats are supported (EPUB, MP3, M4B, metadata.json)
3. Try `--use-embedded-metadata` if no metadata.json files
4. Use `--flat` for single-file audiobooks

### Wrong author/title extracted

**Symptoms:** Books organized under wrong author or title

**Solutions:**
1. Use metadata viewer to see available fields:
   ```bash
   audiobook-organizer metadata --dir=/path
   ```
2. Configure field mapping:
   ```bash
   audiobook-organizer \
     --dir=/path \
     --author-fields="album_artist,artist" \
     --title-field="album"
   ```
3. Use GUI field mapping dialog for visual configuration

### Files not grouped as album

**Symptoms:** Multi-file book split into separate books

**Causes:**
- Inconsistent metadata across files
- Using `--flat` mode (files processed independently)
- Different author/title/album values

**Solutions:**
1. Verify metadata consistency:
   ```bash
   audiobook-organizer metadata --dir=/path
   ```
2. Use non-flat mode (default) for multi-file books
3. Edit file tags to ensure consistent metadata
4. Check track numbers are sequential

### Track numbers missing or wrong

**Symptoms:** Files not in correct order, or numbered incorrectly

**Solutions:**
1. Check track field:
   ```bash
   audiobook-organizer metadata --dir=/path
   ```
2. Configure track field:
   ```bash
   --track-field="track_number"  # If track field is named differently
   ```
3. Verify audio files have track tags
4. Use hybrid mode if metadata.json has track info

### Hybrid mode not working

**Symptoms:** Track numbers still missing despite audio files + metadata.json

**Solutions:**
1. Verify both exist in same directory:
   ```bash
   ls -la /path/to/book/
   # Should show: metadata.json, *.mp3 or *.m4b
   ```
2. Check metadata.json is valid JSON
3. Verify audio files have track tags
4. Use metadata viewer to see merged result

---

## Best Practices

### 1. Preview First

Always use `--dry-run` to preview metadata extraction:

```bash
audiobook-organizer --dir=/path --dry-run --verbose
```

### 2. Start with Metadata Viewer

Before organizing, explore metadata:

```bash
audiobook-organizer metadata --dir=/path
```

### 3. Test Field Mapping

Test field mapping on a small subset:

```bash
audiobook-organizer \
  --dir=/path/to/test-book \
  --author-fields="album_artist,artist" \
  --title-field="album" \
  --dry-run
```

### 4. Use Config Files for Complex Mappings

Save complex field mappings to config file:

```yaml
# ~/.audiobook-organizer.yaml
author-fields: "narrators,album_artist,artist,authors,composer"
title-field: "album,title"
series-field: "series,album,grouping"
track-field: "track,track_number,trck"
```

### 5. Document Your Configuration

Add comments to config files:

```yaml
# Field mapping for Libby MP3 downloads
# - Author is in album_artist field
# - Book title is in album field
# - Narrator is in artist field
author-fields: "album_artist,artist"
title-field: "album"
```

---

## See Also

- [CLI.md](CLI.md) - Command-line reference for field mapping flags
- [GUI.md](GUI.md) - GUI field mapping dialog documentation
- [TUI.md](TUI.md) - TUI field mapping screen guide
- [CONFIGURATION.md](CONFIGURATION.md) - Config file format for field mapping
- [Main README](../README.md) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Metadata questions:** [GitHub Discussions Q&A](https://github.com/jeeftor/audiobook-organizer/discussions/categories/q-a)

When reporting metadata issues, please include:
- File format (MP3, M4B, EPUB, metadata.json)
- Output of `audiobook-organizer metadata --dir=/path`
- Expected vs actual metadata
- Sample file (if possible)
