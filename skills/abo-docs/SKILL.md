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
2. Keep `AGENTS.md` focused on durable repo rules and route repeatable procedures into `skills/abo-*` or `references/abo-assistant/`.
3. Update `CHANGELOG.md` under `Unreleased` for user-visible docs, behavior, tooling, runtime, or workflow changes.
4. Keep command examples repo-native and current.
5. Avoid documenting unsupported or obsolete UI paths.
6. Run Markdown/frontmatter consistency checks where practical; Go tests are not required for docs-only changes unless code changed.

When docs describe verification, include exact commands and note when checks are conditional.
