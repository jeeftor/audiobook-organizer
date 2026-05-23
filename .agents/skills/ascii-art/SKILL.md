---
name: ascii-art
description: Create, critique, and iterate plain ASCII art using no ANSI color or Unicode semigraphics; use for terminal fallbacks, no-color CLI banners, monospace diagrams, 7-bit-safe logos, README-safe art, and log-safe visual layouts.
metadata:
  short-description: Design plain ASCII art
---

# ASCII Art

Use this skill when the output must remain readable as plain text with no color,
no control codes, and no dependency on Unicode block/box glyphs.

## Scope

ASCII art is the plain-text fallback tier. It is not ANSI art.

- Use only printable 7-bit characters unless the user explicitly allows Unicode.
- Do not use ANSI escape sequences.
- Do not rely on color, background fills, box-drawing glyphs, or block elements.
- Make output readable in logs, Markdown, terminals with `NO_COLOR`, and copied text.

If the user wants colored textmode graphics, use `$ansi-art` instead.

## Workflow

1. Define constraints:
   - target width,
   - max height,
   - strict 7-bit ASCII or Unicode allowed,
   - intended context: CLI help, README, comments, diagrams, or logs.
2. Sketch the silhouette using simple characters.
   - Prefer strong outline and whitespace over dense texture.
   - For wordmarks, test legibility at normal terminal font size.
3. Simplify aggressively.
   - Remove decoration that does not survive copy/paste.
   - Avoid diagonals that only align in one font unless the target font is fixed.
4. Add supporting label text only after the main shape reads.
5. Validate:
   - count visible columns,
   - inspect at narrow width,
   - ensure Markdown/code block rendering preserves alignment,
   - ensure it still reads when pasted into a plain text file.

## Character Guidance

Good default palette:

```text
space  .  '  `  -  _  =  +  *  #  /  \  |  (  )  [  ]  <  >
```

For strict compatibility, avoid smart quotes, em/en dashes, box drawing, block
characters, emoji, combining characters, and tabs.

Use diagonals sparingly. ASCII diagonal art is fragile across fonts and widths.

## CLI Logo Rules

- Keep the banner compact; help text should stay close to the top of the screen.
- Prefer a clear wordmark and one subtitle line over a large scene.
- For fallback logo tiers, ASCII should be no more than 4-8 lines unless the user
  explicitly wants a large art piece.
- Make the text label obvious even if the decorative outline fails.

## Optional Reference

Read [references/patterns.md](references/patterns.md) for reusable ASCII layout
patterns and review checks.
