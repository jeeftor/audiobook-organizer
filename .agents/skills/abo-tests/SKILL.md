---
name: abo-tests
description: Select, write, and run Audiobook Organizer verification for Go CLI, organizer, TUI, server/app, web build, documentation, and release hygiene changes.
metadata:
  short-description: Test ABO changes
---

# ABO Tests

You are the Audiobook Organizer test engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/testing.md`.

## Workflow

1. Identify the behavior under test and affected package boundary.
2. Find nearby tests before adding new ones.
3. Before adding or modifying tests, confirm `git status --short --branch` shows a dedicated non-`master` issue branch.
4. For bug fixes, create or identify a failing check first when practical.
5. Prefer focused package tests, then widen to repo-native checks.
6. Run `prek run --all-files` when pre-commit hooks are configured.
7. For web/browser tests, verify the required browser binary before treating failures as product failures:
   - Playwright E2E uses managed browser payloads under `~/Library/Caches/ms-playwright`.
   - If Playwright reports a missing executable such as `chromium_headless_shell-<rev>`, run `npx playwright install chromium` from `web/` and rerun the failing check.
   - A separate cached Chrome Headless Shell may exist under `~/.cache/puppeteer/chrome-headless-shell/`; use it for ad hoc rendered smoke checks when Playwright-managed browsers are unavailable, but do not treat that as a replacement for fixing the Playwright cache.
   - If browser launch fails because of macOS sandbox/permission errors, rerun the same browser check with escalation rather than downgrading to build-only verification.
8. Keep tests user-visible and behavior-oriented; avoid implementation-detail assertions.
9. Report exact commands, status, browser binary path used, and blockers.

Route ABS harness work to `$abo-abs-tests` and current local browser UI work to `$abo-web-ui`.
