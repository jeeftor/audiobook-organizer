import { spawn } from 'node:child_process'
import { once } from 'node:events'
import { copyFile, mkdir, rm, writeFile } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const screenshotDir = join(repoRoot, 'output', 'docs-visuals', 'web-ui')
const sampleRoot = join(repoRoot, 'output', 'docs-web-ui-sample')
const goBuildCache = process.env.GOCACHE || join(repoRoot, 'output', 'go-build-cache')
const metadataSourceDir = 'output/docs-web-ui-sample/metadata-json/source'
const metadataOutputDir = 'output/docs-web-ui-sample/metadata-json/organized'
const embeddedSourceDir = 'output/docs-web-ui-sample/embedded/source'
const embeddedOutputDir = 'output/docs-web-ui-sample/embedded/organized'
const serverURLPattern = /http:\/\/127\.0\.0\.1:(\d+)\/\?token=([a-f0-9]+)/
const screenshotViewport = { width: 1440, height: 1200 }
const serverStartupTimeoutMs = Number(process.env.ABO_DOCS_SERVER_STARTUP_TIMEOUT_MS || 120_000)

async function main() {
  await createSampleLibrary()
  await rm(screenshotDir, { recursive: true, force: true })
  await mkdir(screenshotDir, { recursive: true })

  const server = await startWebServer()
  let browser

  try {
    browser = await launchBrowser()
    const page = await browser.newPage({ viewport: screenshotViewport, deviceScaleFactor: 1 })
    await page.goto(server.url, { waitUntil: 'networkidle' })
    await disableMotion(page)
    await page.getByRole('heading', { name: 'Audiobook Organizer' }).waitFor()

    await page.getByRole('textbox', { name: 'Source folder' }).fill(metadataSourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(metadataOutputDir)
    await page.getByRole('radio', { name: 'metadata.json' }).click()
    await waitForOrganizePreview(page, {
      sourceText: 'output/docs-web-ui-sample/metadata-json',
      destinationText: 'LibriVox Horror Classics',
      metadataFound: 2,
      plannedMoves: 2,
    })
    await page.screenshot({
      path: join(screenshotDir, 'web-ui-metadata-json-preview.png'),
      fullPage: false,
    })

    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    await page.getByRole('heading', { name: 'Reviewed Organize Plan' }).waitFor({ timeout: 10_000 })
    await page.screenshot({
      path: join(screenshotDir, 'web-ui-metadata-json-review.png'),
      fullPage: false,
    })

    await page.goto(server.url, { waitUntil: 'networkidle' })
    await disableMotion(page)
    await page.getByRole('heading', { name: 'Audiobook Organizer' }).waitFor()
    await page.getByRole('textbox', { name: 'Source folder' }).fill(embeddedSourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(embeddedOutputDir)
    await page.getByRole('radio', { name: 'Embedded metadata by file' }).click()
    await waitForOrganizePreview(page, {
      sourceText: 'output/docs-web-ui-sample/embedded',
      destinationText: 'The Case of Charles Dexter Ward',
      metadataFound: 3,
      plannedMoves: 3,
    })
    await page.screenshot({
      path: join(screenshotDir, 'web-ui-embedded-metadata-preview.png'),
      fullPage: false,
    })
  } finally {
    await browser?.close()
    await server.stop()
    await rm(sampleRoot, { recursive: true, force: true })
  }

  console.log('Wrote local docs screenshots:')
  console.log('  output/docs-visuals/web-ui/web-ui-metadata-json-preview.png')
  console.log('  output/docs-visuals/web-ui/web-ui-metadata-json-review.png')
  console.log('  output/docs-visuals/web-ui/web-ui-embedded-metadata-preview.png')
}


async function waitForOrganizePreview(page, { sourceText, destinationText, metadataFound, plannedMoves }) {
  await page.getByRole('heading', { name: 'Organize preview ready' }).waitFor({ timeout: 30_000 })
  await waitForResultValue(page, 'Metadata found', metadataFound)
  await waitForResultValue(page, 'Planned moves', plannedMoves)
  await page.locator('.move-list').filter({ hasText: sourceText }).first().waitFor({ timeout: 10_000 })
  await page.locator('.move-list').filter({ hasText: destinationText }).first().waitFor({ timeout: 10_000 })
  await page.getByRole('button', { name: 'Review & Run', exact: true }).waitFor({ timeout: 10_000 })
  await page.waitForTimeout(500)
}

async function waitForResultValue(page, label, expectedValue) {
  await page.waitForFunction(
    ({ labelText, value }) => {
      const spans = Array.from(document.querySelectorAll('.result-grid span'))
      return spans.some((span) => {
        const strong = span.nextElementSibling
        return span.textContent?.trim() === labelText && strong?.textContent?.trim() === String(value)
      })
    },
    { labelText: label, value: expectedValue },
    { timeout: 10_000 },
  )
}

async function createSampleLibrary() {
  await rm(sampleRoot, { recursive: true, force: true })
  await mkdir(join(sampleRoot, 'metadata-json', 'organized'), { recursive: true })
  await mkdir(join(sampleRoot, 'embedded', 'source'), { recursive: true })
  await mkdir(join(sampleRoot, 'embedded', 'organized'), { recursive: true })
  await createEmbeddedSampleBooks()
  await createMetadataBook({
    directory: 'charles-dexter-ward',
    sourceFile: 'testdata/mp3flat/charlesdexterward_01_lovecraft_64kb.mp3',
    targetFile: '01 - Chapter 1.mp3',
    metadata: {
      title: 'The Case of Charles Dexter Ward',
      authors: ['H. P. Lovecraft'],
      series: ['LibriVox Horror Classics #1'],
      narrators: ['LibriVox Volunteers'],
      publishedYear: '1927',
    },
  })
  await createMetadataBook({
    directory: 'falstaffs-wedding',
    sourceFile: 'testdata/mp3flat/falstaffswedding1766version_1_kenrick_64kb.mp3',
    targetFile: '01 - Act 1.mp3',
    metadata: {
      title: "Falstaff's Wedding",
      authors: ['William Kenrick'],
      series: ['Public Domain Drama #1'],
      narrators: ['LibriVox Volunteers'],
      publishedYear: '1766',
    },
  })
}

async function createEmbeddedSampleBooks() {
  await copyFile(
    join(repoRoot, 'testdata/mp3flat/charlesdexterward_01_lovecraft_64kb.mp3'),
    join(sampleRoot, 'embedded', 'source', 'charlesdexterward_01_lovecraft_64kb.mp3'),
  )
  await copyFile(
    join(repoRoot, 'testdata/mp3flat/falstaffswedding1766version_1_kenrick_64kb.mp3'),
    join(sampleRoot, 'embedded', 'source', 'falstaffswedding1766version_1_kenrick_64kb.mp3'),
  )
  await copyFile(
    join(repoRoot, 'testdata/mp3flat/perouse_01_scott_64kb.mp3'),
    join(sampleRoot, 'embedded', 'source', 'perouse_01_scott_64kb.mp3'),
  )
}

async function createMetadataBook({ directory, sourceFile, targetFile, metadata }) {
  const bookDir = join(sampleRoot, 'metadata-json', 'source', directory)
  await mkdir(bookDir, { recursive: true })
  await writeFile(join(bookDir, 'metadata.json'), JSON.stringify(metadata, null, 2))
  await copyFile(join(repoRoot, sourceFile), join(bookDir, targetFile))
}

async function startWebServer() {
  const child = spawn('go', ['run', '.', 'web', '--host', '127.0.0.1', '--port', '0', '--no-open'], {
    cwd: repoRoot,
    detached: process.platform !== 'win32',
    env: {
      ...process.env,
      GOCACHE: goBuildCache,
    },
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  let output = ''
  const startup = new Promise((resolveStartup, rejectStartup) => {
    const timeout = setTimeout(() => {
      failStartup(new Error(`Timed out waiting for web server URL after ${serverStartupTimeoutMs}ms.\n${output}`))
    }, serverStartupTimeoutMs)

    let settled = false

    function failStartup(error) {
      if (settled) {
        return
      }
      settled = true
      clearTimeout(timeout)
      void stopWebServer(child).finally(() => rejectStartup(error))
    }

    child.once('error', failStartup)
    child.once('exit', (code, signal) => {
      failStartup(new Error(`Web server exited before startup: code=${code} signal=${signal}\n${output}`))
    })

    child.stdout.on('data', (chunk) => {
      output += chunk.toString()
      const match = serverURLPattern.exec(output)
      if (!match) {
        return
      }
      if (settled) {
        return
      }
      settled = true
      clearTimeout(timeout)
      resolveStartup({
        url: `http://127.0.0.1:${match[1]}/?token=${match[2]}`,
        stop: () => stopWebServer(child),
      })
    })

    child.stderr.on('data', (chunk) => {
      output += chunk.toString()
    })
  })

  return startup
}

async function stopWebServer(child) {
  if (child.exitCode !== null || child.killed) {
    return
  }

  signalWebServer(child, 'SIGTERM')
  await Promise.race([
    once(child, 'exit'),
    new Promise((resolveStop) => {
      setTimeout(() => {
        if (child.exitCode === null) {
          signalWebServer(child, 'SIGKILL')
        }
        resolveStop()
      }, 5_000)
    }),
  ])
}

function signalWebServer(child, signal) {
  try {
    if (process.platform !== 'win32') {
      process.kill(-child.pid, signal)
      return
    }
  } catch {
    // Fall back to signaling the direct child if process-group signaling is unavailable.
  }
  child.kill(signal)
}

async function launchBrowser() {
  try {
    return await chromium.launch({
      executablePath: process.env.ABO_DOCS_BROWSER_EXECUTABLE_PATH || process.env.PUPPETEER_EXECUTABLE_PATH || undefined,
    })
  } catch (error) {
    throw new Error(
      `Unable to launch Chromium for docs screenshots. Run "cd web && npm run install:browsers" or set ABO_DOCS_BROWSER_EXECUTABLE_PATH to a local Chrome/Chrome Headless Shell binary.\n${error}`,
    )
  }
}

async function disableMotion(page) {
  await page.addStyleTag({
    content: `
      *, *::before, *::after {
        animation-delay: 0s !important;
        animation-duration: 0s !important;
        transition-delay: 0s !important;
        transition-duration: 0s !important;
      }
    `,
  })
}

try {
  await main()
} catch (error) {
  console.error(error)
  process.exitCode = 1
}
