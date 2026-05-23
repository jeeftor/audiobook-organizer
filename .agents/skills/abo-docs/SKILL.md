---
name: abo-docs
description: Update Audiobook Organizer documentation, AGENTS.md, changelog entries, and maintainer-facing workflow notes while keeping repo-local skill references consistent.
metadata:
  short-description: Maintain ABO docs
---

# ABO Docs

You are the Audiobook Organizer documentation maintainer.

Read `AGENTS.md` and `references/abo-assistant/common.md`.

## Workflow

1. Identify whether the change affects user docs, maintainer workflow, repo-local skills, `CHANGELOG.md`, or PR text.
2. Confirm `git status --short --branch` shows a dedicated non-`master` issue branch before editing non-trivial docs.
3. For functionality or workflow changes, evaluate both user-facing documentation surfaces:
   - Root `README.md`: keep it as the compact GitHub landing page with current capabilities, safe first-run guidance, and links to deeper docs.
   - Static docs site source under `web/src/content/docs/`: update the full workflow, reference, troubleshooting, or generated-visual page that owns the detailed guidance.
4. When the same page also exists under root `docs/`, keep the `docs/` Markdown and `web/src/content/docs/` page synchronized unless the task explicitly changes that mirroring model.
5. Keep `AGENTS.md` focused on durable repo rules and route repeatable procedures into `.agents/skills/abo-*` or `references/abo-assistant/`.
6. Update `CHANGELOG.md` under `Unreleased` for user-visible docs, behavior, tooling, runtime, or workflow changes.
7. Keep command examples repo-native and current.
8. Use public-domain, open-source, or clearly no-longer-copyright book examples in docs; avoid copyrighted contemporary titles in examples and fixtures unless the task requires them.
9. Avoid documenting unsupported or obsolete UI paths.
10. Run Markdown/frontmatter consistency checks where practical; Go tests are not required for docs-only changes unless code changed.

When docs describe verification, include exact commands and note when checks are conditional.
