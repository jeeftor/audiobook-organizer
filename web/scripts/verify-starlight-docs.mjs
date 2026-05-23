import { glob, readFile, stat } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { pages, requiredGeneratedAssets, requiredHomepageGeneratedAssets } from './docs-site-config.mjs'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const siteRoot = join(repoRoot, 'output', 'docs-starlight')
const failures = []

await main()

async function main() {
  await verifySiteOutput()
  await verifyNoPlaceholders()
  await verifyInternalLinks()
  await verifyGeneratedAssets()
  await verifyHomepageVisualReferences()

  if (failures.length > 0) {
    for (const failure of failures) {
      console.error(`docs verify: ${failure}`)
    }
    process.exitCode = 1
    return
  }

  console.log('Starlight docs verification passed.')
}

async function verifySiteOutput() {
  for (const page of pages) {
    const output = page.output === 'index.html'
      ? 'index.html'
      : page.output.replace(/\.html$/, '/index.html')
    await expectFile(join(siteRoot, output), `missing Starlight page ${output}`)
  }
  await expectFile(join(siteRoot, 'favicon.svg'), 'missing Starlight favicon')
}

async function verifyNoPlaceholders() {
  for await (const file of glob('**/*.html', { cwd: siteRoot })) {
    const html = await readFile(join(siteRoot, file), 'utf8')
    if (/migration placeholder/i.test(html)) {
      failures.push(`Starlight page still contains migration placeholder text: ${file}`)
    }
  }
}

async function verifyInternalLinks() {
  for await (const file of glob('**/*.html', { cwd: siteRoot })) {
    const html = await readFile(join(siteRoot, file), 'utf8')
    for (const match of html.matchAll(/href="([^"]+)"/g)) {
      await verifyInternalHref(file, match[1])
    }
  }
}

async function verifyInternalHref(file, href) {
  if (!href.startsWith('/audiobook-organizer/')) {
    return
  }
  if (href.startsWith('/audiobook-organizer/_astro/') || href.startsWith('/audiobook-organizer/pagefind/')) {
    return
  }

  const withoutBase = href.replace('/audiobook-organizer/', '').split('#')[0].split('?')[0]
  const target = withoutBase === ''
    ? 'index.html'
    : withoutBase.endsWith('/')
      ? join(withoutBase, 'index.html')
      : withoutBase

  await expectFile(join(siteRoot, target), `${file} links to missing Starlight target ${href}`)
}

async function verifyGeneratedAssets() {
  const generatedRoot = join(siteRoot, 'assets', 'generated')
  try {
    const info = await stat(generatedRoot)
    if (!info.isDirectory()) {
      return
    }
  } catch {
    console.log('Docs verification skipped generated visual asset checks because output/docs-starlight/assets/generated does not exist.')
    return
  }

  for (const asset of requiredGeneratedAssets) {
    await expectFile(
      join(generatedRoot, asset),
      `missing generated visual asset assets/generated/${asset}`,
    )
  }
}

async function verifyHomepageVisualReferences() {
  const homepagePath = join(siteRoot, 'index.html')
  let homepage
  try {
    homepage = await readFile(homepagePath, 'utf8')
  } catch {
    failures.push('missing generated homepage for visual reference verification')
    return
  }

  for (const asset of requiredHomepageGeneratedAssets) {
    if (!homepage.includes(asset)) {
      failures.push(`generated homepage does not reference ${asset}`)
    }
  }
}

async function expectFile(path, message) {
  try {
    const info = await stat(path)
    if (!info.isFile()) {
      failures.push(message)
    }
  } catch {
    failures.push(message)
  }
}
