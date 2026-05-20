---
name: abo-pr
description: Route Audiobook Organizer pull request work including PR body drafting, creation, CI/review watching, and closeout checks.
metadata:
  short-description: Route ABO PR work
---

# ABO PR

You are the Audiobook Organizer PR coordinator.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/pr.md`.

## Route

- Draft or update PR body: `$abo-pr-writer`.
- Commit, push, and open a PR into `master`: `$abo-pr-create`.
- Existing PR status, comments, checks, or requested changes: `$abo-pr-watcher`.
- Issue completion checks before PR: `$abo-issue-verify`.
- Final issue/changelog/docs hygiene: `$abo-issue-closeout`.
- Finished feature, fix, docs, or chore branch: verify checks, get the PR ready, enable auto-merge or merge back into protected `master`, confirm the issue closed, and clean up the branch or worktree.
- After a PR merge closes tracked work, use `references/abo-assistant/common.md` next-work recommendation guidance before ending the response. Suggest the next issue or closeout step, but do not start it without user direction.

If intent is unclear, ask whether the user wants PR text, PR creation, PR status watching, or closeout verification.
