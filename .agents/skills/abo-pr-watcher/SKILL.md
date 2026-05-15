---
name: abo-pr-watcher
description: Watch Audiobook Organizer PR status, CI checks, review comments, issue comments, and branch freshness, then apply focused fixes when appropriate.
metadata:
  short-description: Watch ABO PRs
---

# ABO PR Watcher

You are the Audiobook Organizer PR watcher.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/pr.md`.

## Workflow

1. Identify the PR from the branch, prompt, URL, or PR number.
2. Inspect PR state, checks, review comments, issue comments, and branch freshness.
3. Classify findings as CI failure, requested change, maintainer question, docs/changelog gap, stale branch, or follow-up.
4. Inspect failed logs with `gh run view --log-failed` when available.
5. Before applying fixes, confirm `git status --short --branch` shows the PR branch and not `master`.
6. Apply mechanical or clearly requested fixes; ask before behavior changes or risky rebases.
7. Run the narrowest relevant checks after fixes.
8. When checks are green, inspect `mergeStateStatus` and `reviewDecision`.
9. If the only blocker is required review and repository auto-merge is available, run `gh pr merge <number> --auto --squash --delete-branch` and report that approval is pending.
10. When checks and review are clean, report whether the PR is merging or ready to merge back into `master` and whether the linked issue will close.
11. Summarize what changed, what remains blocked, and what should rerun.

Do not dismiss failures as flaky without evidence.
Do not bypass protected `master` or admin enforcement for normal work.
