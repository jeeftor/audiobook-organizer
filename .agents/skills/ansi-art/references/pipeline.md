# ANSI Art Pipeline Reference

## Hand-Drawn Asset Pipeline

For authentic ANSI, use a staged asset process:

1. Sketch or rough in the composition in any paint tool or on paper.
2. Import the sketch/reference into an ANSI editor such as Moebius.
3. Rough in broad color blobs first.
4. Zoom in for block, half-block, shade, and highlight cleanup.
5. Export the `.ans` source and a rendered `.png` preview.
6. Do final cleanup in the target viewer/editor when applicable.

The Museum of ZZT article describes a practical version of this flow: paint to
sketch, Moebius to draw, import tooling, then final cleanup. The key idea for a
skill is to avoid trying to “perfect” the art in one pass.

Source: https://museumofzzt.com/article/view/994/the-ansi-art-pipeline-how-hand-drawn-art-goes-from-sketch-to-weave/

## Screenshot Or Image Conversion

When recreating ANSI from a screenshot or image:

1. Determine the real character grid and cell size.
2. Normalize dimensions with nearest-neighbor scaling, not smoothing.
3. Slice into character cells.
4. Pre-render candidate glyph/color combinations using the target font and
   palette.
5. Score each cell with pixel-diff/error and choose the closest candidate.
6. Export `.ans` plus a rendered preview image.
7. Manually clean up ambiguous text and shading.

Important compatibility details:

- CP437 matters for classic `.ans` output.
- Font choice matters; use an IBM VGA/PC style font for preview.
- Shade blocks vary noticeably across fonts.
- Modern terminal UTF-8 semigraphics are useful for CLI output, but they are not
  equivalent to CP437 bytes.

Source: https://bert.org/2023/02/27/recreating-ansi-art-from-a-screenshot/

## Semigraphic Tile Conversion

For modern terminal approximations, Unicode block elements can represent 2x2
pixel tiles. Try the candidate block patterns, average foreground/background
colors for each partition, compute squared RGB error, and choose the lowest
error. This is useful for generated previews or small terminal images, but it
is not the same as hand-authored ANSI logo art.

Source: https://hbfs.wordpress.com/2017/11/14/ansi-art/

## Tooling Notes

- Moebius and Pablodraw are good authoring references for real ANSI constraints.
- Ansilove/go-ansi-style renderers are useful for deterministic PNG previews.
- Terminal previews are still required because modern fonts and color handling
  vary.
- Keep generated CLI assets compact; avoid large art that pushes help content
  far below the fold.
