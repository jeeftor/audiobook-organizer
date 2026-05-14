---
name: abo-abs-tests
description: Work on Audiobook Organizer Audiobookshelf harness validation, ABS E2E tests, matrix updates, reset contracts, and ABS-facing behavior verification.
metadata:
  short-description: Test ABO ABS behavior
---

# ABO ABS Tests

You are the Audiobook Organizer Audiobookshelf validation engineer.

Read `AGENTS.md`, `references/abo-assistant/common.md`, and `references/abo-assistant/testing.md`.

## Workflow

1. Determine whether the change affects ABS discovery, path mapping, metadata mode, scan triggering, import/organize behavior, mounted-library behavior, or ABS-facing web/API flows.
2. Update `test/abs/test-matrix.md` for new or changed behavior before considering implementation complete.
3. Preserve the reset contract: stop containers, rebuild runtime fixtures, restore committed baseline config, start containers, and scan.
4. Never restore SQLite state while ABS containers are running.
5. Prefer focused `go test -tags=abs_e2e ./test/abs/e2e -run TestName -count=1 -v`, then `make abs-test-matrix` when practical.
6. If Docker, downloads, or certificates block validation, report the exact command and blocker.

ABS tests should verify command result, filesystem result, and ABS API/database result when relevant.
