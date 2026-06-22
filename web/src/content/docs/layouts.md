---
title: "Layouts"
description: "Folder layout options and custom layout templates."
---

# Directory Layout Guide

The Audiobook Organizer supports fixed directory layout patterns and custom layout templates to match your preferred organization style. This guide explains each layout with examples and recommendations.

## Overview

Directory layouts control how audiobooks are organized into folder hierarchies. Choose a layout that matches your library management style and player compatibility needs.

**Configure layout via:**
- **GUI:** Layout selector and custom layout template field in the organize workflow
- **TUI:** Settings screen
- **CLI:** `--layout` flag
- **CLI custom template:** `--layout-template` flag
- **Config file:** `layout: "author-series-title"`
- **Config file custom template:** `layout-template: "{author}/{series}/{series-count} - {title}"`
- **Environment:** `AO_LAYOUT=author-series-title`
- **Environment custom template:** `AO_LAYOUT_TEMPLATE="{author}/{series}/{series-count} - {title}"`

---

## Available Layouts

| Layout | Folder pattern | Best fit |
| --- | --- | --- |
| `author-series-title` | `Author/Series/Title/` | Large libraries with many series |
| `author-series-title-number` | `Author/Series/#N - Title/` | Series where reading order matters |
| `author-series` | `Author/Series/` | Single-file books grouped by series |
| `author-title` | `Author/Title/` | Standalones or mixed libraries without reliable series data |
| `author-only` | `Author/` | Flat, single-file libraries with descriptive filenames |
| `series-title` | `Series/Title/` | Series-first browsing |

### 1. `author-series-title` (Default)

**Full hierarchy with all three levels.**

**Pattern:**
```
Author/
  Series/
    Title/
      files...
```

**Example:**
```
L. Frank Baum/
  Oz/
    The Wonderful Wizard of Oz/
      chapter01.mp3
      chapter02.mp3
      metadata.json
    The Marvelous Land of Oz/
      audiobook.m4b
      metadata.json
Arthur Conan Doyle/
  Sherlock Holmes/
    A Study in Scarlet/
      disc1-track01.mp3
      disc1-track02.mp3
```

**Best for:**
- Large libraries with many series
- Authors with multiple series
- Organizing series-focused collections
- Maximum organization depth

**Pros:**
- Clearest hierarchy
- Easy to browse series
- Groups books by series naturally

**Cons:**
- Deepest nesting (3 levels)
- May be overkill for small libraries

**Use when:**
- You have authors with multiple series (for example, L. Frank Baum or Arthur Conan Doyle)
- Your audiobook player supports deep folder navigation
- You prefer maximum organization

---

## Custom Layout Templates

Use `--layout-template` when the fixed layout names do not describe the folder structure you want. A custom template overrides `--layout`.

For an in-terminal reference of fields, fallback syntax, examples, and path safety rules, run:

```bash
audiobook-organizer layout-template
```

```bash
audiobook-organizer \
  --dir=/books \
  --out=/organized \
  --layout-template="{author}/{series|Standalone}/{Vol series_number:02 - }{book_title}{ [narrator]}"
```

Template fields use the same renderer as rename templates. Both `{field}` and `${field}` are accepted, and fallback values use `|`:

| Field | Example value |
| --- | --- |
| `{author}` | `L. Frank Baum` |
| `{title}` or `{book_title}` | `Ozma of Oz` |
| `{series}` | `Oz` |
| `{series_full}` | `Oz #3` |
| `{series_number}` or `{series-count}` | `3` |
| `{series_number:02}` | `03` |
| `{album}` | `Oz` |
| `{track}` | `01` |
| `{narrator}` or `{narrators}` | `Volunteer Reader` |
| `{year}` | `1907` |

**Composite optional segments** combine literal text with field references inside one `{...}` token. If any referenced field inside the token is empty, the entire token is omitted. Field names inside composites are written without nested braces:

```bash
{Vol series_number:02 - }     # "Vol 02 - " or omitted when series number is missing
{ [narrator]}                  # " [Narrator Name]" or omitted when narrator is missing
```

**Empty path segments:** slash-separated parts such as `{author}/{series}/{title}` omit a segment entirely when it renders empty, so standalones become `Author/Title` rather than creating blank folders. Use `{series|Standalone}` when you want a fallback folder name instead.

Templates can also reference raw metadata keys. Dashes are normalized to underscores, so `{publisher-name}` can read a raw `publisher_name` field.

Each path segment is rendered and sanitized independently, so slashes or other unsafe characters in metadata values cannot create extra directories. Absolute templates and `.` or `..` path segments are rejected.

---

### 2. `author-series-title-number`

**Like `author-series-title` but includes series number in title folder name.**

**Pattern:**
```
Author/
  Series/
    #N - Title/
      files...
```

**Example:**
```
L. Frank Baum/
  Oz/
    #1 - The Wonderful Wizard of Oz/
      chapter01.mp3
    #2 - The Marvelous Land of Oz/
      audiobook.m4b
    #3 - Ozma of Oz/
      audiobook.m4b
Arthur Conan Doyle/
  Sherlock Holmes/
    #1 - A Study in Scarlet/
      disc1.mp3
    #2 - The Sign of the Four/
      disc1.mp3
```

**Best for:**
- Series-focused libraries
- Ensuring correct reading order
- Audiobook players that don't sort by series number

**Pros:**
- Series order visible at a glance
- Natural alphabetical sorting matches reading order
- Prevents accidental out-of-order playback

**Cons:**
- Slightly longer folder names
- Requires series number in metadata

**Use when:**
- Series reading order is critical
- Your player doesn't have series awareness
- You want foolproof sorting

---

### 3. `author-series`

**Groups all books in a series into one folder (no title subfolder).**

**Pattern:**
```
Author/
  Series/
    files...
```

**Example:**
```
L. Frank Baum/
  Oz/
    01 - The Wonderful Wizard of Oz.m4b
    02 - The Marvelous Land of Oz.m4b
    03 - Ozma of Oz.m4b
    metadata.json
Arthur Conan Doyle/
  Sherlock Holmes/
    01 - A Study in Scarlet - disc1.mp3
    01 - A Study in Scarlet - disc2.mp3
    02 - The Sign of the Four.m4b
```

**Best for:**
- Series of single-file audiobooks
- Reducing folder depth
- Players that prefer flat series folders

**Pros:**
- Less nesting (2 levels)
- All series books in one location
- Good for single-file audiobooks

**Cons:**
- Multi-file books mixed together in one folder
- Requires filename template to distinguish books
- Can get cluttered for large series

**Use when:**
- All audiobooks are single files (M4B, EPUB)
- You want shallower folder structure
- Your naming convention makes books distinguishable

---

### 4. `author-title`

**Skips series level entirely.**

**Pattern:**
```
Author/
  Title/
    files...
```

**Example:**
```
L. Frank Baum/
  The Wonderful Wizard of Oz/
    chapter01.mp3
    chapter02.mp3
  The Marvelous Land of Oz/
    audiobook.m4b
  American Fairy Tales/
    audiobook.m4b
Jane Austen/
  Pride and Prejudice/
    audiobook.m4b
  Sense and Sensibility/
    part1.mp3
    part2.mp3
```

**Best for:**
- Authors with few or no series
- Standalone books
- Simple organization needs
- Reducing folder depth

**Pros:**
- Simple 2-level hierarchy
- Easy to browse all books by author
- No series required in metadata

**Cons:**
- Series books scattered (not grouped)
- Harder to track series order
- Less organized for series-heavy authors

**Use when:**
- Most books are standalones
- Authors don't have many series
- You prefer simplicity over series grouping
- Metadata lacks series information

---

### 5. `author-only`

**Flattens all books directly under author folder.**

**Pattern:**
```
Author/
  files...
```

**Example:**
```
L. Frank Baum/
  Oz 01 - The Wonderful Wizard of Oz.m4b
  Oz 02 - The Marvelous Land of Oz.m4b
  American Fairy Tales.m4b
Jane Austen/
  Pride and Prejudice.m4b
  Sense and Sensibility - Part 1.mp3
  Sense and Sensibility - Part 2.mp3
```

**Best for:**
- Single-file audiobooks only
- Extremely flat organization
- Players that handle large flat folders well
- Heavy reliance on filename templates

**Pros:**
- Shallowest structure (1 level)
- All author's books in one place
- Fast to navigate

**Cons:**
- No grouping by series or title
- Multi-file books mixed with single-file books
- Requires detailed filename templates
- Can become cluttered quickly

**Use when:**
- All audiobooks are single files
- You have a robust filename template
- Your player prefers flat folders
- You have very few books per author

**⚠️ Warning:** Not recommended for multi-file audiobooks without careful filename templating.

---

### 6. `series-title`

**Organizes by series first, skipping author level.**

**Pattern:**
```
Series/
  Title/
    files...
```

**Example:**
```
Oz/
  The Wonderful Wizard of Oz/
    chapter01.mp3
    chapter02.mp3
  The Marvelous Land of Oz/
    audiobook.m4b
Sherlock Holmes/
  A Study in Scarlet/
    audiobook.m4b
  The Sign of the Four/
    audiobook.m4b
```

**Best for:**
- Series-first organization
- Libraries with minimal author overlap
- Shared series across authors (e.g., anthologies)

**Pros:**
- Series-focused browsing
- Good for shared-universe series
- Simpler when series spans multiple authors

**Cons:**
- Loses author grouping
- Hard to find all books by one author
- Standalone books need special handling

**Use when:**
- Series are more important than authors
- You browse by series, not author
- Handling anthology series
- Most books are part of series

**⚠️ Note:** Standalone books (no series) will be placed directly in root or need special handling.

---

## Layout Selection Guide

### Decision Tree

**Start here: How is your library organized?**

| Question | If yes | If no |
| --- | --- | --- |
| Are series critical to browsing? | Use `author-series-title` or `author-series-title-number`. | Use `author-title` or `author-only`. |
| Do you organize multi-file books? | Avoid `author-only`; prefer `author-series-title` or `author-title`. | Any layout works; `author-only` or `author-series` keeps nesting low. |
| Do you prefer deep organization? | Use `author-series-title`. | Use `author-title` for balance or `author-only` for the flattest structure. |
| Do you browse primarily by series? | Consider `series-title` or `author-series-title`. | Avoid `series-title`; keep author folders. |

### Quick Recommendations

| Library Type | Recommended Layout | Reason |
|--------------|-------------------|---------|
| **Large multi-series collection** | `author-series-title` | Maximum organization |
| **Order-sensitive series** | `author-series-title-number` | Visible reading order |
| **Single-file audiobooks** | `author-series` or `author-only` | Less nesting needed |
| **Mostly standalones** | `author-title` | Series level unnecessary |
| **Series-first browsing** | `series-title` | Matches browse pattern |
| **Small simple library** | `author-title` | Simple and sufficient |

---

## Special Flags

### `--use-series-as-title`

When enabled, treats series name as the title for books that are part of a series.

**Effect on layouts:**

```bash
# Standard behavior
author-series-title → Author/Series/Title/

# With --use-series-as-title
author-series-title → Author/Series/
```

**Example:**
```bash
# Without flag
L. Frank Baum/Oz/The Wonderful Wizard of Oz/

# With flag
audiobook-organizer --layout=author-series-title --use-series-as-title
L. Frank Baum/Oz/
```

**Use cases:**
- Treating entire series as one title
- Reducing folder depth for series
- When series name is sufficient identifier

---

## Layout Examples with Public-Domain Audiobooks

### Example 1: Classic Series Collection

**Source structure:**
```
/downloads/
  Oz_The_Wonderful_Wizard_of_Oz.m4b
  Oz_The_Marvelous_Land_of_Oz.m4b
  Oz_Ozma_of_Oz.m4b
  A_Study_in_Scarlet.m4b
  The_Sign_of_the_Four.m4b
```

**After organizing with `author-series-title`:**
```
/organized/
  L. Frank Baum/
    Oz/
      The Wonderful Wizard of Oz/
        Oz_The_Wonderful_Wizard_of_Oz.m4b
      The Marvelous Land of Oz/
        Oz_The_Marvelous_Land_of_Oz.m4b
      Ozma of Oz/
        Oz_Ozma_of_Oz.m4b
  Arthur Conan Doyle/
    Sherlock Holmes/
      A Study in Scarlet/
        A_Study_in_Scarlet.m4b
      The Sign of the Four/
        The_Sign_of_the_Four.m4b
```

**After organizing with `author-title`:**
```
/organized/
  L. Frank Baum/
    The Wonderful Wizard of Oz/
      Oz_The_Wonderful_Wizard_of_Oz.m4b
    The Marvelous Land of Oz/
      Oz_The_Marvelous_Land_of_Oz.m4b
    Ozma of Oz/
      Oz_Ozma_of_Oz.m4b
  Arthur Conan Doyle/
    A Study in Scarlet/
      A_Study_in_Scarlet.m4b
    The Sign of the Four/
      The_Sign_of_the_Four.m4b
```

### Example 2: Mixed Collection (Series + Standalones)

**Source structure:**
```
/audiobooks/
  Anne of Green Gables/
    chapter01.mp3
    chapter02.mp3
  Anne of Avonlea/
    chapter01.mp3
  Frankenstein.m4b
  Dracula.m4b
```

**After organizing with `author-series-title`:**
```
/organized/
  L. M. Montgomery/
    Anne of Green Gables/
      Anne of Green Gables/
        chapter01.mp3
        chapter02.mp3
      Anne of Avonlea/
        chapter01.mp3
  Mary Shelley/
    __STANDALONE__/
      Frankenstein/
        Frankenstein.m4b
  Bram Stoker/
    __STANDALONE__/
      Dracula/
        Dracula.m4b
```

**After organizing with `author-title`:**
```
/organized/
  L. M. Montgomery/
    Anne of Green Gables/
      chapter01.mp3
      chapter02.mp3
    Anne of Avonlea/
      chapter01.mp3
  Mary Shelley/
    Frankenstein/
      Frankenstein.m4b
  Bram Stoker/
    Dracula/
      Dracula.m4b
```

---

## Changing Layouts

### Re-organizing with Different Layout

You can change the layout and re-organize:

```bash
# First organization
audiobook-organizer --dir=/books --out=/organized --layout=author-title

# Decided you want series grouping
audiobook-organizer --dir=/organized --out=/reorganized --layout=author-series-title
```

**⚠️ Important:** Always use `--dry-run` first to preview changes.

### Preview Different Layouts

```bash
# Preview with layout A
audiobook-organizer --dir=/books --out=/organized --layout=author-series-title --dry-run

# Preview with layout B
audiobook-organizer --dir=/books --out=/organized --layout=author-title --dry-run
```

Compare outputs to choose the best layout for your needs.

---

## Technical Details

### How Layouts are Calculated

Layouts are calculated in `internal/organizer/organizer.go` by the `LayoutCalculator` struct:

```go
type LayoutCalculator struct {
    Layout             string
    UseSeriesAsTitle   bool
    ReplaceSpace       string
}
```

**Path construction logic:**
1. Extract author, series, title from metadata
2. Apply `InvalidSeriesValue` for standalones
3. Construct path based on layout pattern
4. Sanitize path components (remove invalid characters)
5. Apply space replacement if configured

### Layout Patterns

Internal layout patterns map to path construction:

| Layout | Path Pattern |
|--------|-------------|
| `author-series-title` | `{author}/{series}/{title}` |
| `author-series-title-number` | `{author}/{series}/#{number} - {title}` |
| `author-series` | `{author}/{series}` |
| `author-title` | `{author}/{title}` |
| `author-only` | `{author}` |
| `series-title` | `{series}/{title}` |

**Note:** `{author}`, `{series}`, and `{title}` are sanitized and may have spaces replaced.

---

## Troubleshooting

### Layout produces unexpected paths

**Symptoms:** Files organized differently than expected

**Solutions:**
1. Use `--dry-run` to preview before executing
2. Check metadata extraction with metadata viewer:
   ```bash
   audiobook-organizer metadata-tui --dir=/path
   ```
3. Verify author, series, title fields are populated
4. Consider field mapping if using embedded metadata

### Standalones placed in weird folder

**Symptoms:** Books without series go to `__STANDALONE__` folder

**Cause:** Some layouts (like `author-series-title`) require a series, so standalones get special handling

**Solutions:**
- Use `author-title` layout (skips series level)
- Use `author-only` layout (flattens completely)
- Ensure series field is populated in metadata

### Multi-file books scattered

**Symptoms:** Chapters of one book spread across multiple folders

**Cause:** Using `--flat` mode or `author-only` layout

**Solutions:**
- Use non-flat mode (default) for multi-file books
- Choose layout with title level: `author-series-title` or `author-title`
- Ensure consistent metadata across all files

### Too many nested folders

**Symptoms:** Path too deep, hard to navigate

**Solutions:**
- Use shallower layout: `author-title` or `author-only`
- Enable `--use-series-as-title` to collapse series/title levels
- Use `--replace_space=""` to avoid character substitutions

---

## See Also

- [CLI.md](/audiobook-organizer/cli/) - CLI flags for layout configuration
- [GUI.md](/audiobook-organizer/web-ui/) - GUI layout selector
- [TUI.md](/audiobook-organizer/tui/) - TUI settings screen
- [METADATA.md](/audiobook-organizer/metadata/) - Metadata extraction affecting layout
- [Main README](/audiobook-organizer/) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Layout suggestions:** [GitHub Discussions](https://github.com/jeeftor/audiobook-organizer/discussions)

When reporting layout issues, please include:
- Layout used (`--layout` value)
- Input directory structure
- Expected vs actual output paths
- Metadata for affected files (use metadata viewer)
