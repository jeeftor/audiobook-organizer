<section class="doc-hero product-hero">
  <div class="hero-copy">
    <h1>Audiobook Organizer for Audiobookshelf (and More)</h1>
    <p class="lead">Clean up audiobook folders with dry-run previews, metadata-aware layouts, rename templates, undo logs, and Audiobookshelf workflows.</p>
    <div class="hero-actions">
      <a class="button" href="getting-started.md">Start safely</a>
      <a class="button secondary" href="interfaces.md">Choose workflow</a>
    </div>
  </div>
  <figure class="hero-media">
    <picture>
      <source srcset="assets/generated/web-ui/web-ui-metadata-json-preview.webp" type="image/webp">
      <img src="assets/generated/web-ui/web-ui-metadata-json-preview.png" alt="Local web UI showing a metadata.json organize preview" width="1440" height="1200" loading="eager" decoding="async" fetchpriority="high">
    </picture>
    <figcaption>Browser review before files move</figcaption>
  </figure>
</section>

Audiobook Organizer is for audiobook libraries that have drifted into inconsistent folder names, filenames, metadata sources, or Audiobookshelf paths. It lets you inspect what it can read, preview what it would change, run the operation, and keep an undo path while you verify the result.

<section class="visual-grid" aria-label="Generated workflow demos">
  <figure>
    <img src="assets/generated/cli/cli-organize-run.gif" alt="Animated CLI organize run showing a full-cycle non-dry-run workflow" width="1400" height="760" loading="lazy" decoding="async" fetchpriority="low">
    <figcaption>Full-cycle CLI organization</figcaption>
  </figure>
  <figure>
    <img src="assets/generated/tui/tui-organize-preview.gif" alt="Animated TUI organize preview workflow" width="1440" height="860" loading="lazy" decoding="async" fetchpriority="low">
    <figcaption>Interactive terminal preview</figcaption>
  </figure>
  <figure>
    <img src="assets/generated/cli/cli-rename-preview.gif" alt="Animated CLI rename preview workflow" width="1400" height="480" loading="lazy" decoding="async" fetchpriority="low">
    <figcaption>Template rename preview</figcaption>
  </figure>
</section>

## Common Tasks

| Task | Start With |
| --- | --- |
| Organize books into `Author/Series/Title` | [Organize](organize.md) |
| Rename files from title, author, series, track, or disc fields | [Rename Files](RENAME_FEATURE.md) |
| Use Audiobookshelf-created `metadata.json` files | [Audiobookshelf](audiobookshelf.md) |
| No `metadata.json`, but audio files have tags | [Explore Metadata](explore-metadata.md) |
| MP3 tags use non-standard fields | [Metadata Sources](METADATA.md#field-mapping) |
| Previous organization or rename needs to be reverted | [Safety And Undo](safety-and-undo.md) |
| Planned paths look wrong | [Troubleshooting](troubleshooting.md) |

## Audiobookshelf Users

If you use Audiobookshelf, enable **Store metadata with item** before the first organize run. This makes ABS save a `metadata.json` file beside each book, and Audiobook Organizer reads that file when it previews folder changes.

![Audiobookshelf setting for storing metadata.json files](store_metadata.jpg)

After a non-dry-run organization, trigger an Audiobookshelf scan and clean up any stale missing-book entries if ABS still points at old paths. See [Audiobookshelf](audiobookshelf.md) for the full setup and cleanup flow.

## What It Does

<section class="capability-grid" aria-label="Audiobook Organizer capabilities">
  <article>
    <h3>Organize</h3>
    <p>Move or copy books into layouts such as <code>Author/Series/Title</code>, including custom layout templates.</p>
  </article>
  <article>
    <h3>Rename Files</h3>
    <p>Build filenames from title, author, series, track, disc, narrator, and other metadata fields.</p>
  </article>
  <article>
    <h3>Explore Metadata</h3>
    <p>Read <code>metadata.json</code>, embedded EPUB/MP3/M4B metadata, and Audiobookshelf metadata before changing files.</p>
  </article>
  <article>
    <h3>Work safely</h3>
    <p>Use dry-run previews, reviewed execution, skip/error summaries, and undo logs for organization and rename runs.</p>
  </article>
  <article>
    <h3>Coordinate with ABS</h3>
    <p>Discover Audiobookshelf libraries, test container-to-host path mappings, organize from ABS metadata, and trigger scans.</p>
  </article>
  <article>
    <h3>Pick your interface</h3>
    <p>Use the local web UI, scriptable CLI, keyboard-first TUI, rename flows, metadata tools, or ABS command group.</p>
  </article>
</section>

## Start Here

| Goal | Start Here |
| --- | --- |
| Install the binary | [Installation](INSTALLATION.md) |
| Make a safe first run | [Getting Started](getting-started.md) |
| Pick web UI, CLI, TUI, or ABS | [Choose An Interface](interfaces.md) |
| Organize audiobooks | [Organize](organize.md) |
| Rename files from metadata templates | [Rename Files](RENAME_FEATURE.md) |
| Explore metadata before changing files | [Explore Metadata](explore-metadata.md) |
| Configure Audiobookshelf `metadata.json` files | [Audiobookshelf](audiobookshelf.md) |
| Use the browser workflow | [Local Web UI](GUI.md) |
| Work interactively in a terminal | [TUI](TUI.md) |
| Organize with Audiobookshelf metadata | [Audiobookshelf](audiobookshelf.md) |
| Understand dry-run and undo | [Safety And Undo](safety-and-undo.md) |
| Choose metadata sources and mappings | [Metadata Sources](METADATA.md) |
| Configure folder layouts | [Layouts](LAYOUTS.md) |
| Configure defaults and environment variables | [Configuration](CONFIGURATION.md) |
| See what changed between releases | [Changelog](../CHANGELOG.md) |
| Troubleshoot a preview, rename, ABS path, or browser issue | [Troubleshooting](troubleshooting.md) |

## First-Run Path

<ol class="workflow-list">
  <li><strong>Install</strong><span>Get the single <code>audiobook-organizer</code> binary.</span></li>
  <li><strong>Choose</strong><span>Use web UI, CLI, TUI, rename, metadata, or ABS workflows.</span></li>
  <li><strong>Preview</strong><span>Run a dry-run against a small source folder.</span></li>
  <li><strong>Review</strong><span>Check planned paths and metadata warnings.</span></li>
  <li><strong>Run</strong><span>Execute once the plan looks right.</span></li>
  <li><strong>Undo</strong><span>Keep the generated log until you are satisfied.</span></li>
</ol>

See [Getting Started](getting-started.md) for exact commands.
