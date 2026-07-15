import { spawn } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { mkdtemp, rename } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { expect, test, type Page } from '@playwright/test'
import { startTestServer, type TestServer } from './server'

const repoRoot = new URL('../../..', import.meta.url).pathname

let server: TestServer
let absEnv: Record<string, string>

test.skip(process.env.ABO_ABS_PLAYWRIGHT !== '1', 'Set ABO_ABS_PLAYWRIGHT=1 to run Docker-backed ABS UI tests.')

test.beforeAll(async () => {
  absEnv = loadABSTestingEnv()
  server = await startTestServer()
})

test.beforeEach(async () => {
  await runRepoCommand('make', ['abs-dev-reset-scan'], 8 * 60_000)
})

test.afterAll(async () => {
  await server?.stop()
})

test('rejects an invalid ABS token and keeps the workflow locked', async ({ page }) => {
  test.setTimeout(120_000)

  const absURL = requiredEnv('ABS_PLAIN_URL')
  const audiobookRoot = join(repoRoot, 'test', 'abs', 'runtime', 'plain', 'audiobooks')

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill(audiobookRoot)
  await page.getByLabel('ABS server URL').fill(absURL)
  await page.getByLabel('ABS API token').fill('invalid-token')
  await page.getByLabel('ABS path prefix').fill('/audiobooks')
  await page.getByLabel('Local path prefix').fill(audiobookRoot)
  await page.getByRole('button', { name: 'Test Connection' }).click()

  await expect(page.locator('.inline-alert')).toBeVisible()
  await expect(page.getByLabel('ABS library')).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Validate Paths' })).toBeDisabled()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
})

test('organizes a mounted library using real Audiobookshelf metadata', async ({ page }) => {
  test.setTimeout(120_000)

  const absURL = requiredEnv('ABS_PLAIN_URL')
  const absToken = requiredEnv('ABS_TOKEN')
  const audiobookRoot = join(repoRoot, 'test', 'abs', 'runtime', 'plain', 'audiobooks')

  await loadApp(page)
  await page.getByRole('button', { name: /Organize/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill(audiobookRoot)
  await page.getByRole('textbox', { name: 'Output folder' }).fill(audiobookRoot)
  await page.getByRole('radio', { name: 'Audiobookshelf metadata' }).click()

  await page.getByLabel('ABS server URL').fill(absURL)
  await page.getByLabel('ABS API token').fill(absToken)
  await page.getByLabel('ABS path prefix').fill('/audiobooks')
  await page.getByLabel('Local path prefix').fill(audiobookRoot)
  await page.getByRole('button', { name: 'Test Connection' }).click()
  await selectLibraryByName(page, 'Audiobooks')
  await page.getByRole('button', { name: 'Validate Paths' }).click()
  await expect(page.getByText('ABS libraries loaded and path mappings validated.')).toBeVisible()

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('heading', { name: 'Reviewed Organize Plan' })).toBeVisible()
  await expect(page.locator('.move-list.selectable-list')).toContainText('Charles Dickens')
  await expect(page.locator('.move-list.selectable-list')).toContainText('Lewis Carroll')
  await expect(page.getByRole('button', { name: /Run 2 Selected Moves/ })).toBeEnabled()

  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Run Organize will change files for 2 selected move(s).')
    await dialog.accept()
  })
  await page.getByRole('button', { name: /Run 2 Selected Moves/ }).click()
  await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
  await expect(page.getByText('Open Audiobookshelf to trigger a library scan, inspect the refreshed state, and clean only confirmed missing old paths.')).toBeVisible()
  await expect(page.locator('.move-list.operation-list')).toContainText('Charles Dickens')
  await expect(page.locator('.move-list.operation-list')).toContainText('Lewis Carroll')
})

test('drives ABS setup and operations against the real ABS harness', async ({ page }) => {
  test.setTimeout(120_000)

  const absURL = requiredEnv('ABS_PLAIN_URL')
  const absToken = requiredEnv('ABS_TOKEN')
  const audiobookRoot = join(repoRoot, 'test', 'abs', 'runtime', 'plain', 'audiobooks')

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill(audiobookRoot)
  await page.getByLabel('ABS server URL').fill(absURL)
  await page.getByLabel('ABS API token').fill(absToken)
  await page.getByLabel('ABS path prefix').fill('/audiobooks')
  await page.getByLabel('Local path prefix').fill(audiobookRoot)

  await page.getByRole('button', { name: 'Test Connection' }).click()
  const librarySelect = page.getByLabel('ABS library')
  await expect(librarySelect).toBeVisible()
  await selectLibraryByName(page, 'Audiobooks')
  await expect(page.locator('.library-option.selected').filter({ hasText: 'Audiobooks' })).toBeVisible()

  await page.getByRole('button', { name: 'Validate Paths' }).click()
  await expect(page.getByText('ABS libraries loaded and path mappings validated.')).toBeVisible()
  await expect(page.getByText(audiobookRoot).first()).toBeVisible()

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('button', { name: 'Load ABS Items' })).toBeEnabled()
  await expect(page.getByRole('button', { name: 'Check Library State' })).toBeEnabled()
  await page.getByRole('button', { name: 'Load ABS Items' }).click()
  await page.getByRole('button', { name: 'Check Library State' }).click()

  await expect(page.getByRole('heading', { name: 'ABS Operation Results' })).toBeVisible()
  await expectSummaryValue(page, 'Metadata items', '2')
  await expectSummaryValue(page, 'Library state items', '2')
  await expectSummaryValue(page, 'Missing / invalid', '0 / 0')
  await expect(page.getByText(audiobookRoot).first()).toBeVisible()

  const missingSource = join(audiobookRoot, 'loose')
  const missingTarget = join(await mkdtemp(join(tmpdir(), 'abo-web-abs-missing-')), 'loose')
  await rename(missingSource, missingTarget)

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await page.getByRole('button', { name: 'Trigger Scan' }).click()
  await expect(page.getByText(/Scan triggered for/)).toBeVisible()

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await waitForMissingState(page)
  await expect(page.locator('.move-list em').filter({ hasText: 'Missing' })).toBeVisible()

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('button', { name: 'Clean Missing Items' })).toBeDisabled()

  await page.getByLabel('I understand this removes ABS missing item records').check()
  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Clean missing ABS item records')
    await dialog.accept()
  })
  await page.getByRole('button', { name: 'Clean Missing Items' }).click()
  await expect(page.getByText(/Cleanup completed for/)).toBeVisible()

  await page.getByRole('button', { name: 'Trigger Scan' }).click()
  await expect(page.getByText(/Scan triggered for/)).toBeVisible()
  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await waitForCleanState(page)
  await expectSummaryValue(page, 'Library state items', '1')
  await expectSummaryValue(page, 'Missing / invalid', '0 / 0')

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('heading', { name: 'ABS Operation Results' })).toBeVisible()
  await expect(page.locator('.review-layout .result-grid')).toContainText('Library state items')
  await expect(page.locator('.review-layout .result-grid')).toContainText('Last cleanup')
})

async function waitForMissingState(page: Page): Promise<void> {
  await waitForLibraryState(page, async () => {
    const missingBadgeCount = await page.locator('.move-list em').filter({ hasText: 'Missing' }).count()
    return missingBadgeCount > 0
  })
}

async function waitForCleanState(page: Page): Promise<void> {
  await waitForLibraryState(page, async () => {
    const missingBadgeCount = await page.locator('.move-list em').filter({ hasText: 'Missing' }).count()
    const summary = await summaryValue(page, 'Missing / invalid')
    return missingBadgeCount === 0 && summary === '0 / 0'
  })
}

async function waitForLibraryState(page: Page, done: () => Promise<boolean>): Promise<void> {
  const deadline = Date.now() + 60_000
  let lastSummary = ''

  while (Date.now() < deadline) {
    const checkState = page.getByRole('button', { name: 'Check Library State' })
    await expect(checkState).toBeEnabled()
    await checkState.click()
    await expect(page.getByRole('heading', { name: /ABS Operation/ })).toBeVisible()
    lastSummary = await summaryValue(page, 'Missing / invalid')
    if (await done()) {
      return
    }
    await page.waitForTimeout(2_000)
  }

  throw new Error(`Timed out waiting for ABS library state. Last Missing / invalid summary: ${lastSummary}`)
}

async function selectLibraryByName(page: Page, libraryName: string): Promise<void> {
  const option = page.getByLabel('ABS library').locator('option', { hasText: libraryName }).first()
  const value = await option.getAttribute('value')
  if (!value) {
    throw new Error(`ABS library option not found: ${libraryName}`)
  }
  await page.getByLabel('ABS library').selectOption(value)
}

async function loadApp(page: Page): Promise<void> {
  const consoleMessages: string[] = []
  page.on('console', (message) => {
    if (['error', 'warning'].includes(message.type())) {
      consoleMessages.push(`${message.type()}: ${message.text()}`)
    }
  })

  await page.goto(server.url)
  await expect(page.locator('#app')).not.toBeEmpty()

  expect(consoleMessages, 'No browser console warnings/errors during initial render').toEqual([])
}

async function expectSummaryValue(page: Page, label: string, value: string): Promise<void> {
  await expect(summaryLocator(page, label)).toHaveText(value)
}

async function summaryValue(page: Page, label: string): Promise<string> {
  return (await summaryLocator(page, label).textContent())?.trim() ?? ''
}

function summaryLocator(page: Page, label: string) {
  return page.locator(`.result-grid >> xpath=./span[normalize-space(.)="${label}"]/following-sibling::strong[1]`)
}

function loadABSTestingEnv(): Record<string, string> {
  const envPath = join(repoRoot, 'test', 'abs', '.env.testing')
  const data = readFileSync(envPath, 'utf8')
  const values: Record<string, string> = {}
  for (const line of data.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) {
      continue
    }
    const separator = trimmed.indexOf('=')
    if (separator === -1) {
      continue
    }
    values[trimmed.slice(0, separator)] = trimmed.slice(separator + 1)
  }
  return values
}

function requiredEnv(name: string): string {
  const value = process.env[name] || absEnv[name]
  if (!value) {
    throw new Error(`${name} is required for ABS Playwright tests`)
  }
  return value
}

async function runRepoCommand(command: string, args: string[], timeoutMs: number): Promise<void> {
  await new Promise<void>((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: repoRoot,
      env: {
        ...process.env,
        ABS_ENV_FILE: 'test/abs/.env.testing',
        NO_COLOR: '1',
        TERM: 'dumb',
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    let output = ''
    const timeout = setTimeout(() => {
      child.kill('SIGTERM')
      reject(new Error(`${command} ${args.join(' ')} timed out after ${timeoutMs}ms\n${output}`))
    }, timeoutMs)

    child.stdout.on('data', (chunk: Buffer) => {
      output += chunk.toString()
    })
    child.stderr.on('data', (chunk: Buffer) => {
      output += chunk.toString()
    })
    child.once('error', (error) => {
      clearTimeout(timeout)
      reject(error)
    })
    child.once('exit', (code, signal) => {
      clearTimeout(timeout)
      if (code === 0) {
        resolve()
        return
      }
      reject(new Error(`${command} ${args.join(' ')} failed: code=${code} signal=${signal}\n${output}`))
    })
  })
}
