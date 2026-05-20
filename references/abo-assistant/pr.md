# Audiobook Organizer PR Reference

Use this shared reference for PR writing, creation, watching, and closeout.

## PR Preconditions

- Branch is not `master`, `main`, `develop`, or `dev`.
- Branch uses the appropriate work-type prefix: `feature/`, `fix/`, `docs/`, or `chore/`.
- Branch is tied to a GitHub issue.
- Issue origin has been classified as maintainer-created, user-originated, or unclear using `references/abo-assistant/common.md`.
- Unrelated dirty worktree changes are not included.
- Relevant tests, lint, and builds have been run or explicitly documented as blocked.
- User-facing workflow changes have real E2E acceptance evidence; mocked UI/API tests are supplemental only unless the maintainer explicitly accepted the documented gap.
- User-originated or unclear issues that need reporter confirmation or manual interaction have a documented confirmation path before the PR uses closing keywords.
- Pre-commit hooks have been run with `prek run --all-files` when hook config exists, or their absence is documented.
- User-visible changes have a `CHANGELOG.md` entry under `Unreleased`.
- ABS-facing changes update `test/abs/test-matrix.md` when relevant.
- `master` branch protection requires all configured checks to pass before merge. Repository auto-merge is enabled for the single-maintainer workflow, so a separate approving review is not required unless branch protection is intentionally changed.

## PR Body Shape

Include:

- `Resolves #<issue>`
- Summary.
- Tests run, with exact commands.
- Real E2E evidence for user-facing workflow changes, or the maintainer-accepted reason it is blocked.
- Issue origin and reporter-confirmation status when the issue is user-originated or unclear.
- Docs/changelog status.
- Follow-up issues or known gaps, if any.

Prefer concise reviewer-oriented text. Do not hide unrun tests.
Use closing keywords for maintainer-created issues when the PR fully resolves them. For user-originated or unclear issues that need reporter confirmation or manual interaction, use a non-closing issue reference until confirmation or explicit maintainer approval is documented.

## PR Commands

- `gh pr view --json number,title,state,url,baseRefName,headRefName,isDraft,mergeStateStatus,statusCheckRollup`
- `gh pr checks <number>`
- `gh pr diff <number>`
- `gh pr create --base master --head <branch> --title "<title>" --body-file <file>`
- `gh pr merge <number> --auto --squash --delete-branch`

Prefer `--body-file` over `--fill` so issue links, test notes, and changelog status are preserved.
Prefer Squash and merge when the PR is ready unless the maintainer asks for another merge strategy.
Prefer enabling auto-merge once checks are green. If GitHub reports `REVIEW_REQUIRED`, report that branch protection is out of sync with the single-maintainer workflow; do not bypass branch protection.

## Closeout Rules

- Issues normally close through PR merge back into `master`.
- A branch is not done when implementation is committed, pushed, or opened as a PR. Closeout means the PR is ready, required checks pass, auto-merge is enabled or the PR merges, the linked issue closes, and stale branches or worktrees are cleaned up.
- Use closing keywords in the PR body when the PR fully resolves a maintainer-created issue.
- For user-originated or unclear issues, preserve reporter validation when needed: do not auto-close through PR keywords until reporter confirmation, maintainer approval to close without it, or documented obsolescence/duplication exists.
- Directly close an issue only when the user explicitly asks, the issue is obsolete/duplicate, or the work intentionally completed outside PR merge.
- Before closeout, verify acceptance criteria against code, tests, docs, changelog, and PR state.
- Do not close a user-facing workflow issue based only on mocked/stubbed tests. Confirm real E2E evidence or documented maintainer acceptance of the gap.
- After closeout, use the next-work recommendation guidance in `references/abo-assistant/common.md` before ending the user-facing response. Suggest what to do next, but do not start it without user direction.

## Watcher Duties

1. Inspect PR status, checks, review comments, issue comments, and branch freshness.
2. Classify findings as CI failure, requested change, maintainer question, docs/changelog gap, stale branch, or follow-up.
3. Apply mechanical or clearly requested fixes.
4. Ask before risky rebases, behavior changes, or ambiguous maintainer feedback.
5. If required checks are green and the PR is otherwise mergeable, enable auto-merge with squash/delete-branch when available.
6. Summarize what changed, what remains blocked, which checks should rerun, and the next recommended work item when the PR closes an issue.
