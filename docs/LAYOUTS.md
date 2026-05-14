# Directory Layout Guide

The Audiobook Organizer supports **six directory layout patterns** to match your preferred organization style. This guide explains each layout with examples and recommendations.

## Overview

Directory layouts control how audiobooks are organized into folder hierarchies. Choose a layout that matches your library management style and player compatibility needs.

**Configure layout via:**
- **GUI:** Dropdown selector on book list screen
- **TUI:** Settings screen
- **CLI:** `--layout` flag
- **Config file:** `layout: "author-series-title"`
- **Environment:** `AO_LAYOUT=author-series-title`

---

## Available Layouts

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
Brandon Sanderson/
  Mistborn/
    The Final Empire/
      chapter01.mp3
      chapter02.mp3
      metadata.json
    The Well of Ascension/
      audiobook.m4b
      metadata.json
  Stormlight Archive/
    The Way of Kings/
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
- You have authors with multiple series (e.g., Brandon Sanderson, Stephen King)
- Your audiobook player supports deep folder navigation
- You prefer maximum organization

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
Brandon Sanderson/
  Mistborn/
    #1 - The Final Empire/
      chapter01.mp3
    #2 - The Well of Ascension/
      audiobook.m4b
    #3 - The Hero of Ages/
      audiobook.m4b
  Stormlight Archive/
    #1 - The Way of Kings/
      disc1.mp3
    #2 - Words of Radiance/
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
Brandon Sanderson/
  Mistborn/
    01 - The Final Empire.m4b
    02 - The Well of Ascension.m4b
    03 - The Hero of Ages.m4b
    metadata.json
  Stormlight Archive/
    01 - The Way of Kings - disc1.mp3
    01 - The Way of Kings - disc2.mp3
    02 - Words of Radiance.m4b
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
Brandon Sanderson/
  The Final Empire/
    chapter01.mp3
    chapter02.mp3
  The Way of Kings/
    audiobook.m4b
  Elantris/
    audiobook.m4b
Stephen King/
  The Shining/
    audiobook.m4b
  It/
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
Brandon Sanderson/
  Mistborn 01 - The Final Empire.m4b
  Mistborn 02 - The Well of Ascension.m4b
  Stormlight 01 - The Way of Kings.m4b
  Elantris.m4b
Stephen King/
  The Shining.m4b
  It - Part 1.mp3
  It - Part 2.mp3
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
Mistborn/
  The Final Empire/
    chapter01.mp3
    chapter02.mp3
  The Well of Ascension/
    audiobook.m4b
Stormlight Archive/
  The Way of Kings/
    audiobook.m4b
  Words of Radiance/
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

1. **By series importance:**
   - Series are critical → `author-series-title` or `author-series-title-number`
   - Series not important → `author-title` or `author-only`

2. **By file structure:**
   - Multi-file books → Avoid `author-only`, prefer `author-series-title` or `author-title`
   - Single-file books → Any layout works, `author-only` or `author-series` for simplicity

3. **By depth preference:**
   - Prefer deep organization → `author-series-title`
   - Prefer shallow organization → `author-only`
   - Balanced → `author-title`

4. **By primary browse method:**
   - Browse by author → Avoid `series-title`
   - Browse by series → Consider `series-title` or `author-series-title`

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
Brandon Sanderson/Mistborn/The Final Empire/

# With flag
audiobook-organizer --layout=author-series-title --use-series-as-title
Brandon Sanderson/Mistborn/
```

**Use cases:**
- Treating entire series as one title
- Reducing folder depth for series
- When series name is sufficient identifier

---

## Layout Examples with Real Audiobooks

### Example 1: Fantasy Series Collection

**Source structure:**
```
/downloads/
  Mistborn_The_Final_Empire.m4b
  Mistborn_The_Well_of_Ascension.m4b
  Mistborn_The_Hero_of_Ages.m4b
  The_Way_of_Kings.m4b
  Words_of_Radiance.m4b
```

**After organizing with `author-series-title`:**
```
/organized/
  Brandon Sanderson/
    Mistborn/
      The Final Empire/
        Mistborn_The_Final_Empire.m4b
      The Well of Ascension/
        Mistborn_The_Well_of_Ascension.m4b
      The Hero of Ages/
        Mistborn_The_Hero_of_Ages.m4b
    Stormlight Archive/
      The Way of Kings/
        The_Way_of_Kings.m4b
      Words of Radiance/
        Words_of_Radiance.m4b
```

**After organizing with `author-title`:**
```
/organized/
  Brandon Sanderson/
    The Final Empire/
      Mistborn_The_Final_Empire.m4b
    The Well of Ascension/
      Mistborn_The_Well_of_Ascension.m4b
    The Hero of Ages/
      Mistborn_The_Hero_of_Ages.m4b
    The Way of Kings/
      The_Way_of_Kings.m4b
    Words of Radiance/
      Words_of_Radiance.m4b
```

### Example 2: Mixed Collection (Series + Standalones)

**Source structure:**
```
/audiobooks/
  Harry Potter 1/
    chapter01.mp3
    chapter02.mp3
  Harry Potter 2/
    chapter01.mp3
  Enders_Game.m4b
  Ready_Player_One.m4b
```

**After organizing with `author-series-title`:**
```
/organized/
  J.K. Rowling/
    Harry Potter/
      Harry Potter and the Philosopher's Stone/
        chapter01.mp3
        chapter02.mp3
      Harry Potter and the Chamber of Secrets/
        chapter01.mp3
  Orson Scott Card/
    __STANDALONE__/
      Ender's Game/
        Enders_Game.m4b
  Ernest Cline/
    __STANDALONE__/
      Ready Player One/
        Ready_Player_One.m4b
```

**After organizing with `author-title`:**
```
/organized/
  J.K. Rowling/
    Harry Potter and the Philosopher's Stone/
      chapter01.mp3
      chapter02.mp3
    Harry Potter and the Chamber of Secrets/
      chapter01.mp3
  Orson Scott Card/
    Ender's Game/
      Enders_Game.m4b
  Ernest Cline/
    Ready Player One/
      Ready_Player_One.m4b
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
   audiobook-organizer metadata --dir=/path
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

- [CLI.md](CLI.md) - CLI flags for layout configuration
- [GUI.md](GUI.md) - GUI layout selector
- [TUI.md](TUI.md) - TUI settings screen
- [METADATA.md](METADATA.md) - Metadata extraction affecting layout
- [Main README](../README.md) - Project overview

---

## Feedback & Support

- **Bug reports:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Layout suggestions:** [GitHub Discussions](https://github.com/jeeftor/audiobook-organizer/discussions)

When reporting layout issues, please include:
- Layout used (`--layout` value)
- Input directory structure
- Expected vs actual output paths
- Metadata for affected files (use metadata viewer)
