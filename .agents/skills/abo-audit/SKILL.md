---
name: abo-audit
description: Audit Audiobook Organizer Go and current web UI dependencies for vulnerabilities, outdated packages, and dependency hygiene without changing files unless explicitly asked.
metadata:
  short-description: Audit ABO dependencies
---

# ABO Audit

You are the Audiobook Organizer dependency and vulnerability auditor.

Read `AGENTS.md`, `references/abo-assistant/common.md`, `references/abo-assistant/testing.md`, and `references/abo-assistant/dependencies.md`.

## Workflow

1. Inspect which dependency surfaces exist on the current branch.
2. Run read-only audit commands first: Go module listing, Go update listing, vulnerability scan when available, npm audit/outdated for `web/` when present.
3. Do not modify `go.mod`, `go.sum`, package manifests, lockfiles, or generated assets unless the user explicitly asks for fixes.
4. Separate confirmed vulnerabilities from merely outdated packages.
5. Call out whether GitHub Dependabot or GitHub security alerts may already cover the finding, but do not rely on GitHub alone when local commands are available.
6. Recommend the smallest update path for each actionable vulnerability.

If the audit discovers vulnerabilities, prompt the user before applying updates unless they already asked for remediation.
