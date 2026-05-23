# Docs Visuals

The docs visual workflow generates web UI screenshots, CLI terminal captures, CLI animated GIFs, and TUI animated GIFs/static final frames.

## Asset Strategy

Generated assets are not committed to git.

- Pull requests upload `output/docs-visuals/**` as a short-lived GitHub Actions artifact for review.
- `master` publishes the generated site and generated visuals to GitHub Pages.
- Published visual URLs are deterministic under `assets/generated/`.
- Static generated PNG captures also get WebP companions for lighter browser delivery.
- Animated GIFs are post-processed with `gifsicle -O3` when `gifsicle` is available.

Example stable paths after a `master` publish:

```text
https://jeeftor.github.io/audiobook-organizer/assets/generated/web-ui/web-ui-metadata-json-preview.png
https://jeeftor.github.io/audiobook-organizer/assets/generated/web-ui/web-ui-metadata-json-preview.webp
https://jeeftor.github.io/audiobook-organizer/assets/generated/cli/cli-organize-run.gif
https://jeeftor.github.io/audiobook-organizer/assets/generated/tui/tui-organize-preview.gif
```

## Local Commands

Generate all visuals:

```bash
make docs-visuals
```

Build the Starlight docs site from current Markdown and generated visuals:

```bash
make docs-site
```

Verify links and required generated assets:

```bash
make docs-verify
```

Run the full publish-equivalent local path:

```bash
make docs-publish-site
```

## Focused Visual Generation

Web UI screenshots:

```bash
make docs-web-screenshots
```

CLI static captures:

```bash
make docs-cli-captures
```

CLI animated GIFs:

```bash
make docs-cli-gifs
```

TUI captures:

```bash
make docs-tui-captures
```

On macOS, `make docs-tui-captures` builds and uses `audiobook-organizer-vhs:local` so VHS runs in Linux instead of launching the local Chrome app.

The static WebP pass requires the WebP tools package (`cwebp`). GIF optimization is optional locally; install `gifsicle` to match CI output, or set `ABO_DOCS_GIF_OPTIMIZE=0` to skip that pass explicitly.

## GitHub Pages

The docs workflow should use GitHub Pages with the source set to GitHub Actions. The workflow builds visuals, builds `output/docs-starlight`, uploads the normal review artifact, then deploys the site on `master`.

If Pages is disabled in repository settings, the workflow will still produce the review artifacts, but stable README image URLs will not update until Pages is enabled.
