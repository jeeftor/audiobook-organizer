# Audiobookshelf Test Harness

This directory provides resettable local Audiobookshelf instances for ABS integration work.

It follows the official Audiobookshelf Docker Compose shape:

- `ghcr.io/advplyr/audiobookshelf`
- plain server host port `13378` mapped to container port `80`
- metadata-enabled server host port `13379` mapped to container port `80`
- separate bind mounts for `/audiobooks`, `/books`, `/config`, and `/metadata`

The default compose image is the upstream GHCR image. Override it when needed:

```bash
ABS_IMAGE=audiobookshelf:local make abs-dev-up
```

## Layout

```text
test/abs/
  docker-compose.yml
  staging-data/
    audiobooks/  # durable local LibriVox M4B fixture cache
    books/       # durable local Project Gutenberg EPUB fixture cache
  runtime/
    plain/
      audiobooks/  # copied from staging-data, metadata.json stripped
      books/       # copied from staging-data, metadata.json stripped
    metadata/
      audiobooks/  # copied from staging-data, metadata.json written
      books/       # copied from staging-data, metadata.json written
  state/
    plain/
      config/    # ABS database and settings for the plain server
      metadata/  # ABS cache, covers, logs, backups for the plain server
    metadata-enabled/
      config/    # ABS database and settings for the metadata-enabled server
      metadata/  # ABS cache, covers, logs, backups for the metadata-enabled server
  scripts/
```

The media and live ABS state directories are intentionally ignored by Git so you can reset and reseed them repeatedly without growing the repository.

`baseline-config/` is different: it is an intentional test fixture. Commit those sanitized ABS databases and migration files when they define the root login, API key, library setup, tags, and server settings required by repeatable tests. `state/` is the local working copy restored from that baseline and then changed by scans; do not commit `state/`.

`staging-data/` is the durable local fixture cache. `runtime/` is throwaway data mounted into ABS. Resetting clears ABS state and rebuilds both runtime copies from `staging-data/` without downloading again.

The two ABS services are intentionally separate:

- `abs-plain` on <http://localhost:13378> uses `runtime/plain` and should keep "Store metadata with item" off.
- `abs-metadata` on <http://localhost:13379> uses `runtime/metadata` and should enable "Store metadata with item".

Within each ABS server, add two separate libraries:

- Audiobooks: `/audiobooks`
- Books: `/books`

## Test Workflow

The harness is designed to test whether Audiobook Organizer can update an ABS-backed library correctly.

1. Reset both ABS servers to the captured baseline.
2. Copy deliberately messy audiobook and EPUB files into each server's mounted library folders.
3. Trigger ABS scans through the ABS API so ABS records the bad starting paths.
4. Run Audiobook Organizer against those folders using ABS, embedded, or sidecar metadata.
5. Verify the organizer proposes and performs the expected moves.
6. Trigger ABS scans again so ABS observes the updated filesystem layout.

The runtime fixture paths are intentionally wrong, for example:

```text
runtime/plain/audiobooks/unsorted-audio/drop-001/not-alice.m4b
runtime/plain/audiobooks/loose/holiday_story_final.m4b
runtime/plain/books/imported/ebook-001.epub
runtime/plain/books/random/shelley-book.epub
runtime/plain/books/to-sort/austen.epub
```

The metadata-enabled runtime mirrors those bad paths and adds `metadata.json` sidecars next to the files. The plain runtime has no sidecars.

## Commands

Start Docker Desktop or another Docker daemon before running the compose targets.

Seed public-domain test media into `staging-data/`, then refresh both runtime copies:

```bash
make abs-dev-seed
```

Initialize clean ABS instances with empty mounted libraries for first-time account, setting, and library setup:

```bash
make abs-dev-init
```

Start Audiobookshelf:

```bash
make abs-dev-up
```

This removes orphaned containers from older harness layouts so stale containers do not keep ports like `13378` allocated.

Start the servers and wait for both to respond:

```bash
make abs-dev-wait
```

Open both servers, create the initial admin account, then add libraries:

- Plain: <http://localhost:13378>
- Metadata-enabled: <http://localhost:13379>

Use this local test login in both:

```text
username: root
password: password
```

For GitHub Actions or other disposable environments, use the committed baseline fixture by default. The API configuration script is still useful when creating or refreshing a baseline from empty ABS servers:

```bash
make abs-dev-configure
```

That initializes `root/password` if needed, sets the plain server to `storeMetadataWithItem=false`, sets the metadata-enabled server to `storeMetadataWithItem=true`, creates the `/audiobooks` and `/books` libraries, and writes tokens to the ignored ABS env file.

Capture the configured databases into the committed baseline config fixture:

```bash
ABS_PLAIN_TOKEN=... ABS_METADATA_TOKEN=... make abs-dev-capture-baseline
```

This writes:

- `test/abs/baseline-config/plain/config/`
- `test/abs/baseline-config/metadata-enabled/config/`
- `test/abs/.env.local`

Restore that baseline later:

```bash
make abs-dev-restore-baseline
```

Stop Audiobookshelf:

```bash
make abs-dev-down
```

Reset the ABS databases, settings, cache, metadata, and logs, then restore both runtime copies from `staging-data/`:

```bash
make abs-dev-reset
```

This restores the baseline configs into ignored `state/`, starts both ABS services, and waits for them to respond.

The reset script intentionally stops if Docker cannot stop the running ABS containers. Do not copy baseline SQLite files over a live ABS container; SQLite can keep the old database open and the UI/API may continue to report the old state.

Trigger scans for the configured `Audiobooks` and `Ebooks` libraries on both ABS instances:

```bash
make abs-dev-scan
```

The committed baseline DBs include the same active API key and signing secret on both ABS instances. The scan script loads `test/abs/.env.testing` by default and uses its single `ABS_TOKEN` for both servers. Set `ABS_ENV_FILE=test/abs/.env.local` when you intentionally want local overrides while developing or refreshing the baseline. Per-instance `ABS_PLAIN_TOKEN` and `ABS_METADATA_TOKEN` still override `ABS_TOKEN` when needed.

The scan script uses the ABS API:

- `POST /api/libraries/<ID>/scan?force=1`
- `GET /api/libraries/<ID>/items?limit=1`

It waits for expected counts by default:

- `/audiobooks`: `2`
- `/books`: `3`

Override with `ABS_EXPECT_AUDIOBOOKS`, `ABS_EXPECT_BOOKS`, or `ABS_SCAN_TIMEOUT` if the fixture set changes.

Reset and scan in one step:

```bash
make abs-dev-reset-scan
```

Run the CI-style smoke path using the committed baseline fixture:

```bash
make abs-ci-smoke
```

This seeds public-domain media, resets runtime/state, restores the committed baseline fixture into ignored `state/`, starts both ABS services, and scans both libraries.

Reset everything, including staged public-domain media:

```bash
make abs-dev-reset-all
```

## Test Data Sources

The seed script downloads public-domain fixtures:

- LibriVox M4B: Alice's Adventures in Wonderland (Abridged)
- LibriVox M4B: A Christmas Carol
- Project Gutenberg EPUB: Alice's Adventures in Wonderland
- Project Gutenberg EPUB: Frankenstein
- Project Gutenberg EPUB: Pride and Prejudice

The generated `metadata.json` files give Audiobook Organizer and ABS sidecar metadata to inspect in addition to embedded metadata.

You can also skip the downloader and manually place fixtures in:

```text
test/abs/staging-data/audiobooks/
test/abs/staging-data/books/
```

Then run `make abs-dev-reset` to rebuild both runtime libraries from those files.

## Useful Local Paths

For CLI and web API testing:

```text
Plain ABS URL:        http://localhost:13378
Metadata ABS URL:     http://localhost:13379
Plain ABS SQLite:     test/abs/state/plain/config/absdatabase.sqlite
Metadata ABS SQLite:  test/abs/state/metadata-enabled/config/absdatabase.sqlite
Plain audiobooks:     test/abs/runtime/plain/audiobooks
Plain books:          test/abs/runtime/plain/books
Metadata audiobooks:  test/abs/runtime/metadata/audiobooks
Metadata books:       test/abs/runtime/metadata/books
Container paths:
  /audiobooks
  /books
```

Example path mapping:

```bash
--abs-path-map="/audiobooks:$(pwd)/test/abs/runtime/plain/audiobooks"
```

Use the API token from the ABS user settings after creating the admin account.
