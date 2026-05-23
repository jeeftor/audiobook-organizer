# Logo Layer Restart

This is a clean restart of the logo extraction work.

The files here intentionally preserve too much rather than too little. The previous pass clipped parts of the `A` and `K` because cleanup masks used hard crop bands. This restart keeps a broad raw mask so the original wordmark geometry remains intact before any cleanup or tracing decisions.

Files:

- `logo-source.png`: untouched copy of `docs/logo.png`.
- `wordmark-region-reference.png`: simple crop of the lower logo region for visual comparison.
- `logo-text-mask-raw.png`: broad raw mask for likely wordmark pixels. It includes background fragments by design.
- `logo-text-mask-raw-trimmed.png`: trimmed view of the raw mask.
- `logo-text-layer-raw.png`: original pixels clipped by the raw mask, full canvas size.
- `logo-text-layer-raw-trimmed.png`: trimmed view of the raw wordmark layer.
- `logo-text-mask-raw-for-trace.pbm`: Potrace input derived from the raw mask.
- `logo-text-trace-raw.svg`: Potrace output from the raw mask.
- `logo-layered-raw.svg`: minimal SVG wrapper for the raw extracted wordmark layer.

Next cleanup should be manual and localized. Avoid hard horizontal or vertical crop bands over the wordmark, because the brush strokes extend unpredictably.
