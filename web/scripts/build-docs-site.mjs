import { copyFile, mkdir, readdir, readFile, rm, stat, writeFile } from 'node:fs/promises'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { marked } from 'marked'
import { pages, siteBaseURL } from './docs-site-config.mjs'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const siteRoot = join(repoRoot, 'output', 'docs-site')
const generatedVisualsRoot = join(repoRoot, 'output', 'docs-visuals')
const siteGeneratedRoot = join(siteRoot, 'assets', 'generated')
const siteDocsAssetRoot = join(siteRoot, 'assets', 'docs')

const outputBySource = new Map(
  pages.map((page) => [normalizePath(page.source), page.output]),
)

marked.setOptions({
  gfm: true,
  headerIds: true,
  mangle: false,
})

async function main() {
  await rm(siteRoot, { recursive: true, force: true })
  await mkdir(siteRoot, { recursive: true })
  await copyStaticDocsAssets()
  await copyGeneratedVisuals()

  for (const page of pages) {
    await renderPage(page)
  }

  await writeFile(join(siteRoot, 'robots.txt'), 'User-agent: *\nAllow: /\n')
  await writeFile(join(siteRoot, 'sitemap.txt'), pages.map((page) => new URL(page.output, siteBaseURL).toString()).join('\n') + '\n')
  console.log(`Wrote docs site to ${relative(repoRoot, siteRoot)}`)
}

async function renderPage(page) {
  const sourcePath = join(repoRoot, page.source)
  const markdown = await readFile(sourcePath, 'utf8')
  const content = rewriteLinks(marked.parse(markdown), page)
  const outputPath = join(siteRoot, page.output)
  await mkdir(dirname(outputPath), { recursive: true })
  await writeFile(outputPath, renderHTML(page, content))
}

function rewriteLinks(html, page) {
  const rewritten = html
    .replace(/href="([^"]+)"/g, (_match, href) => `href="${rewriteHref(href, page)}"`)
    .replace(/src="([^"]+)"/g, (_match, src) => `src="${rewriteSrc(src, page)}"`)

  return linkStandaloneImages(rewritten)
}

function rewriteHref(href, page) {
  if (isExternal(href) || href.startsWith('#')) {
    return href
  }

  const [rawPath, anchor = ''] = href.split('#')
  const sourceDir = dirname(page.source)
  const normalized = normalizePath(join(sourceDir, rawPath))

  if (rawPath === '../README.md' || rawPath === 'README.md' || normalized === 'README.md') {
    return `index.html${anchor ? `#${anchor}` : ''}`
  }

  const mapped = outputBySource.get(normalized)
  if (mapped) {
    return relativeURL(dirname(page.output), mapped, anchor)
  }

  if (rawPath.endsWith('.md')) {
    return href
  }

  return href
}

function rewriteSrc(src, page) {
  if (isExternal(src) || src.startsWith('#')) {
    return src
  }

  if (src.startsWith('assets/generated/')) {
    return relativeURL(dirname(page.output), src)
  }

  const normalized = normalizePath(join(dirname(page.source), src))
  if (normalized.startsWith('docs/')) {
    const assetPath = normalized.slice('docs/'.length)
    return relativeURL(dirname(page.output), join('assets', 'docs', assetPath))
  }

  return src
}

function linkStandaloneImages(html) {
  return html.replace(
    /<p><img src="([^"]+)" alt="([^"]*)"><\/p>/g,
    (_match, src, alt) => {
      const href = imageHref(src)
      return `<figure class="doc-image"><a href="${href}" target="_blank" rel="noopener"><img src="${src}" alt="${alt}"></a></figure>`
    },
  )
}

function imageHref(src) {
  const docsAssetMatch = src.match(/assets\/docs\/([^#?]+)/)
  if (docsAssetMatch) {
    return `https://github.com/jeeftor/audiobook-organizer/blob/master/docs/${docsAssetMatch[1]}`
  }
  return src
}

function relativeURL(fromDir, target, anchor = '') {
  const normalizedFrom = fromDir === '.' ? '' : fromDir
  let url = normalizePath(relative(normalizedFrom || '.', target))
  if (!url.startsWith('.')) {
    url = `./${url}`
  }
  return `${url}${anchor ? `#${anchor}` : ''}`
}

function renderHTML(page, content) {
  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${escapeHTML(page.title)} | Audiobook Organizer Docs</title>
  <meta name="description" content="Audiobook Organizer for Audiobookshelf and local audiobook libraries. Preview, organize, rename, inspect metadata, choose layouts, use dry-run and undo workflows, and publish stable docs visuals.">
  <link rel="stylesheet" href="${relativeURL(dirname(page.output), 'assets/site.css')}">
</head>
<body>
  <a class="skip-link" href="#content">Skip to content</a>
  <div class="layout">
    <aside class="sidebar">
      <a class="brand" href="${relativeURL(dirname(page.output), 'index.html')}">
        <span class="brand-name">Audiobook Organizer</span>
        <span class="brand-subtitle">Docs</span>
      </a>
      ${renderNav(page)}
    </aside>
    <main id="content" class="content">
      ${content}
      <footer class="site-footer">
        These are AI-generated docs. Any errors are the AI's fault, obviously.
        <a href="https://github.com/jeeftor/audiobook-organizer/issues">Open an issue</a>.
      </footer>
    </main>
  </div>
</body>
</html>
`
}

function renderNav(currentPage) {
  const groups = new Map()
  for (const page of pages) {
    const group = groups.get(page.group) || []
    group.push(page)
    groups.set(page.group, group)
  }

  return Array.from(groups.entries())
    .map(([groupName, groupPages]) => {
      const links = groupPages.map((page) => {
        const active = page.output === currentPage.output ? ' aria-current="page"' : ''
        return `<a${active} href="${relativeURL(dirname(currentPage.output), page.output)}">${escapeHTML(page.title)}</a>`
      }).join('\n')
      return `<nav aria-label="${escapeHTML(groupName)}"><h2>${escapeHTML(groupName)}</h2>${links}</nav>`
    })
    .join('\n')
}

async function copyStaticDocsAssets() {
  await mkdir(siteDocsAssetRoot, { recursive: true })
  const entries = await readdir(join(repoRoot, 'docs'), { withFileTypes: true })
  for (const entry of entries) {
    if (!entry.isFile()) {
      continue
    }
    if (!['.png', '.jpg', '.jpeg', '.gif', '.svg', '.webp'].includes(extname(entry.name).toLowerCase())) {
      continue
    }
    await copyFile(join(repoRoot, 'docs', entry.name), join(siteDocsAssetRoot, entry.name))
  }
  await writeFile(join(siteRoot, 'assets', 'site.css'), siteCSS())
}

async function copyGeneratedVisuals() {
  try {
    const info = await stat(generatedVisualsRoot)
    if (!info.isDirectory()) {
      return
    }
  } catch {
    return
  }
  await copyDirectory(generatedVisualsRoot, siteGeneratedRoot)
}

async function copyDirectory(source, target) {
  await mkdir(target, { recursive: true })
  const entries = await readdir(source, { withFileTypes: true })
  for (const entry of entries) {
    if (entry.name === '.DS_Store') {
      continue
    }
    const sourcePath = join(source, entry.name)
    const targetPath = join(target, entry.name)
    if (entry.isDirectory()) {
      await copyDirectory(sourcePath, targetPath)
    } else if (entry.isFile()) {
      await copyFile(sourcePath, targetPath)
    }
  }
}

function siteCSS() {
  return `:root {
  color-scheme: light;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  line-height: 1.55;
  color: #182233;
  background: #eef2f6;
}

* { box-sizing: border-box; }

body {
  margin: 0;
}

a { color: #1260a8; }

.skip-link {
  position: absolute;
  left: 16px;
  top: -48px;
  background: #172033;
  color: white;
  padding: 8px 12px;
  z-index: 10;
}

.skip-link:focus { top: 16px; }

.layout {
  display: grid;
  grid-template-columns: 268px minmax(0, 1fr);
  min-height: 100vh;
}

.sidebar {
  border-right: 1px solid #1f2a3d;
  background: #121a27;
  padding: 20px 22px;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: auto;
}

.brand {
  display: grid;
  gap: 2px;
  padding: 2px 2px 18px;
  border-bottom: 1px solid #d7deea;
  text-decoration: none;
  margin-bottom: 12px;
}

.brand-name {
  color: #ffffff;
  font-size: 20px;
  font-weight: 800;
  line-height: 1.12;
}

.brand-subtitle {
  color: #8fa1bb;
  font-size: 12px;
  font-weight: 750;
  text-transform: uppercase;
}

nav { margin: 0 0 24px; }
nav h2 {
  color: #8fa1bb;
  font-size: 12px;
  margin: 0 0 8px;
  text-transform: uppercase;
  letter-spacing: 0;
}
nav a {
  display: block;
  color: #d7deea;
  text-decoration: none;
  padding: 7px 9px;
  border-radius: 6px;
  font-size: 14px;
}
nav a:hover,
nav a[aria-current="page"] {
  background: #213149;
  color: #ffffff;
}

.content {
  width: min(100%, 1120px);
  padding: 46px 44px 86px;
}

h1, h2, h3 {
  line-height: 1.2;
  color: #101828;
}

h1 {
  font-size: 42px;
  margin: 0 0 18px;
}

h2 {
  margin-top: 40px;
  padding-top: 10px;
  border-top: 1px solid #cdd6e2;
}

p, li { font-size: 16px; }

.lead {
  color: #d9e6f7;
  font-size: 20px;
  line-height: 1.45;
  margin: 0 0 22px;
}

.doc-hero {
  display: grid;
  grid-template-columns: minmax(0, 0.82fr) minmax(360px, 1fr);
  gap: 28px;
  align-items: center;
  margin: 0 0 34px;
  padding: 34px;
  border-radius: 12px;
  background: #172033;
  color: #f8fafc;
  box-shadow: 0 24px 70px rgb(23 32 51 / 18%);
}

.product-hero {
  margin-bottom: 28px;
}

.doc-hero h1 {
  color: #ffffff;
  font-size: 48px;
  margin-bottom: 14px;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 42px;
  padding: 0 16px;
  border-radius: 7px;
  background: #48c6b0;
  color: #0f172a;
  font-weight: 750;
  text-decoration: none;
}

.button.secondary {
  background: transparent;
  border: 1px solid #56677f;
  color: #f8fafc;
}

.hero-media,
.visual-grid figure {
  margin: 0;
  overflow: hidden;
  border: 1px solid #26364e;
  border-radius: 10px;
  background: #0f172a;
  box-shadow: 0 14px 36px rgb(15 23 42 / 22%);
}

.hero-media img,
.visual-grid img {
  display: block;
  width: 100%;
  aspect-ratio: 16 / 10;
  object-fit: cover;
  object-position: top left;
  border-radius: 0;
}

.hero-media figcaption,
.visual-grid figcaption {
  padding: 10px 12px;
  color: #d7deea;
  font-size: 13px;
  background: #121a27;
}

.visual-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin: 28px 0 34px;
}

.capability-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin: 22px 0 34px;
}

.capability-grid article {
  min-height: 150px;
  padding: 18px;
  border: 1px solid #d1dae7;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 10px 24px rgb(15 23 42 / 5%);
}

.capability-grid h3 {
  margin: 0 0 9px;
}

.capability-grid p {
  margin: 0;
  color: #536176;
}

.workflow-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  padding: 0;
  list-style: none;
}

.workflow-list li {
  display: grid;
  gap: 5px;
  min-height: 118px;
  padding: 16px;
  border: 1px solid #d1dae7;
  border-radius: 8px;
  background: #ffffff;
}

.workflow-list strong {
  color: #101828;
  font-size: 18px;
}

.workflow-list span {
  color: #536176;
}

.doc-image {
  margin: 22px 0;
}

.doc-image a:first-child {
  display: block;
}

.doc-image img {
  display: block;
  border: 1px solid #d1dae7;
  background: #ffffff;
  box-shadow: 0 12px 30px rgb(15 23 42 / 10%);
}

.doc-image figcaption {
  margin-top: 8px;
  font-size: 13px;
}

.media-callout {
  display: grid;
  grid-template-columns: minmax(260px, 420px) minmax(0, 1fr);
  gap: 22px;
  align-items: start;
  margin: 22px 0;
}

.media-callout-image {
  display: block;
}

.media-callout-image img {
  display: block;
  width: 100%;
  border: 1px solid #d1dae7;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 12px 30px rgb(15 23 42 / 10%);
}

.media-callout-copy {
  padding-top: 2px;
}

.media-callout-copy p:first-child {
  margin-top: 0;
}

.image-pair {
  display: grid;
  grid-template-columns: minmax(150px, 0.72fr) minmax(260px, 1fr);
  gap: 18px;
  align-items: start;
  margin: 22px 0;
}

.image-pair figure {
  margin: 0;
}

.image-pair img {
  display: block;
  width: auto;
  max-width: 100%;
  max-height: 260px;
  border: 1px solid #d1dae7;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 12px 30px rgb(15 23 42 / 10%);
}

.image-pair figcaption {
  margin-top: 8px;
  color: #536176;
  font-size: 13px;
}

.image-pair figcaption a {
  color: #1260a8;
}

.site-footer {
  margin-top: 56px;
  padding-top: 18px;
  border-top: 1px solid #cdd6e2;
  color: #627084;
  font-size: 13px;
}

.site-footer a {
  color: #1260a8;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin: 18px 0;
  background: #ffffff;
  box-shadow: 0 10px 28px rgb(15 23 42 / 6%);
}

th, td {
  border: 1px solid #d9dee8;
  padding: 10px 12px;
  text-align: left;
  vertical-align: top;
}

th { background: #e8eef6; }

code {
  background: #edf0f5;
  border-radius: 4px;
  padding: 0.12em 0.28em;
}

pre {
  overflow: auto;
  padding: 16px;
  border-radius: 8px;
  background: #111827;
  color: #f8fafc;
}

pre code {
  background: transparent;
  padding: 0;
}

img {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
}

@media (max-width: 860px) {
  .layout { display: block; }
  .sidebar {
    position: sticky;
    z-index: 5;
    height: auto;
    padding: 12px;
    border-right: 0;
    border-bottom: 1px solid #d9dee8;
    overflow-x: auto;
    white-space: nowrap;
  }
  .brand {
    display: inline-block;
    padding: 0;
    border-bottom: 0;
    margin: 0 12px 0 0;
    vertical-align: middle;
  }
  .brand-name { font-size: 15px; }
  .brand-subtitle { display: none; }
  .sidebar nav {
    display: contents;
  }
  .sidebar nav h2 {
    display: none;
  }
  .sidebar nav a {
    display: inline-flex;
    margin-right: 4px;
  }
  .content { padding: 32px 20px 64px; }
  h1 { font-size: 34px; }
  .doc-hero {
    grid-template-columns: 1fr;
    padding: 24px;
  }
  .doc-hero h1 { font-size: 36px; }
  .visual-grid,
  .capability-grid,
  .workflow-list,
  .image-pair,
  .media-callout {
    grid-template-columns: 1fr;
  }
}
`
}

function normalizePath(path) {
  return path.replaceAll('\\', '/').replace(/^\.\//, '')
}

function isExternal(value) {
  return /^(https?:|mailto:|tel:)/.test(value)
}

function escapeHTML(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
}

try {
  await main()
} catch (error) {
  console.error(error)
  process.exitCode = 1
}
