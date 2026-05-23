---
title: "Explore Metadata"
description: "Inspect metadata before changing files."
---

Use metadata exploration before organizing or renaming when titles, authors, series, tracks, discs, or narrators are missing or inconsistent.

## Quick Inspection

Text output:

```bash
audiobook-organizer metadata --dir=/books/source
```

JSON output:

```bash
audiobook-organizer metadata --dir=/books/source --json
```

Pretty output:

```bash
audiobook-organizer metadata --dir=/books/source --pretty
```

Interactive exploration:

```bash
audiobook-organizer metadata-tui --dir=/books/source
```

## What The CLI Shows

Use `--pretty` when you want a readable terminal inspection before choosing an organize or rename strategy:

```bash
audiobook-organizer metadata --dir=/books/source --pretty
```

<section class="media-feature">
  <figure>
    <picture>
      <source srcset="/audiobook-organizer/assets/generated/cli/cli-metadata-inspect.webp" type="image/webp">
      <img src="/audiobook-organizer/assets/generated/cli/cli-metadata-inspect.png" alt="Verbose metadata command output showing source, title, author, series, track, album, and additional fields">
    </picture>
    <figcaption>Pretty metadata inspection before changing files</figcaption>
  </figure>
  <div>
    <p>The output shows the metadata source and the fields Audiobook Organizer can read before it plans any filesystem changes.</p>
    <p>Use this when a dry-run preview looks wrong, when files do not have <code>metadata.json</code>, or when you need to decide whether field mapping is required.</p>
    <p><code>--verbose</code> is kept as an alias for the same formatter-backed output.</p>
  </div>
</section>

## Sources To Check

| Source | Use When |
| --- | --- |
| `metadata.json` | Each book folder already has a `metadata.json` file |
| Embedded EPUB/MP3/M4B | Files carry useful title, author, series, or track tags |
| Audiobookshelf | ABS already has cleaner metadata than the filesystem |
| Field mapping | Metadata exists but uses non-standard field names |

## How To Decide

Use the normal organize command first when every book directory has a useful `metadata.json` file:

```bash
audiobook-organizer --dir=/books/source --dry-run
```

Use embedded metadata when there are no `metadata.json` files but the audio files have EPUB, MP3, or M4B tags:

```bash
audiobook-organizer --dir=/books/source --use-embedded-metadata --dry-run
```

Use flat mode when a folder contains individual audiobook files rather than one directory per book:

```bash
audiobook-organizer --dir=/books/source --flat --dry-run
```

Use field mapping when the preview finds metadata but puts the wrong values in author, title, series, track, or disc fields.

## Field Mapping

If metadata exists under custom fields, map it instead of editing every file first:

```bash
audiobook-organizer \
  --dir=/books/source \
  --title-field=book_title \
  --author-fields=authors,writer \
  --dry-run
```

See [Metadata Sources](/audiobook-organizer/metadata/) for extraction details and [Metadata Command](/audiobook-organizer/metadata-command/) for command reference.

## Next Steps

- [Organize Audiobooks](/audiobook-organizer/organize/)
- [Rename Files](/audiobook-organizer/rename/)
- [Audiobookshelf](/audiobook-organizer/audiobookshelf/)
