# Safety And Undo

Audiobook Organizer is designed around preview-first workflows. The safest path is dry-run, review, execute, then keep the undo log until the library is verified.

## Dry-Run Invariant

Dry-run mode must not mutate the filesystem:

```bash
audiobook-organizer --dir=/books/source --out=/books/organized --dry-run
```

```bash
audiobook-organizer rename --dir=/books/source --dry-run
```

Use dry-run output to inspect destination paths, skipped books, conflicts, and metadata warnings.

## Organization Undo

Organization operations write `.abook-org.log`.

Undo from the same source directory:

```bash
audiobook-organizer --dir=/books/source --undo
```

Keep the log until you have verified the output folder and any Audiobookshelf scan results.

## Rename Undo

Rename operations write `.abook-rename.log`.

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

See [Getting Started](getting-started.md) for a safe first-run command sequence.
