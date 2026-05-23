---
name: ansi-art
description: Create, critique, and iterate terminal logo assets across graphics, ANSI textmode art, plain ASCII fallback, and no-logo terminal tiers; use for CLI banners, BBS-style ANSI logos, CP437/semigraphic art, terminal image fallbacks, and visual QA of terminal branding.
metadata:
  short-description: Design terminal ANSI/ASCII art
---

# ANSI Art

Use this skill when designing or reviewing terminal branding, ANSI logos, ASCII
fallbacks, or image-to-textmode experiments.

## Fallback Ladder

Keep these output tiers distinct:

1. **Graphics**: Kitty, iTerm2, Sixel, or another native image protocol. Show only the bitmap/image logo.
2. **ANSI**: colored textmode art using foreground/background color plus block, shade, half-block, or box characters.
3. **ASCII**: plain no-color 7-bit fallback that remains readable in `NO_COLOR`, `TERM=dumb`, logs, and copied text.
4. **Nothing**: explicit opt-out, CI, non-interactive output, or any context where logo output would be noisy.

Do not call colored figlet text “ANSI art” unless it uses color and textmode
composition deliberately. ANSI art should read like a compact block/semigraphic
asset, not just large ASCII letters with color applied.

## Core Workflow

1. Define constraints first:
   - target width and height in terminal cells,
   - graphics/ANSI/ASCII/no-logo tier,
   - palette depth: 16-color, 256-color, or truecolor,
   - character set: CP437-authentic, UTF-8 semigraphics, or strict 7-bit ASCII,
   - target background: dark, light, or unknown.
2. Sketch the silhouette before adding color.
   - For logos, block out the wordmark and subtitle placement first.
   - For compact CLI banners, keep the main mark under 70 columns when possible.
3. Add color as a second pass.
   - Pick one light source.
   - Build a small palette ramp: shadow, base, highlight, extreme highlight, black/deep cut.
   - Use colored backgrounds and full/half/shade blocks to create mass.
4. Add detail last.
   - Use `░`, `▒`, `▓`, `█`, `▀`, `▄`, `▌`, `▐`, `+`, `_`, and box characters sparingly.
   - Add tiny bright accents only where they improve readability.
   - Use black/background space as a real drawing tool.
5. Create fallbacks as separate assets.
   - ANSI fallback is not “ASCII plus color”.
   - ASCII fallback is not “ANSI with escape codes stripped” unless it still reads well.
6. Preview in the actual terminal and at least one narrow terminal width.

## Craft Rules

- Start with broad silhouette and proportion; do not shade before the full shape is roughed in.
- Avoid rectangular, untouched block fills. Distort contours slightly so the asset feels drawn.
- Avoid mechanical straight gradients. Prefer curved or broken highlight patches.
- Do not let large full-color regions of different colors touch without a dark gap unless that contact is intentional.
- Use shading for transitions and texture, not for hiding weak shape design.
- Use fewer extremes in hard colors such as red, green, and blue; cyan/gray ramps shade more forgivingly.
- Keep compact logos legible before adding texture. If a detail disappears at real terminal size, remove it.
- Check no-color and monochrome readability before treating the ANSI asset as done.

## Implementation Notes

- For Go CLI output, embed ANSI assets as raw strings or structured spans. Keep reset codes explicit.
- Prefer functions that can render both ANSI and ASCII from related, but not identical, source layouts.
- Use ANSI escapes only around visible runs so alignment is not affected.
- Keep width checks escape-aware: strip `\x1b[...m` before counting visible columns.
- Respect `NO_COLOR`; it should choose the ASCII tier, not a colored tier.
- Respect CI and non-interactive output; they should choose the nothing tier unless the user explicitly asks for art.

## Optional References

Read [references/craft.md](references/craft.md) when designing a logo or critiquing ANSI art quality.

Read [references/pipeline.md](references/pipeline.md) when converting from sketches/screenshots/images or producing `.ans` plus rendered previews.
