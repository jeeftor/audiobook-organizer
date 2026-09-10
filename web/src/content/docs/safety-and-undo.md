---
title: "Safety And Undo"
description: "Dry-run, review, execute, and undo behavior."
---

Audiobook Organizer is designed around preview-first workflows. The safest path is dry-run, review, execute, then keep the undo log until the library is verified.

## Dry-Run Invariant

Dry-run mode must not mutate the filesystem:

```bash
audiobook-organizer --dir=/books/source --out=/books/organized --dry-run
```

```bash
audiobook-organizer rename --dir=/books/source --dry-run
```

Use dry-run output to inspect destination paths, skipped books, conflicts, and metadata warnings. You can preview a new output folder before it exists. Dry-run does not create output folders, remove empty folders, or restore files when combined with `--undo`.

## Destination Safety

Organization refuses to replace an occupied destination. If a move fails, the command reports the error and logs only completed moves. Resolve the reported conflict before retrying; do not delete a file unless you have verified which copy you want to keep.

Rename resolves collisions with numeric suffixes, including files already on disk. Identical names in different directories do not conflict. Selecting fewer rows in the web review preserves the original collision-resolved filenames while the source files and configuration are unchanged. If you change the files on disk, refresh the preview before running.

For rename templates, `--strict` rejects missing required simple fields before renaming any file. A fallback such as `{year|Unknown}` or an optional composite such as `{year - }` still permits a missing year.

## Organization Undo

Organization operations write `.abook-org.log`.

Later runs using the same log retain earlier recovery history. Undo reverses all
recorded operations, newest first. A malformed or unreadable existing log blocks
a new non-dry-run operation; preserve the log and resolve the error before retrying.

Undo with the same source and output directories used for the original run:

```bash
audiobook-organizer --dir=/books/source --out=/books/organized --undo
```

If an original path is occupied or a restore fails, undo reports the failure and retains the pending operations in the log. Successfully restored operations are removed from that log. Resolve the conflict and rerun the same undo command. Restores across filesystems use a copy-and-delete fallback.

Keep the log until you have verified the output folder and any Audiobookshelf scan results.

## Rename Undo

Rename operations write `.abook-rename.log`.

Successive runs in the same directory retain earlier rename history. If undo
partly succeeds, only failed operations remain in the log. Resolve the reported
conflict and retry; already-restored files are not replayed.

Undo from the renamed directory:

```bash
audiobook-organizer rename --dir=/books/source --undo
```

## Safer First Runs

1. Start with a small folder.
2. Use a separate `--out` directory.
3. Run with `--dry-run --verbose`.
4. Fix missing metadata before execution.
5. Run the real command.
6. Verify the output.
7. Keep the undo log until you no longer need rollback.

## Riskier Options

Use these only after the preview is understood:

| Option | Risk |
| --- | --- |
| In-place organization with no `--out` | Source folders change directly |
| `--remove-empty` | Empty source directories are removed after moves |
| Large recursive source directories | More skipped/error cases can be hidden in long output |
| Incorrect ABS path mapping | ABS metadata may point at paths the host cannot access |

See [Getting Started](/audiobook-organizer/getting-started/) for a safe first-run command sequence.
