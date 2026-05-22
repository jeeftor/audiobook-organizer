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
