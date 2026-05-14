# ABS Test Matrix

This document defines the planned Audiobookshelf integration test coverage for
Audiobook Organizer.

## GitHub Tracking

The automated matrix work is tracked by parent issue
[#27](https://github.com/jeeftor/audiobook-organizer/issues/27) and these
sub-issues:

| Issue | Area | Matrix rows |
| --- | --- | --- |
| [#30](https://github.com/jeeftor/audiobook-organizer/issues/30) | Smoke and reset contract | H1 |
| [#31](https://github.com/jeeftor/audiobook-organizer/issues/31) | `metadata.json` lifecycle | M1A-M1D |
| [#37](https://github.com/jeeftor/audiobook-organizer/issues/37) | `metadata.json` ebook lifecycle | M1B, M1D |
| [#32](https://github.com/jeeftor/audiobook-organizer/issues/32) | Embedded metadata lifecycle | M2A-M2D |
| [#39](https://github.com/jeeftor/audiobook-organizer/issues/39) | Embedded already-indexed audiobook current behavior | M2A |
| [#35](https://github.com/jeeftor/audiobook-organizer/issues/35) | Embedded audiobook import lifecycle | M2C |
| [#38](https://github.com/jeeftor/audiobook-organizer/issues/38) | Embedded ebook import lifecycle | M2D |
| [#28](https://github.com/jeeftor/audiobook-organizer/issues/28) | Flat mode lifecycle | M3A-M3D |
| [#36](https://github.com/jeeftor/audiobook-organizer/issues/36) | Flat audiobook import lifecycle | M3C |
| [#29](https://github.com/jeeftor/audiobook-organizer/issues/29) | ABS API metadata mode | A1-A6 |

## Reset Contract

Each destructive test must start from the same ABS state:

1. Stop both ABS containers.
2. Rebuild `test/abs/runtime/**` from `test/abs/staging-data/**`.
3. Restore committed `test/abs/baseline-config/**` into ignored
   `test/abs/state/**`.
4. Start both ABS containers.
5. Trigger an initial ABS scan so ABS records the deliberately messy starting
   paths.

The current command for that is:

```bash
make abs-dev-reset-scan
```

`baseline-config/**` is the committed fixture. `state/**` is the ignored local
copy that ABS mutates during scans and tests.

### Optional Warm Reset

If scan time becomes a problem, add a later cache step:

1. Run the full reset and initial scan once.
2. Stop ABS.
3. Copy both `runtime/**` and `state/**` to an ignored
   `initial-scanned-state/` cache.
4. Before each test, stop ABS and restore both trees from that cache.

That cache must include both filesystem runtime data and SQLite state. Caching
only the database is not enough because organizer tests move files.

## Current Fixtures

The reset scripts create two ABS instances and two libraries per instance:

| Instance | URL | Library | Mounted path | Metadata setting |
| --- | --- | --- | --- | --- |
| plain | `http://localhost:13378` | Audiobooks | `/audiobooks` | `storeMetadataWithItem=false` |
| plain | `http://localhost:13378` | Ebooks | `/books` | `storeMetadataWithItem=false` |
| metadata-enabled | `http://localhost:13379` | Audiobooks | `/audiobooks` | `storeMetadataWithItem=true` |
| metadata-enabled | `http://localhost:13379` | Ebooks | `/books` | `storeMetadataWithItem=true` |

Initial expected item counts after reset and scan:

| Library path | Count |
| --- | ---: |
| `/audiobooks` | 2 |
| `/books` | 3 |

The plain runtime has no `metadata.json` sidecars. The metadata-enabled runtime
has `metadata.json` sidecars next to the messy fixture files.

Reset also rebuilds ignored import-only source folders that are not mounted into
ABS:

| Source path | Fixture source | Purpose |
| --- | --- | --- |
| `test/abs/runtime/import-input/audiobooks` | committed `testdata/m4b` files | Hierarchical embedded metadata import into ABS. |
| `test/abs/runtime/import-input/books` | committed `testdata/epub` files | Hierarchical embedded metadata ebook import into ABS. |
| `test/abs/runtime/flat-input/audiobooks` | committed `testdata/mp3flat` files | Loose-file flat import into ABS. |

## Test Axes

Primary axes:

- Organizer metadata source: `metadata.json`, embedded metadata, ABS API
  metadata.
- Placement workflow: reorganize files already indexed by ABS, or import new
  files into an ABS-mounted library.
- File layout workflow: hierarchical directories or flat loose files.
- ABS instance: plain versus metadata-enabled.
- Library type: audiobook files versus EPUB files.
- Operation shape: dry-run preview, filesystem move, ABS rescan/update.

The intended source-of-truth split is:

- Files already indexed by ABS should eventually use ABS metadata. ABS knows the
  current item path, user edits, tags, series, matches, and any metadata that is
  not present on disk.
- `metadata.json` mode is for ABS-style sidecar metadata when the sidecars are
  present on disk.
- Embedded metadata mode is mainly for importing files that are not yet indexed
  by ABS, both hierarchical and flat.

Default layout for integration tests should be `author-title` unless a test is
specifically about series layout. It gives stable, easy-to-assert paths:

```bash
--layout author-title
```

## Matrix

| ID | Mode | Instance | Library | Command shape | Expected result |
| --- | --- | --- | --- | --- | --- |
| H1 | Harness smoke | both | both | `make abs-dev-reset-scan` | Both servers reachable; plain has metadata setting `0`; metadata-enabled has metadata setting `1`; both have 2 audiobook items and 3 ebook items. |
| M1A | `metadata.json` | metadata-enabled | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_AudiobooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. Moves audiobook directories using sidecar metadata; verifies old files are gone and new files exist; scans ABS; verifies old ABS rows are missing and new organized rows are active; calls the ABS library issues cleanup endpoint; rescans; verifies final ABS state has 2 active organized items and 0 missing items. |
| M1B | `metadata.json` | metadata-enabled | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_BooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. Moves ebook directories using sidecar metadata; verifies old files are gone and new files exist; scans ABS; verifies old ABS rows are missing and new organized rows are active; calls the ABS library issues cleanup endpoint; rescans; verifies final ABS state has 3 active organized items and 0 missing items. |
| M1C | `metadata.json` negative control | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_AudiobooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. No `metadata.json` exists, so no files move. Old messy paths remain; ABS scan leaves 2 active original items and 0 missing items. |
| M1D | `metadata.json` negative control | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_BooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. No `metadata.json` exists, so no files move. Old messy paths remain; ABS scan leaves 3 active original items and 0 missing items. |
| M2A | Embedded metadata, already indexed | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedAlreadyIndexed_AudiobooksCurrentBehavior -count=1 -v` | Implemented as current-code coverage only. The already-indexed nested audiobook library has no `metadata.json` sidecars, and embedded mode currently finds `0` metadata rows and performs `0` moves. The test verifies the original messy paths stay active and ABS remains clean after rescan. Longer term, this workflow should use ABS metadata instead. |
| M2B | Embedded metadata, already indexed | plain | Ebooks | `go run . --dir test/abs/runtime/plain/books --use-embedded-metadata --layout author-title` | Current-code coverage only. Moves already-indexed EPUB directories using embedded EPUB metadata. Longer term, this workflow should use ABS metadata instead. |
| M2C | Embedded import, hierarchical | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedMetadataImport_AudiobooksLifecycle -count=1 -v` | Implemented. Imports hierarchical M4B directories from `runtime/import-input/audiobooks` into the ABS-mounted audiobook library; verifies source files moved, organized author/title folders exist, ABS scan adds the imported items, and missing count remains `0`. |
| M2D | Embedded import, hierarchical | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedMetadataImport_BooksLifecycle -count=1 -v` | Implemented. Imports hierarchical EPUB directories from `runtime/import-input/books` into the ABS-mounted ebook library; verifies source files moved, organized author/title folders exist, ABS scan adds the imported items, and missing count remains `0`. |
| M3A | Flat mechanics, non-ABS output | plain source | Audiobooks | `go run . --dir test/abs/runtime/plain/audiobooks --out <tmp>/flat-audiobooks --flat --layout author-title` | Processes supported files individually across nested messy folders and writes organized files to a temporary output directory. This proves flat mechanics, but does not test ABS path updates because output is outside the mounted ABS library. |
| M3B | Flat mechanics, non-ABS output | plain source | Ebooks | `go run . --dir test/abs/runtime/plain/books --out <tmp>/flat-books --flat --layout author-title` | Processes loose EPUB files individually across nested messy folders and writes organized files to a temporary output directory. This proves flat mechanics, but does not test ABS path updates. |
| M3C | Flat import into ABS | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestFlatModeImport_AudiobooksLifecycle -count=1 -v` | Implemented. Imports loose MP3 files from `runtime/flat-input/audiobooks` into the ABS-mounted audiobook library; verifies per-file author/title folders, ABS scan adds the imported items, and missing count remains `0`. |
| M3D | Flat import into ABS | plain | Ebooks | `go run . --dir test/abs/runtime/flat-input/books --out test/abs/runtime/plain/books --flat --layout author-title` | New fixture needed. Imports loose EPUB files from outside ABS into the ABS-mounted ebook library. ABS scan should add imported items. |
| R1 | REST harness, `metadata.json` lifecycle | both | both | `make abs-test-rest` | Implemented. Runs the real Docker-backed ABS reset, scan, organizer move, missing detection, cleanup, rescan, and final-state checks from the `metadata.json` matrix rows, but drives organizer and ABS operations through the local web REST API. |
| A1 | ABS discovery | plain | both | `go run . abs scan --abs-url http://localhost:13378 --abs-token <token>` | Works today. Lists both libraries and item counts. Does not move files. |
| A2 | ABS manual path mapping | plain | Audiobooks | `go run . abs scan --abs-url http://localhost:13378 --abs-token <token> --abs-library Audiobooks --abs-path-map "/audiobooks:<abs>/test/abs/runtime/plain/audiobooks" --dir <abs>/test/abs/runtime/plain/audiobooks --check-files` | Works today as preview/connectivity coverage. It fetches ABS metadata, maps ABS paths to host paths, checks files, and calculates target paths. It does not perform organization. |
| A3 | ABS all-libraries preview | plain | both | `go run . abs scan --abs-url http://localhost:13378 --abs-token <token> --abs-all-libraries --abs-path-map "/audiobooks:<abs>/test/abs/runtime/plain/audiobooks" --abs-path-map "/books:<abs>/test/abs/runtime/plain/books" --dir <abs>/test/abs/runtime/plain --check-files` | Works today if all-libraries mode handles both mappings. Confirms ABS metadata can be loaded across both libraries. Does not move files. |
| A4 | ABS scan trigger | plain | Audiobooks | `go run . abs scan-trigger --abs-url http://localhost:13378 --abs-token <token> --abs-library <id>` | Works today. Confirms organizer CLI can trigger ABS scan; detailed scan completion can still use `scan-libraries.sh` polling. |
| A5 | ABS organize, already indexed | plain | Audiobooks | Future: `go run . abs organize --abs-url http://localhost:13378 --abs-token <token> --abs-library Audiobooks --abs-path-map "/audiobooks:<abs>/test/abs/runtime/plain/audiobooks" --dir <abs>/test/abs/runtime/plain/audiobooks --layout author-title` | Future behavior. Uses ABS metadata as the source of truth to move already-indexed files even when no `metadata.json` sidecars exist and embedded metadata is absent, sparse, or wrong. |
| A6 | ABS organize, already indexed | plain | Ebooks | Future: same as A5 for `Ebooks` and `/books` | Future behavior. Confirms ABS metadata can drive organization of an already-indexed ebook library. |

## Per-Test Verification

Each mode test should verify three layers:

1. Process result: command exits `0`; no unexpected stderr; log file exists after
   non-dry-run organizer moves.
2. Filesystem result: expected old messy paths are gone or still present for
   negative controls; expected organized paths exist.
3. ABS result: trigger a post-move scan; API or SQLite shows expected item count
   and no duplicate/missing items.

Useful checks:

```bash
sqlite3 test/abs/state/plain/config/absdatabase.sqlite \
  "select count(*) from libraryItems where libraryId = '<library-id>';"

sqlite3 test/abs/state/plain/config/absdatabase.sqlite \
  "select path from libraryItems order by path;"

sqlite3 test/abs/state/plain/config/absdatabase.sqlite \
  "select count(*) from libraryItems where isMissing = 1;"
```

Prefer API assertions where possible, but SQLite assertions are acceptable for
path-level checks because the harness already owns the ABS database fixture.

## GitHub Matrix

GitHub Actions runs the implemented ABS matrix as parallel job rows. Each row
gets its own runner, Docker daemon, ABS containers, fixture restore, and scan
cycle, so the fixed ABS ports do not conflict:

| Row | `ABS_TEST_RUN` |
| --- | --- |
| `metadata-json` | `TestMetadataJSONMode` |
| `embedded-already-indexed` | `TestEmbeddedAlreadyIndexed` |
| `embedded-import` | `TestEmbeddedMetadataImport` |
| `flat-import` | `TestFlatModeImport` |
| `rest-metadata-json` | `TestRESTHarness_MetadataJSONModeLifecycle` |

Local `make abs-test-matrix` still runs all implemented rows serially. To run
one row locally, pass the same regex:

```bash
make abs-test-matrix ABS_TEST_RUN=TestRESTHarness_MetadataJSONModeLifecycle
```

## Reset Per Test

Recommended runner shape:

```text
for each test:
  make abs-dev-reset-scan
  run organizer command
  trigger ABS scan for touched library or libraries
  wait for expected counts and zero missing items
  assert filesystem paths
  assert ABS item paths/counts
```

The runner should stop on the first failed reset. Never restore SQLite files
while ABS containers are running.

## Post-Move ABS Scan Contract

After organizer moves files that ABS already indexed, the ABS database may need
one or more scans to reconcile old and new paths. The test should not assume item
IDs remain stable until we prove ABS behavior for this version.

Post-move assertions should wait until:

- expected item count is present for the touched library,
- `isMissing = 0` for that library,
- all library item paths are under the expected organized path prefix,
- old messy paths are absent,
- mapped host files exist.

If a first scan leaves missing items while also adding the new paths, run a
second forced scan before failing. If the second scan still has missing items,
the test should fail and print old/new `libraryItems.path`, `relPath`, and
`isMissing` values.

This scan reconciliation is central to what we need to validate: organizer moves
must be followed by ABS scans that resolve missing files and update the ABS view
of the library.

## Embedded Metadata Import Workflow

Embedded metadata should primarily cover files that are not yet in ABS:

1. Source files start outside mounted ABS libraries, for example
   `test/abs/runtime/import-input/**`.
2. Organizer reads embedded metadata and writes organized output into
   `test/abs/runtime/plain/audiobooks` or `test/abs/runtime/plain/books`.
3. ABS scan discovers new items.
4. Assertions verify item count increases or the expected imported paths appear.

This applies to both hierarchical embedded mode and flat mode. Once a file is
already indexed in ABS, the preferred future metadata source is ABS itself.

## Flat Mode Import Fixtures

Flat mode import uses loose files under `test/abs/runtime/flat-input/**`, which
is rebuilt from committed `testdata` fixtures during reset and is not mounted
into ABS. The organizer writes output into the ABS-mounted library root with
`--out`.

If `--out` is omitted, flat mode organizes each file relative to that file's
current bad directory. If `--out` is the ABS library root while the source is
already under that same root, flat mode skips the source because it avoids
processing files inside its output directory.

Current audiobook flat import source:

```text
test/abs/runtime/flat-input/audiobooks/
```

Future ebook flat import source:

```text
test/abs/runtime/flat-input/books/
```

Flat import tests use `--out` to place organized files into `/audiobooks` or
`/books`, trigger ABS scan, and assert new ABS items. Because the source files
were outside ABS before the organizer run, these tests should not create missing
ABS rows.

## ABS Mode Gaps

The current `audiobook-organizer abs scan` command is preview/connectivity-only.
It fetches ABS metadata, maps paths, and prints proposed target paths, but it
does not yet move files through the organizer core.

The desired ABS mode is a new execution path for files already indexed by ABS:

```text
ABS API item metadata + ABS path mapping -> organizer target path -> filesystem move -> ABS scan
```

This mode is especially useful for the plain ABS instance, where no
`metadata.json` sidecars are stored with items. It should also work when embedded
metadata is missing, incomplete, or lower quality than ABS metadata.

Until ABS organization is implemented, ABS-mode tests should cover:

- authentication and discovery,
- manual path mapping,
- library item fetch and file existence checks,
- scan trigger.

SQLite path discovery is a separate implementation gap. The current committed
ABS baseline schema has `libraryFolders(path)`, while `internal/abs/path_mapper.go`
expects a newer or different `folders` table with `fullPath`. Until that code is
updated, use explicit `--abs-path-map` in harness tests.

Once ABS mode can perform organization, add tests mirroring M1/M2:

- ABS metadata as source of truth,
- move files in the mounted runtime library,
- trigger ABS scan,
- assert ABS item paths changed to the organizer target paths.
It should be tested against the plain instance first because that instance does
not write `metadata.json` sidecars, forcing the command to use ABS metadata.

## Initial Implementation Order

1. Add a smoke script that wraps `make abs-dev-reset-scan` and asserts counts and
   metadata settings.
2. Add a post-move scan helper that waits for expected count and zero missing
   items, and can optionally retry one forced scan.
3. Add metadata.json tests M1A-M1D.
4. Add embedded already-indexed coverage M2A-M2B only as current-code coverage. M2A is implemented.
5. Add import fixtures and embedded hierarchical import tests M2C-M2D.
6. Add flat mechanics tests M3A-M3B with temporary non-ABS output.
7. Add flat import fixtures and tests M3C-M3D.
8. Add ABS preview tests A1-A4.
9. Stub or implement ABS organize mode A5-A6.
10. Fix or replace SQLite path discovery for this ABS schema.
