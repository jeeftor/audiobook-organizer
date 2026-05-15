# Audiobook Organizer PR Reference

Use this shared reference for PR writing, creation, watching, and closeout.

## PR Preconditions

- Branch is not `master`, `main`, `develop`, or `dev`.
- Branch uses the appropriate work-type prefix: `feature/`, `fix/`, `docs/`, or `chore/`.
- Branch is tied to a GitHub issue.
- Unrelated dirty worktree changes are not included.
- Relevant tests, lint, and builds have been run or explicitly documented as blocked.
- Pre-commit hooks have been run with `prek run --all-files` when hook config exists, or their absence is documented.
- User-visible changes have a `CHANGELOG.md` entry under `Unreleased`.
- ABS-facing changes update `test/abs/test-matrix.md` when relevant.

## PR Body Shape

Include:

- `Resolves #<issue>`
- Summary.
- Tests run, with exact commands.
- Docs/changelog status.
- Follow-up issues or known gaps, if any.

Prefer concise reviewer-oriented text. Do not hide unrun tests.

## PR Commands

- `gh pr view --json number,title,state,url,baseRefName,headRefName,isDraft,mergeStateStatus,statusCheckRollup`
- `gh pr checks <number>`
- `gh pr diff <number>`
- `gh pr create --base master --head <branch> --draft --title "<title>" --body-file <file>`

Prefer `--body-file` over `--fill` so issue links, test notes, and changelog status are preserved.
Prefer Squash and merge when the PR is ready unless the maintainer asks for another merge strategy.

## Closeout Rules

- Issues normally close through PR merge back into `master`.
- A branch is not done when implementation is committed, pushed, or opened as a draft PR. Closeout means the PR is ready, required checks and review are satisfied, the PR is merged, the linked issue closes, and stale branches or worktrees are cleaned up.
- Use closing keywords in the PR body when the PR fully resolves the issue.
- Directly close an issue only when the user explicitly asks, the issue is obsolete/duplicate, or the work intentionally completed outside PR merge.
- Before closeout, verify acceptance criteria against code, tests, docs, changelog, and PR state.

## Watcher Duties

1. Inspect PR status, checks, review comments, issue comments, and branch freshness.
2. Classify findings as CI failure, requested change, maintainer question, docs/changelog gap, stale branch, or follow-up.
3. Apply mechanical or clearly requested fixes.
4. Ask before risky rebases, behavior changes, or ambiguous maintainer feedback.
5. Summarize what changed, what remains blocked, and which checks should rerun.
