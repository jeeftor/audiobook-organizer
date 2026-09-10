# Review Regression Coverage (#212)

Rows 1–18 correspond to the original project review; rows 19–22 cover recovery gaps found during the second acceptance pass and PR review.

| Finding | Corrected behavior | Regression coverage |
| --- | --- | --- |
| 1 | Organize refuses occupied destinations without replacing their bytes | `TestSafetyOrganizeRejectsOccupiedDestination`, `TestSafetyCopyFallbackPreservesOccupiedTargetAndSupportsUndo`, CLI `occupied_destination_and_failed_move_log` |
| 2 | Rename reserves no-op and occupied targets and rejects late collisions | `TestRenameSafetyReservesExistingAndGeneratedTargets`, `TestRenameSafetyLateCollisionDoesNotOverwrite`, CLI `rename_preserves_noop_target` |
| 3 | Undo with dry-run leaves files and logs unchanged | `TestSafetyUndoDryRunPreservesFilesAndLog`, CLI `dry_undo_and_retry_failed_undo` |
| 4 | Metadata separators cannot escape the output root, including through an existing destination symlink | `TestSafetyMetadataComponentsAndSymlinkContainment`, CLI `metadata_cannot_escape_output` |
| 5 | Flat organization moves only selected files | `TestSafetyFlatSelectionAndLayouts`, real browser flat-selection test |
| 6 | Move failures return errors and only completed files enter the undo log | `TestSafetyFailedMoveNeverAppearsInLogOrSummary`, `TestSafetyPartialMoveLogsOnlyCompletedFiles`, CLI collision test |
| 7 | Failed undo retains pending operations for retry; copy fallback is safe in both directions | `TestSafetyFailedUndoRetainsOnlyPendingMovesAndRetries`, `TestSafetyCopyFallbackPreservesOccupiedTargetAndSupportsUndo`, CLI undo retry |
| 8 | ABS pagination uses zero-based pages and terminates on incomplete responses | `TestReviewPaginationContract`, `TestReviewPaginationStopsOnIncompleteResponse`, live ABS R1 |
| 9 | Dry-run and declined empty-directory cleanup terminate without deleting directories | `TestSafetyDryRunNewOutputAndEmptyCleanup`, CLI `declined_cleanup_terminates` |
| 10 | Dry-run permits a nonexistent output without creating it | `TestSafetyDryRunNewOutputAndEmptyCleanup`, CLI new-output test, real browser new-output preview |
| 11 | SQLite uses the registered driver and actual ABS folder schema read-only | `TestNewPathMapperFromSQLite`, CLI `sqlite_discovery_from_committed_abs_database`, live ABS R2 |
| 12 | Single-file organization honors all seven supported layouts | `TestSafetyFlatSelectionAndLayouts` layout subtests |
| 13 | Selecting a rename subset preserves the reviewed collision suffix | `TestRenameSafetySelectionPreservesPreview`, real browser second-file selection test |
| 14 | Same filenames in separate directories are not conflicts | `TestRenameSafetySeparateDirectoriesDoNotConflict`, app rename tests, real browser rename test |
| 15 | ABS mappings use whole components and longest prefixes in both directions | `TestReviewPathMappingBoundary`, live ABS R3 |
| 16 | TUI worker state does not race with rendering; completion arrives through Update | `TestReviewProcessConcurrentView` under `-race`, verifying real file moves and final model state |
| 17 | Tagged integration suite compiles and is included in Make and CI verification | `make test-integration`, `make test-all`, CI tagged integration step |
| 18 | Strict rename validates required fields before changing files, while fallbacks and optional groups remain valid | `TestRenameSafetyStrictMissingFieldAbortsBeforeMutation`, `TestRenameSafetyStrictValidFileWaitsForWholePlanValidation`, CLI `strict_missing_year` |
| 19 | Separate organize and rename runs preserve earlier history; corrupt logs block new moves | CLI `organize_history_survives_separate_runs`, `rename_history_and_partial_undo_retry`, `invalid_recovery_log_blocks_new_moves` |
| 20 | Partial rename undo retains only pending entries and can be retried | CLI `rename_history_and_partial_undo_retry` |
| 21 | Blocked chained undo stops before older dependencies can move an unrelated occupant | `TestRecoveryChainedRenameUndoPreservesIntermediateOccupant`, `TestRecoveryChainedOrganizeUndoPreservesIntermediateOccupant`, CLI `chained_rename_undo_preserves_intermediate_occupant` |
| 22 | Log-publication failures roll back new directory, single-file, and rename moves; failed rollback reports remaining paths | `TestRecoveryLogFailureRollsBackOrganize`, `TestRecoveryLogFailureRollsBackRename`, `TestRecoveryRollbackFailureReportsRemainingMove`, CLI `log_write_failure_rolls_back_real_moves` |

Run the default suite, tagged integration suite, and race checks:

```bash
go test ./...
make test-integration
go test -race ./internal/organizer ./internal/app ./internal/server ./internal/tui/models
make lint
```

Browser tests use the real local server and filesystem fixtures:

```bash
cd web
npm exec -- playwright test --project=chromium-desktop
```

Live ABS regression tests create 101 additional EPUB copies in the isolated test
library, then verify 104 distinct items through the API and CLI. They read the
live SQLite database and check mapped paths against real files. Follow the reset
contract in `test/abs/test-matrix.md`; do not run concurrent ABS suites.

```bash
make abs-test-matrix ABS_TEST_RUN=TestABSReviewPaginationSQLiteAndMappings
```

Unit mocks are supplemental for ABS behavior. A successful live run is still
required before closing the ABS acceptance criteria.
