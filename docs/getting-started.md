# Getting Started

This guide walks through a safe first organization run. It uses a separate output directory so you can inspect the result before changing your source library.

## Audiobookshelf First-Run Flow

For Audiobookshelf-managed libraries, use this cycle: configure sidecar metadata, preview, organize, scan ABS, clean up old missing paths if ABS still reports them, and keep the undo log until the library is verified.

![Audiobookshelf organize lifecycle](abs-organize-lifecycle.svg)

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

## 1. Configure Audiobookshelf

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

## 2. Preview With Dry Run

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

Run a dry run before moving files:

```bash
audiobook-organizer \
  --dir=/books/source \
  --out=/books/organized \
  --dry-run \
  --verbose
```

Check the planned destination paths, metadata warnings, and skipped books. Dry-run mode does not mutate the filesystem.

## 3. Organize

When the preview looks right, remove `--dry-run`:

```bash
audiobook-organizer \
  --dir=/books/source \
  --out=/books/organized
```

The organizer writes `.abook-org.log` so the operation can be undone.

## 4. Trigger Audiobookshelf Scan

After a non-dry-run organization, run or trigger an Audiobookshelf library scan so ABS can discover the organized paths and reconcile moved files.

See [Audiobookshelf](audiobookshelf.md) for path mapping checks and scan commands.

## 5. Clean Missing Items

After the scan, ABS may still list old filesystem paths as missing. That is a normal cleanup step after moving files, not a failed organize run. Review the missing items in ABS, then use the missing-books cleanup action when the organized paths are already visible in the library.

<section class="image-pair" aria-label="Audiobookshelf missing item cleanup screenshots">
  <figure>
    <a href="https://github.com/jeeftor/audiobook-organizer/blob/master/docs/issues.jpg" target="_blank" rel="noopener">
      <img src="issues.jpg" alt="Audiobookshelf issues view showing missing books">
    </a>
    <figcaption>Review missing old paths in the ABS Issues view.</figcaption>
  </figure>
  <figure>
    <a href="https://github.com/jeeftor/audiobook-organizer/blob/master/docs/remove_books.jpg" target="_blank" rel="noopener">
      <img src="remove_books.jpg" alt="Audiobookshelf remove missing books action">
    </a>
    <figcaption>Remove missing entries after ABS has found the organized files.</figcaption>
  </figure>
</section>

## 6. Verify And Keep The Undo Log

Check the organized library in your filesystem and in Audiobookshelf before deleting the undo log. Keep `.abook-org.log` until you are satisfied with the result.

If needed, undo the last organization operation from the source directory:

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
