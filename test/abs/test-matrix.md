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

Custom layout template coverage belongs in focused organizer, REST, and browser
tests unless the ABS lifecycle itself is changing. `abs organize` accepts
`--layout-template` and routes it through the same organizer target-path logic as
the root organize command, so A5-A6 continue to cover the ABS reconciliation
lifecycle while non-Docker tests cover template rendering and path safety.

## Matrix

| ID | Mode | Instance | Library | Command shape | Expected result |
| --- | --- | --- | --- | --- | --- |
| H1 | Harness smoke | both | both | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSHarnessSmokeResetContract -count=1 -v` | Implemented. Runs `make abs-dev-reset-scan`, verifies both servers are reachable through the ABS API, asserts plain has `storeMetadataWithItem=0`, asserts metadata-enabled has `storeMetadataWithItem=1`, and verifies both instances have exactly 2 audiobook items, 3 ebook items, and 0 missing items after the initial scan. |
| M1A | `metadata.json` | metadata-enabled | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_AudiobooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. Moves audiobook directories using sidecar metadata; verifies old files are gone and new files exist; scans ABS; verifies old ABS rows are missing and new organized rows are active; calls the ABS library issues cleanup endpoint; rescans; verifies final ABS state has 2 active organized items and 0 missing items. |
| M1B | `metadata.json` | metadata-enabled | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_BooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. Moves ebook directories using sidecar metadata; verifies old files are gone and new files exist; scans ABS; verifies old ABS rows are missing and new organized rows are active; calls the ABS library issues cleanup endpoint; rescans; verifies final ABS state has 3 active organized items and 0 missing items. |
| M1C | `metadata.json` negative control | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_AudiobooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. No `metadata.json` exists, so no files move. Old messy paths remain; ABS scan leaves 2 active original items and 0 missing items. |
| M1D | `metadata.json` negative control | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode_BooksLifecycle -count=1 -v` | Implemented as a table-driven lifecycle row. No `metadata.json` exists, so no files move. Old messy paths remain; ABS scan leaves 3 active original items and 0 missing items. |
| M2A | Embedded metadata, already indexed | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedAlreadyIndexed_AudiobooksCurrentBehavior -count=1 -v` | Implemented as current-code coverage only. The already-indexed nested audiobook library has no `metadata.json` sidecars, and embedded mode currently finds `0` metadata rows and performs `0` moves. The test verifies the original messy paths stay active and ABS remains clean after rescan. Longer term, this workflow should use ABS metadata instead. |
| M2B | Embedded metadata, already indexed | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedAlreadyIndexed_BooksCurrentBehavior -count=1 -v` | Implemented as current-code coverage only. Moves already-indexed EPUB directories using embedded EPUB metadata; verifies old files are gone and organized author/title paths exist; scans ABS; verifies old ABS rows are missing and new organized rows are active; cleans missing rows; rescans; verifies final ABS state has 3 active organized items and 0 missing items. Longer term, this workflow should use ABS metadata instead. |
| M2C | Embedded import, hierarchical | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedMetadataImport_AudiobooksLifecycle -count=1 -v` | Implemented. Imports hierarchical M4B directories from `runtime/import-input/audiobooks` into the ABS-mounted audiobook library; verifies source files moved, organized author/title folders exist, ABS scan adds the imported items, and missing count remains `0`. |
| M2D | Embedded import, hierarchical | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestEmbeddedMetadataImport_BooksLifecycle -count=1 -v` | Implemented. Imports hierarchical EPUB directories from `runtime/import-input/books` into the ABS-mounted ebook library; verifies source files moved, organized author/title folders exist, ABS scan adds the imported items, and missing count remains `0`. |
| M3A | Flat mechanics, non-ABS output | flat input | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestFlatModeMechanics_NonABSOutput -count=1 -v` | Implemented as a table-driven mechanics row. Resets the ABS runtime fixture, processes loose MP3 flat-input files outside the mounted ABS libraries to a temporary non-ABS output directory, verifies source files moved out, verifies author/title output paths and `.abook-org.log`, and deliberately avoids ABS path assertions because output is outside the mounted library. |
| M3B | Flat mechanics, non-ABS output | plain source | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestFlatModeMechanics_NonABSOutput -count=1 -v` | Implemented as a table-driven mechanics row. Resets the ABS runtime fixture, processes already-mounted EPUB files in flat mode to a temporary non-ABS output directory, verifies source files moved out, verifies author/title output paths and `.abook-org.log`, and deliberately avoids ABS path assertions because output is outside the mounted library. |
| M3C | Flat import into ABS | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestFlatModeImport_AudiobooksLifecycle -count=1 -v` | Implemented. Imports loose MP3 files from `runtime/flat-input/audiobooks` into the ABS-mounted audiobook library; verifies per-file author/title folders, ABS scan adds the imported items, and missing count remains `0`. |
| M3D | Flat import into ABS | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestFlatModeImport_BooksLifecycle -count=1 -v` | Implemented. Reset now creates loose EPUB flat-input fixtures outside the mounted ABS libraries. The lifecycle imports them into the mounted ebook library, verifies per-file author/title folders, triggers an ABS scan, and verifies the imported paths are active with zero missing rows. |
| R1 | REST harness, `metadata.json` lifecycle | both | both | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_MetadataJSONModeLifecycle -count=1 -v` | Implemented. Runs the real Docker-backed ABS reset, scan, organizer move, missing detection, cleanup, rescan, and final-state checks from the `metadata.json` matrix rows, but drives organizer and ABS operations through the local web REST API. |
| R2 | REST harness, embedded import | plain | Audiobooks, Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_EmbeddedMetadataImportLifecycle -count=1 -v` | Implemented. Imports hierarchical M4B and EPUB fixtures from outside ABS into mounted ABS libraries through `/api/organize/run`, scans through REST, and verifies imported paths are active with zero missing rows. |
| R3 | REST harness, flat import | plain | Audiobooks, Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_FlatModeImportLifecycle -count=1 -v` | Implemented. Imports loose MP3 and EPUB fixtures from outside ABS into mounted ABS libraries through `/api/organize/run`, scans through REST, and verifies imported flat-mode paths are active with zero missing rows. |
| R4 | REST harness, web ABS setup endpoints | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_ABSSetupEndpoints -count=1 -v` | Implemented. Runs the real Docker-backed ABS reset, loads libraries through `/api/abs/libraries`, and validates the manual `/audiobooks` to host path mapping through `/api/abs/test-paths`. This covers the real API boundary used by the web ABS setup controls; browser UI contract coverage in `web/tests/e2e/gui-smoke.spec.ts` verifies URL/token testing unlocks the discovered-library selector before path validation. |
| R5 | REST harness, web ABS operation endpoints | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_ABSOperationEndpoints -count=1 -v` | Implemented. Runs the real Docker-backed ABS reset, loads mapped ABS metadata through `/api/abs/items`, reads active/missing state through `/api/abs/library-state`, triggers scans through `/api/abs/scan-trigger`, moves a fixture folder out of the mounted library, and removes the resulting missing row through `/api/abs/clean-missing`. This covers the real API boundary used by the web ABS operation controls; browser UI contract coverage remains in `web/tests/e2e/gui-smoke.spec.ts`. |
| R6 | REST harness, ABS metadata-source organize | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_ABSMetadataSourceOrganizeLifecycle -count=1 -v` | Implemented. The acceptance contract drives `/api/organize/preview` and `/api/organize/run` with `metadata_source: abs` and the validated ABS connection, verifies preview is non-mutating, verifies the run moves the mounted files using ABS metadata, then scans, cleans stale rows, rescans, and verifies the final ABS library state. |
| R7 | REST harness, ABS metadata-source rename preview | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_ABSRenameMetadataPreview -count=1 -v` | Implemented. Resets and scans real Docker-backed ABS, then drives `/api/rename/preview` with `metadata_source: abs`, a real API token, and a validated mapped library. It verifies every candidate receives non-empty ABS metadata without mutating the mounted files. |
| W1 | Browser UI, ABS setup and operations | plain | Audiobooks | `make gui-test-abs` | Implemented. Runs the real Docker-backed ABS reset, starts the real Go web UI, verifies an invalid API token leaves library selection, path validation, and review locked, then drives a valid URL/token through library discovery, path mapping validation, ABS item loading, library-state loading, scan triggering, missing-row detection after moving a mounted folder, destructive cleanup gating, cleanup, rescan, and final clean rendered state. |
| W2 | Browser UI, guided ABS metadata-source organize | plain | Audiobooks | `make gui-test-abs` | Implemented. Playwright enters **Guide Me**, selects Organize and Audiobookshelf API, then configures a real ABS connection and path mapping through the handed-off advanced controls. It inspects the non-mutating reviewed plan, runs the reviewed moves, and verifies the organized rendered result plus the required ABS scan-and-clean guidance against the Docker-backed fixture. |
| W3 | Browser UI, ABS metadata-source rename | plain | Audiobooks | `make gui-test-abs` | Implemented. Playwright will configure Rename with a real ABS token and path mapping, preview mapped ABS metadata, and execute a selected filename change with the normal confirmation and undo-log result. |
| A1 | ABS discovery | plain | both | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSMetadataMode_PreviewAndScanTrigger -count=1 -v` | Implemented. Lists both libraries and exercises discovery without moving files. |
| A2 | ABS manual path mapping | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSMetadataMode_PreviewAndScanTrigger -count=1 -v` | Implemented. Fetches ABS metadata, maps ABS paths to host paths, checks files, and verifies audiobook authors/titles appear in preview output. |
| A3 | ABS all-libraries preview | plain | both | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSMetadataMode_PreviewAndScanTrigger -count=1 -v` | Implemented. Confirms all-libraries mode loads both `/audiobooks` and `/books` with explicit path mappings. |
| A4 | ABS scan trigger | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSMetadataMode_PreviewAndScanTrigger -count=1 -v` | Implemented. Triggers a scan through the CLI and verifies the ABS API state remains clean. |
| A5 | ABS organize, already indexed | plain | Audiobooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSMetadataMode_OrganizeAudiobooksLifecycle -count=1 -v` | Implemented. Uses ABS metadata as the source of truth to move already-indexed audiobook folders when no `metadata.json` sidecars are present; verifies filesystem moves, organizer log, ABS missing/new rows after scan, missing cleanup, and final clean ABS state. |
| A6 | ABS organize, already indexed | plain | Ebooks | `go test -tags=abs_e2e ./test/abs/e2e -run TestABSMetadataMode_OrganizeBooksLifecycle -count=1 -v` | Implemented. Seeds explicit ABS author metadata through the ABS media-update API, then organizes already-indexed EPUB folders using ABS metadata and verifies the same filesystem, scan, cleanup, and final-state lifecycle as A5. |
| A7 | ABS organize, custom layout template | plain | Audiobooks | Covered by `go test ./cmd ./internal/app ./internal/server ./internal/organizer` and `npx playwright test tests/e2e/organize-real.spec.ts -g "custom layout template" --project chromium-desktop` | Implemented without a new Docker ABS lifecycle row. `abs organize` exposes `--layout-template` and maps it into the shared organizer config; focused command, REST, app, organizer, and real browser filesystem tests verify the custom target path behavior. A5-A6 continue to validate ABS scan/missing-row reconciliation. |

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
| `smoke-reset` | `TestABSHarnessSmokeResetContract` |
| `metadata-json` | `TestMetadataJSONMode` |
| `embedded-already-indexed` | `TestEmbeddedAlreadyIndexed` |
| `embedded-import` | `TestEmbeddedMetadataImport` |
| `flat-mechanics` | `TestFlatModeMechanics` |
| `flat-import` | `TestFlatModeImport` |
| `rest-metadata-json` | `TestRESTHarness_MetadataJSONModeLifecycle` |
| `rest-embedded-import` | `TestRESTHarness_EmbeddedMetadataImportLifecycle` |
| `rest-flat-import` | `TestRESTHarness_FlatModeImportLifecycle` |
| `rest-abs-setup` | `TestRESTHarness_ABSSetupEndpoints` |
| `rest-abs-operations` | `TestRESTHarness_ABSOperationEndpoints` |
| `rest-abs-metadata-organize` | `TestRESTHarness_ABSMetadataSourceOrganizeLifecycle` |
| `abs-metadata-mode` | `TestABSMetadataMode` |

The browser-backed ABS row `W1` runs in its own GitHub Actions job with
`make gui-test-abs` because it requires both Playwright-managed Chromium and the
Docker ABS harness.

Local `make abs-test-matrix` still runs all implemented rows serially. To run
one row locally, pass the same regex:

```bash
make abs-test-matrix ABS_TEST_RUN=TestRESTHarness_MetadataJSONModeLifecycle
```

To run the full REST-backed set locally:

```bash
make abs-test-rest
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

## ABS Metadata Organization

`audiobook-organizer abs scan` is the preview/connectivity command. It fetches
ABS metadata, maps paths, and prints proposed target paths.

`audiobook-organizer abs organize` is the execution path for files already
indexed by ABS:

```text
ABS API item metadata + ABS path mapping -> organizer target path -> filesystem move -> ABS scan
```

This mode is especially useful for the plain ABS instance, where no
`metadata.json` sidecars are stored with items. It should also work when embedded
metadata is missing, incomplete, or lower quality than ABS metadata.

SQLite path discovery is a separate implementation gap. The current committed
ABS baseline schema has `libraryFolders(path)`, while `internal/abs/path_mapper.go`
expects a newer or different `folders` table with `fullPath`. Until that code is
updated, use explicit `--abs-path-map` in harness tests.

## Initial Implementation Order

1. Add a smoke script that wraps `make abs-dev-reset-scan` and asserts counts and
   metadata settings.
2. Add a post-move scan helper that waits for expected count and zero missing
   items, and can optionally retry one forced scan.
3. Add metadata.json tests M1A-M1D.
4. Add embedded already-indexed coverage M2A-M2B only as current-code coverage. Implemented.
5. Add import fixtures and embedded hierarchical import tests M2C-M2D.
6. Add flat mechanics tests M3A-M3B with temporary non-ABS output. Implemented.
7. Add flat import fixtures and tests M3C-M3D. Implemented.
8. Add ABS preview tests A1-A4. Implemented.
9. Implement ABS organize mode A5-A6. Implemented.
10. Fix or replace SQLite path discovery for this ABS schema.
