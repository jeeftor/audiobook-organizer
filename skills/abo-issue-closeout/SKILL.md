---
name: abo-issue-closeout
description: Finish Audiobook Organizer issue hygiene: verify completion, update changelog/docs/test notes, comment status, and close only when appropriate.
metadata:
  short-description: Close out ABO issues
---

# ABO Issue Closeout

You are the Audiobook Organizer issue closeout engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/pr.md`.

## Workflow

1. Run `$abo-issue-verify` logic first: issue criteria, branch diff, tests, docs, changelog, and ABS matrix.
2. Add missing `CHANGELOG.md`, docs, or `test/abs/test-matrix.md` updates when required.
3. Comment on the issue with what changed, tests run, and any follow-up work.
4. If a PR will close the issue, ensure the PR body uses `Resolves #<issue>` and do not manually close it.
5. Directly close only when the user explicitly asks, the issue is duplicate/obsolete, or the work intentionally completed without a PR.
6. If closing directly, include the reason and verification summary in the closing comment.

Do not close an issue with failing or unrun required checks unless the user explicitly accepts the risk and the reason is documented.
