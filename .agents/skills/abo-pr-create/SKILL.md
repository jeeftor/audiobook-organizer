---
name: abo-pr-create
description: Commit, push, and create Audiobook Organizer pull requests into master after verification and PR body preparation.
metadata:
  short-description: Create ABO PRs
---

# ABO PR Create

You are the Audiobook Organizer PR release engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/pr.md`.

## Preconditions

- Do not create PRs from `master`, `main`, `develop`, or `dev`.
- Do not create PRs from branches without an appropriate work-type prefix: `feature/`, `fix/`, `docs/`, or `chore/`.
- Do not include unrelated dirty worktree changes.
- Do not create the PR with failing required checks unless the user explicitly accepts that status.

## Workflow

1. Inspect branch, status, staged/unstaged/untracked files, remotes, and upstream.
2. Verify the branch is tied to an issue and targets `master`.
3. Confirm or create the PR body using `$abo-pr-writer`.
4. Run or confirm relevant checks from `references/abo-assistant/testing.md`, including `prek run --all-files` when pre-commit hooks are configured.
5. Stage only intended files.
6. Re-check `git status --short --branch` before committing, and do not commit from `master`.
7. Commit with a concise message tied to the issue.
8. Re-check `git status --short --branch` before pushing, and do not push directly to `master`.
9. Push the branch and set upstream.
10. Create a draft PR with `gh pr create --base master --head <branch> --draft --body-file <file>`.
11. Verify the PR URL and draft state before reporting success.

End with the PR URL, tests run, and any checks still pending.
