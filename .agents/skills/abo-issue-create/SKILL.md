---
name: abo-issue-create
description: Create or reuse GitHub issues for Audiobook Organizer work, then prepare the correct issue branch from master. Use before non-trivial code or documentation edits in this repository.
metadata:
  short-description: Create ABO issues and branches
---

# ABO Issue Create

You are the Audiobook Organizer issue starter.

Read `AGENTS.md` and `references/abo-assistant/common.md`.

## Workflow

1. Inspect branch, worktree status, remotes, and default branch availability.
2. Search existing issues with `gh issue list --state all --search "<keywords>"`.
3. Reuse a matching open issue when one exists; otherwise create one with goal, motivation, and acceptance criteria.
4. Comment on the issue when scope or branch setup needs to be recorded.
5. Create a dedicated branch from `master` or `origin/master`. Do not branch from a dirty feature branch unless the user explicitly asks.
6. If the current checkout is dirty with unrelated work, use a separate `git worktree` for the new issue branch.
7. End with the issue number, branch name, and smallest verifiable goal.

Prefer issue titles that describe the user-visible or maintainer-visible outcome. Keep acceptance criteria concrete enough for `$abo-issue-verify`.
