---
name: abo-issue-closeout
description: Close out Audiobook Organizer issues with verification, status comments, and PR-aware hygiene.
metadata:
  short-description: Close out ABO issues
---

# ABO Issue Closeout

You are the Audiobook Organizer issue closeout engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/pr.md`.

## Workflow

1. Run `$abo-issue-verify` logic first: issue criteria, branch diff, tests, docs, changelog, and ABS matrix.
2. Before adding missing files, confirm `git status --short --branch` shows the dedicated non-`master` issue branch.
3. Add missing `CHANGELOG.md`, docs, or `test/abs/test-matrix.md` updates when required. For functionality or workflow changes, missing docs include the relevant `README.md` overview, static docs site page under `web/src/content/docs/`, and mirrored root `docs/` page when present.
4. Confirm real E2E acceptance evidence for user-facing workflow changes; do not close on mocked/stubbed tests alone unless the maintainer explicitly accepted and the issue comment documents the gap.
5. Classify the issue as maintainer-created, user-originated, or unclear using `references/abo-assistant/common.md`.
6. For user-originated or unclear issues, do not auto-close when reporter confirmation or manual interaction is needed. Comment with what changed, tests run, the expected reporter validation, and the next action; close only after reporter confirmation, explicit maintainer approval, obsolescence, or duplication is documented.
7. Comment on the issue with what changed, tests run, and any follow-up work.
8. If a PR will close the issue and closeout is allowed, ensure the PR body uses `Resolves #<issue>` and do not manually close it. For user-originated issues that need reporter confirmation, avoid closing keywords until confirmation or maintainer approval is documented.
9. Treat the issue as open until the resolving PR is ready, required checks pass, auto-merge is enabled or the PR has merged back into `master`, and the linked issue has closed or the reporter-confirmation block is documented.
10. After merge, confirm the linked issue closed when closure was intended and the feature branch or worktree was cleaned up. Repository delete-branch-on-merge should remove the remote branch, but verify it.
11. Directly close only when the user explicitly asks, the issue is duplicate/obsolete, the work intentionally completed without a PR, or reporter confirmation/maintainer approval is documented.
12. If closing directly, include the reason and verification summary in the closing comment.
13. After issue closeout, follow `references/abo-assistant/common.md` next-work recommendation guidance and tell the user the best next item to pick. Do not start the next item unless the user asks.

Do not close an issue with failing or unrun required checks unless the user explicitly accepts the risk and the reason is documented.
