# ANSI Art Craft Reference

## What Counts As ANSI

Classic ANSI art combines a text-cell character set, usually IBM PC CP437, with
escape sequences for foreground/background color. The 16colo.rs beginner guide
summarizes the original scene constraints as CP437-style characters, 16 colors,
80 columns, and 25-line screen chunks. It recommends real ANSI editors such as
Moebius or Pablodraw and warns that generic tools can produce custom formats
that do not preserve those constraints.

ASCII is different: monochrome line art, mostly 7-bit keyboard characters. ANSI
uses color and semigraphic glyphs as a low-resolution pixel medium.

Sources:

- https://forum.16colo.rs/t/how-to-start-drawing-ansi-art/36
- https://raurir.com/posts/ansi-art/
- https://bert.org/2023/02/27/recreating-ansi-art-from-a-screenshot/

## Shape And Readability

- Work in cells, not pixels.
- Block out the entire silhouette before shading.
- Use negative space and dark gaps to separate forms.
- Exaggerate shapes when the grid cannot support literal detail.
- For wordmarks, preserve letter recognition first; texture comes second.
- Avoid over-detailing compact logos. Let the viewer complete small forms.

The Knight/Fuel tutorial is especially useful for compact logo thinking: leave
some spaces black, thicken shapes when needed, avoid abrupt hue jumps, and add
small highlights only where they sell form.

Source: https://www.roysac.com/tutorial/ansitut-tk-fluph.html

## Shading

- Choose a light source before coloring.
- Put darker values on the opposite side consistently.
- Shading should transition or texture; it should not become straight-line
  banding.
- Use curved or broken highlight patches instead of mechanical gradients.
- Add tiny extreme highlights and black cuts inside shaded regions for punch.
- Use `+`, `_`, shade blocks, and half blocks as texture only where they improve
  readability.
- Reuse accent colors already present elsewhere so “random” texture looks
  intentional.

Halaster’s shading tutorial emphasizes avoiding basic geometric shapes, using
shading as detail, adding small extremes, and treating cyan/gray ramps as easier
than hard red/green/blue ramps.

Source: http://www.roysac.com/tutorial/ansitut-hal-shade.html

## Palette Rules

- Default to a dark background unless the target terminal is known to be light.
- Use a small palette map:
  - deep cut/background,
  - shadow,
  - base,
  - highlight,
  - extreme highlight/accent.
- For 16-color ANSI, remember that classic backgrounds are effectively 8-color
  while foregrounds can use bright/bold variants.
- For modern terminal ANSI, 256-color spans are acceptable when the fallback
  ladder includes ASCII.

## Review Checklist

- Visible width stays within the target columns after stripping ANSI escapes.
- The logo reads at normal terminal font size, not only in a screenshot.
- Color regions have enough contrast on dark and light backgrounds, or the
  target background is documented.
- No-color fallback is intentional and not merely escape-stripped noise.
- Escape codes are reset at the end of every colored run or line.
- The output does not print graphics escapes in unsupported terminals.
