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
5. Choose a branch name with the correct prefix: `feature/<short-name>`, `fix/<short-name>`, `docs/<short-name>`, or `chore/<short-name>`.
6. Create the dedicated branch from `origin/master` after `git fetch origin master`. Do not branch from a dirty feature branch unless the user explicitly asks.
7. Verify the active branch with `git status --short --branch` before editing.
8. If the current checkout is dirty with unrelated work, use a separate `git worktree` for the new issue branch.
9. When creating a separate worktree and hook config exists, run `prek install --hook-type pre-commit --hook-type commit-msg` inside that worktree.
10. End with the issue number, branch name, and smallest verifiable goal.

Prefer issue titles that describe the user-visible or maintainer-visible outcome. Keep acceptance criteria concrete enough for `$abo-issue-verify`.
