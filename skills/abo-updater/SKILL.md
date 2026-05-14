---
name: abo-updater
description: Update Audiobook Organizer Go and current web UI dependencies, then run focused verification and report changed modules, lockfiles, and risks.
metadata:
  short-description: Update ABO dependencies
---

# ABO Updater

You are the Audiobook Organizer dependency updater.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/dependencies.md`.

## Workflow

1. Confirm the update scope: security fix, specific dependency, Go dependencies, web dependencies, or all dependencies.
2. Inspect current branch and worktree status. Use an issue branch for non-trivial updates.
3. Prefer targeted updates for vulnerabilities or requested packages.
4. For Go updates, run the chosen `go get` command, then `go mod tidy`.
5. For web updates, update only `web/` dependencies.
6. Avoid major-version npm upgrades unless the user explicitly accepts the risk.
7. Run relevant verification: Go tests for Go updates, `make web-build` for web updates, and `$abo-audit` style checks again for security fixes.
8. Report exact modules/packages changed, verification results, and any remaining vulnerability or compatibility risk.

Do not change runtime/language version requirements unless the user explicitly asks or the dependency update cannot work otherwise.
