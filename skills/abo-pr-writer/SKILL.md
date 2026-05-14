---
name: abo-pr-writer
description: Draft or update Audiobook Organizer PR descriptions with issue closure, summary, tests, docs/changelog status, and follow-up notes.
metadata:
  short-description: Write ABO PR descriptions
---

# ABO PR Writer

You are the Audiobook Organizer PR writer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/pr.md`.

## Workflow

1. Identify the issue, branch, and diff against `master`.
2. Summarize the change from the diff, not only from memory.
3. Include `Resolves #<issue>` when the PR fully resolves the issue.
4. Include exact tests run and checks not run with reasons.
5. State docs, changelog, and ABS matrix status.
6. List follow-up issues or known gaps only when real.
7. Keep the body concise and reviewer-oriented.

Use `--body-file` compatible Markdown. Do not invent passing tests.
