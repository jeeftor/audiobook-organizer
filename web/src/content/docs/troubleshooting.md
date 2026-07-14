---
title: "Troubleshooting"
description: "Troubleshoot previews, renames, Audiobookshelf paths, and browser issues."
---

## Dry Run Shows Missing Metadata

Check the metadata source:

- `metadata.json` should live beside the book files.
- Embedded metadata requires `--use-embedded-metadata` or `--flat`.
- Non-standard fields may need mapping flags such as `--title-field` or `--author-fields`.

Use the metadata command to inspect what the organizer can read:

```bash
audiobook-organizer metadata --dir=/books/source
```

## Planned Paths Look Wrong

Check the selected layout:

```bash
audiobook-organizer --dir=/books/source --layout=author-series-title --dry-run
```

For custom templates, inspect the field reference:

```bash
audiobook-organizer layout-template
```

See [Layouts](/audiobook-organizer/layouts/).

## Rename Conflicts

Run rename in dry-run mode and inspect duplicate targets:

```bash
audiobook-organizer rename --dir=/books/source --dry-run
```

Use a template with enough unique fields, such as track or disc number.

## Audiobookshelf Cannot Find Files

Validate container-to-host path mapping:

```bash
audiobook-organizer abs scan \
  --abs-url=http://localhost:13378 \
  --abs-token="$ABS_TOKEN" \
  --abs-path-map="/audiobooks:/mnt/media/audiobooks" \
  --dir=/mnt/media/audiobooks \
  --check-files
```

See [Audiobookshelf](/audiobook-organizer/audiobookshelf/).

## Browser Does Not Open

Use `--no-open` and copy the printed local URL:

```bash
audiobook-organizer web --no-open
```

The server binds to `127.0.0.1` by default and uses a temporary token. Open the complete startup URL, including its `?token=...` parameter; otherwise the UI cannot call its API.

## Docs Visual Generation Fails

See [Docs Visuals](/audiobook-organizer/development/docs-visuals/). On macOS, TUI VHS captures run through the local Docker image to avoid native Chrome crash dialogs.
