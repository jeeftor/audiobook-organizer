# Getting Started

This guide walks through a safe first organization run. It uses a separate output directory so you can inspect the result before changing your source library.

## Install

Install from Homebrew, Go, Docker, or a release archive:

```bash
brew tap jeeftor/tap
brew install audiobook-organizer
```

```bash
go install github.com/jeeftor/audiobook-organizer@latest
```

See [Installation](INSTALLATION.md) for platform-specific options and package notes.

## If You Use Audiobookshelf

<section class="media-callout">
  <a class="media-callout-image" href="https://github.com/jeeftor/audiobook-organizer/blob/master/docs/store_metadata.jpg" target="_blank" rel="noopener">
    <img src="store_metadata.jpg" alt="Audiobookshelf setting for storing metadata.json files">
  </a>
  <div class="media-callout-copy">
    <p>Before organizing an Audiobookshelf-managed library, enable <strong>Store metadata with item</strong> in the Audiobookshelf library settings.</p>
    <p>That setting writes a <code>metadata.json</code> file beside each book when ABS metadata is generated or updated.</p>
    <p>Those sidecar files are the safest first metadata source for local organization because they keep the book-level title, author, series, narrator, and year data beside the audio files.</p>
  </div>
</section>

The normal ABS cycle is: configure sidecar metadata, preview, organize, scan ABS, clean up old missing paths if ABS still reports them, and keep the undo log until the library is verified.

![Audiobookshelf organize lifecycle](abs-organize-lifecycle.svg)

See [Audiobookshelf](audiobookshelf.md) for cleanup screenshots, path mapping checks, and scan commands.

## Choose A Small Source Folder

Start with a small sample from your own library:

```text
/books/source/
  The Case of Charles Dexter Ward/
    metadata.json
    01 - Chapter 1.mp3
```

Use a separate output folder for the first run:

```text
/books/organized/
```

If the source folder does not have `metadata.json` files, inspect embedded metadata before organizing:

```bash
audiobook-organizer metadata --dir=/books/source --use-embedded-metadata
```

For a flat folder of individual audiobook files, preview with `--flat` instead of the default directory-as-book behavior.

## Preview First

Run a dry run before moving files:

```bash
audiobook-organizer \
  --dir=/books/source \
  --out=/books/organized \
  --dry-run \
  --verbose
```

Check the planned destination paths, metadata warnings, and skipped books. Dry-run mode does not mutate the filesystem.

## Run For Real

When the preview looks right, remove `--dry-run`:

```bash
audiobook-organizer \
  --dir=/books/source \
  --out=/books/organized
```

The organizer writes `.abook-org.log` so the operation can be undone.

## Undo If Needed

Undo the last organization operation from the source directory:

```bash
audiobook-organizer --dir=/books/source --undo
```

Rename operations use a separate `.abook-rename.log` file:

```bash
audiobook-organizer rename --dir=/books/source --undo
```

See [Safety And Undo](safety-and-undo.md) for the invariants and log behavior.

## Next Steps

- Use [Layouts](LAYOUTS.md) to choose a directory structure.
- Use [Metadata Sources](METADATA.md) if metadata is missing or stored in custom fields.
- Use [Rename](RENAME_FEATURE.md) to standardize filenames after metadata is correct.
- Use [Audiobookshelf](audiobookshelf.md) when your source of truth is an ABS server.
