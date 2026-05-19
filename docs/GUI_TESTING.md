# GUI Testing

The GUI is the local Vue web UI served by `audiobook-organizer web`. GUI tests should exercise the real Go server whenever possible so authentication, embedded static assets, routing, and browser behavior are verified together.

## Test Layers

1. Go unit tests for `internal/app` and `internal/server`.
2. Frontend type/build checks with `npm run build` in `web/`.
3. Playwright E2E tests that start `go run . web --no-open`, parse the session token, and drive the rendered browser UI.
4. ABS-backed GUI workflows once the UI is wired to real ABS actions.

## Commands

```bash
# Install frontend dependencies
make web-install

# Install Playwright-managed Chromium for browser tests
cd web && npm run install:browsers

# Run Go REST endpoint tests without Docker or a browser
make gui-rest-test

# Build Vue assets and run Playwright headless tests
make gui-test

# Run the Docker-backed browser test for real ABS setup and operations
make gui-test-abs

# Run headed for local debugging
make gui-test-headed

# Open the Playwright UI runner
make gui-test-ui
```

CI runs the normal browser suite in the `Web UI Playwright` job. The job
installs frontend dependencies with `npm ci`, installs Playwright-managed
Chromium with Linux browser dependencies, and then runs `make gui-test`. On
failure it uploads the Playwright HTML report, traces, screenshots, and videos
from `web/playwright-report/` and `web/test-results/`.

CI runs the Docker-backed ABS browser suite separately in the `Web UI ABS
Playwright` job with `make gui-test-abs`. That job seeds the ABS fixtures,
resets both ABS containers through the normal harness reset contract, starts the
real Go web UI, and drives the browser through the real ABS setup and operation
endpoints.

Direct npm equivalents:

```bash
cd web
npm install
npm run install:browsers
npm run test:e2e
ABO_ABS_PLAYWRIGHT=1 npm run test:e2e -- --project=chromium-desktop tests/e2e/abs-real.spec.ts
npm run test:e2e:headed
npm run test:e2e:ui
```

## Browser Setup

The Playwright suite is configured to run both desktop and mobile projects with Playwright-managed Chromium. Install that browser payload from `web/`:

```bash
npm run install:browsers
```

On macOS, Playwright stores the managed browser revisions under `~/Library/Caches/ms-playwright`; on Linux CI runners, it uses the matching Playwright cache directory for that operating system. The suite should launch from that managed cache after `npm install` and `npm run install:browsers`.

If Playwright reports a missing executable such as `chromium_headless_shell-<rev>`, rerun `npm run install:browsers` from `web/`, then rerun `npm run test:e2e`. A separate Chrome Headless Shell under `~/.cache/puppeteer/chrome-headless-shell/` is useful for ad hoc manual rendering checks only; it is not the supported browser source for `npm run test:e2e`.

## Current Coverage

The current suite covers:

- REST tests for auth, config/options, static app serving, method validation,
  malformed JSON, organize preview, rename preview, and no-Docker ABS path
  mapping validation.
- The local Go web server starts with a generated session token and serves the
  embedded web app.
- Authenticated API endpoints reject missing tokens and accept the generated
  session token.
- The dashboard renders without browser console warnings or errors.
- Workflow navigation, backend bootstrap state, folder picker/drop behavior,
  narrow viewport overflow checks, and bootstrap fallback states.
- Real organize preview and execution against temporary filesystem fixtures.
- Real rename preview against temporary filesystem fixtures, including
  conflicts, skipped files, extraction errors, and the current deferred
  execution state.
- Mocked browser contract checks for ABS setup and operations.
- Real browser-driven ABS setup and operation coverage against the Docker ABS
  harness, including URL/token entry, library discovery, path mapping
  validation, metadata-item loading, library-state loading, scan triggering,
  destructive cleanup gating, and missing-item cleanup.

## Expansion Plan

As the GUI moves from scaffold to real workflows, add tests in this order:

1. Real source/output configuration state.
2. Organize preview API calls with fixture directories.
3. Rename preview API calls with fixture files.
4. Browser-driven ABS organize/import flows after ABS metadata setup.
5. Accessibility checks for keyboard navigation and visible focus states.
6. Visual regression snapshots for the main desktop and mobile layouts.
