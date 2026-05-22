# Explore Metadata

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

Interactive exploration:

```bash
audiobook-organizer metadata-tui --dir=/books/source
```

## Sources To Check

| Source | Use When |
| --- | --- |
| `metadata.json` | Each book folder already has sidecar metadata |
| Embedded EPUB/MP3/M4B | Files carry useful title, author, series, or track tags |
| Audiobookshelf | ABS already has cleaner metadata than the filesystem |
| Field mapping | Metadata exists but uses non-standard field names |

## How To Decide

Use the normal organize command first when every book directory has a useful `metadata.json` file:

```bash
audiobook-organizer --dir=/books/source --dry-run
```

Use embedded metadata when there are no sidecar files but the audio files have EPUB, MP3, or M4B tags:

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

See [Metadata Sources](METADATA.md) for extraction details and [Metadata Command](METADATA_COMMAND.md) for command reference.

## Next Steps

- [Organize Audiobooks](organize.md)
- [Rename Files](RENAME_FEATURE.md)
- [Audiobookshelf](audiobookshelf.md)
