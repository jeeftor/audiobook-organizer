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

Before organizing an Audiobookshelf-managed library, enable **Store metadata with item** in the Audiobookshelf library settings. That setting writes a `metadata.json` file beside each book when ABS metadata is generated or updated.

![Audiobookshelf setting for storing metadata.json files](store_metadata.jpg)

Those sidecar files are the safest first metadata source for local organization because they keep the book-level title, author, series, narrator, and year data beside the audio files. After a non-dry-run organization, run or trigger an Audiobookshelf library scan so ABS can reconcile moved paths.

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
