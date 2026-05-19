---
name: abo-workflow
description: Route Audiobook Organizer maintainer work across issue, implementation, testing, web UI, ABS, and PR skills. Use for broad repo workflow requests or when the user asks which Audiobook Organizer skill should handle a task.
metadata:
  short-description: Route Audiobook Organizer work
---

# ABO Workflow

You are the Audiobook Organizer workflow coordinator.

Read `AGENTS.md` and `references/abo-assistant/common.md`.

## Intake

When the user asks to start, pick, or work through repository issues but does not name a specific issue or task, ask whether they want to:

- create a new feature/fix idea, then route through `$abo-issue-create`;
- choose from the existing GitHub issue list, then route through `$abo-issue-watcher`.

If they choose existing issues, inspect open issues first and help pick one before creating a branch or worktree. If they choose new work, capture the goal, motivation, and acceptance criteria before creating the issue and branch.

## Route

- New tracked work or branch setup: `$abo-issue-create`.
- Focused feature implementation: `$abo-feature`.
- Bug reproduction and fixes: `$abo-bugfix`.
- Existing issue status, labels, comments, or next steps: `$abo-issue-watcher`.
- Verify issue acceptance criteria before closeout or PR: `$abo-issue-verify`.
- Closeout hygiene, changelog/docs/test status, and issue final update: `$abo-issue-closeout`.
- Test selection, new tests, and verification plans: `$abo-tests`.
- Audiobookshelf harness or ABS-facing behavior: `$abo-abs-tests`.
- Current local browser UI work in `web/`, `internal/server`, or `internal/app`: `$abo-web-ui`.
- Documentation, AGENTS.md, repo-local skills, or changelog-only work: `$abo-docs`.
- PR drafting, creation, or watching: `$abo-pr`.

If intent is clear, apply the relevant specialist workflow directly in the current agent. If the user explicitly asks for subagents or parallel work, delegate bounded tasks to specialist subagents.

## Closeout

When a routed task completes, use `references/abo-assistant/common.md` next-work recommendation guidance before ending the response. Recommend the next issue, closeout step, or parallel-safe pairing based on the current open issue list and dependency chain, but wait for the user before starting new work.
