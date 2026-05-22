# Organize Audiobooks

Organize is the main Audiobook Organizer workflow. It previews and then moves or copies audiobook folders into consistent layouts using sidecar metadata, embedded metadata, or Audiobookshelf metadata.

## Safe Organize Flow

1. Choose a small source folder.
2. Choose an output folder for the first run.
3. Run a dry-run preview.
4. Review planned destinations and warnings.
5. Run the operation.
6. Keep `.abook-org.log` until the result is verified.

```bash
audiobook-organizer \
  --dir=/books/source \
  --out=/books/organized \
  --dry-run \
  --verbose
```

When the plan is correct:

```bash
audiobook-organizer \
  --dir=/books/source \
  --out=/books/organized
```

## Common Layouts

Use a built-in layout:

```bash
audiobook-organizer --dir=/books/source --layout=author-series-title --dry-run
```

Use a custom layout template:

```bash
audiobook-organizer \
  --dir=/books/source \
  --layout-template="{author}/{series}/{series-count} - {title}" \
  --dry-run
```

See [Layouts](LAYOUTS.md) for built-in layouts and template fields.

## Metadata Sources

Organize can read:

- `metadata.json`
- embedded EPUB metadata
- embedded MP3 tags
- embedded M4B metadata
- Audiobookshelf metadata

See [Explore Metadata](explore-metadata.md) when the preview shows missing or incorrect fields.

## Interfaces

- Browser workflow: [Local Web UI](GUI.md)
- Scriptable workflow: [CLI](CLI.md)
- Interactive terminal workflow: [TUI](TUI.md)
- ABS metadata workflow: [Audiobookshelf](audiobookshelf.md)
