import { access, readFile, readdir, stat } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { pages, requiredGeneratedAssets, requiredHomepageGeneratedAssets } from './docs-site-config.mjs'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const siteRoot = join(repoRoot, 'output', 'docs-site')
const failures = []

async function main() {
  await verifyMarkdownSources()
  await verifyMarkdownLinks()
  await verifySiteOutput()
  await verifyGeneratedAssets()
  await verifyHomepageVisualReferences()

  if (failures.length > 0) {
    for (const failure of failures) {
      console.error(`docs verify: ${failure}`)
    }
    process.exitCode = 1
    return
  }

  console.log('Docs verification passed.')
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

async function verifyMarkdownSources() {
  for (const page of pages) {
    await expectFile(join(repoRoot, page.source), `missing docs source ${page.source}`)
  }
}

async function verifyMarkdownLinks() {
  const docsFiles = await listMarkdownFiles(join(repoRoot, 'docs'))
  docsFiles.push(join(repoRoot, 'README.md'))

  for (const file of docsFiles) {
    const markdown = await readFile(file, 'utf8')
    const links = Array.from(markdown.matchAll(/\[[^\]]+\]\(([^)]+)\)/g)).map((match) => match[1])
    for (const link of links) {
      await verifyMarkdownLink(file, link)
    }
  }
}

async function verifyMarkdownLink(file, link) {
  if (link.startsWith('#') || /^(https?:|mailto:|tel:)/.test(link)) {
    return
  }

  const [target] = link.split('#')
  if (!target || target.startsWith('data:')) {
    return
  }

  const resolved = resolve(dirname(file), target)
  try {
    await access(resolved)
  } catch {
    failures.push(`${relativePath(file)} links to missing local target ${link}`)
  }
}

async function verifySiteOutput() {
  for (const page of pages) {
    await expectFile(join(siteRoot, page.output), `missing generated site page ${page.output}`)
  }
  await expectFile(join(siteRoot, 'assets', 'site.css'), 'missing generated site stylesheet')
}

async function verifyGeneratedAssets() {
  const generatedRoot = join(siteRoot, 'assets', 'generated')
  try {
    const info = await stat(generatedRoot)
    if (!info.isDirectory()) {
      return
    }
  } catch {
    console.log('Docs verification skipped generated visual asset checks because output/docs-site/assets/generated does not exist.')
    return
  }

  for (const asset of requiredGeneratedAssets) {
    await expectFile(
      join(generatedRoot, asset),
      `missing generated visual asset assets/generated/${asset}`,
    )
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

async function listMarkdownFiles(root) {
  const files = []
  const entries = await readdir(root, { withFileTypes: true })
  for (const entry of entries) {
    const path = join(root, entry.name)
    if (entry.isDirectory()) {
      files.push(...await listMarkdownFiles(path))
    } else if (entry.isFile() && entry.name.endsWith('.md')) {
      files.push(path)
    }
  }
  return files
}

function relativePath(path) {
  return path.replace(`${repoRoot}/`, '')
}

try {
  await main()
} catch (error) {
  console.error(error)
  process.exitCode = 1
}
