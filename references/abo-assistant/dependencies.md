# Audiobook Organizer Dependency Reference

Use this reference for dependency audits and updates.

## Scope

Audit and update only dependency surfaces that exist on the current branch:

- Go module: `go.mod`, `go.sum`.
- Current web UI frontend: `web/package.json`, `web/package-lock.json` when present.
- GitHub Actions workflows: `.github/workflows/*.yml`.
- Docker base images and runtime tooling only when the task explicitly includes runtime/container updates.

Use only the dependency files for the current local browser UI.

## Read-Only Audit Commands

- Go dependency graph: `go list -m all`
- Available Go updates: `go list -m -u all`
- Go vulnerability scan: `govulncheck ./...`
- npm audit: `cd web && npm audit`
- npm outdated: `cd web && npm outdated`

If `govulncheck` is not installed, prefer `go install golang.org/x/vuln/cmd/govulncheck@latest` only after confirming network access is allowed. If network or certificates block the install, report the blocker.

## Update Commands

- Update one Go module: `go get example.com/module@latest`
- Update all direct and indirect Go modules: `go get -u ./...`
- Tidy after Go changes: `go mod tidy`
- Update one npm package: `cd web && npm install package@latest`
- Apply compatible npm fixes: `cd web && npm audit fix`

Avoid forceful or major-version npm updates unless the user explicitly accepts the risk.

## Verification After Updates

- Go-only dependency changes: `go mod tidy`, targeted tests if relevant, then `make test-unit` when practical.
- Web dependency changes: `make web-build`.
- Security fixes touching behavior: run affected tests plus `$abo-audit` style vulnerability checks again.
- Changelog is usually not required for invisible dependency bumps unless they fix a user-visible issue, security issue, runtime behavior, Docker/runtime dependency, or release artifact behavior.

## Reporting

For audits, separate:

- Confirmed vulnerabilities reachable in this repo.
- Vulnerabilities reported in dependencies but not used by this code path, when tool output supports that distinction.
- Outdated packages with no known vulnerability.
- Tooling or network blockers.

For updates, report exact modules/packages changed and verification results.
