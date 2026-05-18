import { expect, test, type Page } from '@playwright/test'
import { access, mkdir, mkdtemp, realpath, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { startTestServer, type TestServer } from './server'

let server: TestServer

test.beforeAll(async () => {
  server = await startTestServer()
})

test.afterAll(async () => {
  await server?.stop()
})

test('runs organize preview and execution against real filesystem fixtures', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createOrganizeFixture()
  try {
    const organizeRequests: string[] = []
    page.on('request', (request) => {
      const url = new URL(request.url())
      if (url.pathname.startsWith('/api/organize/')) {
        organizeRequests.push(url.pathname)
      }
    })

    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)
    await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()

    await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
    await page.getByRole('button', { name: 'Create Dry-run Preview' }).click()

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Metadata found', '1')
    await expectSummaryValue(page, 'Planned moves', '1')
    await expectSummaryValue(page, 'Warnings', '2')
    await expect(page.getByText(fixture.missingDir)).toBeVisible()
    await expect(page.getByText(fixture.expectedDir)).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(false)
    expect(organizeRequests).toContain('/api/organize/preview')

    await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
    await page.getByRole('button', { name: 'Review Preview & Continue' }).click()
    await expect(page.getByRole('heading', { name: 'Execute the reviewed plan' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Run Organize' })).toBeEnabled()

    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run Organize' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect(page.getByText(fixture.expectedLog)).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(true)
    expect(organizeRequests).toContain('/api/organize/run')
  } finally {
    await fixture.cleanup()
  }
})

type OrganizeFixture = {
  sourceDir: string
  outputDir: string
  missingDir: string
  expectedDir: string
  expectedFile: string
  expectedLog: string
  cleanup: () => Promise<void>
}

async function createOrganizeFixture(): Promise<OrganizeFixture> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'source')
  const outputDir = join(root, 'output')
  const bookDir = join(sourceDir, 'fixture-book')
  const missingDir = join(sourceDir, 'missing-metadata')

  await mkdir(bookDir, { recursive: true })
  await mkdir(missingDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await writeFile(
    join(bookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Fixture Book',
      authors: ['Fixture Author'],
      series: ['Fixture Series #1'],
    }),
  )
  await writeFile(join(bookDir, 'audio.mp3'), 'fake audio data')
  await writeFile(join(missingDir, 'orphan.mp3'), 'fake audio data without metadata')

  const expectedDir = join(outputDir, 'Fixture Author', 'Fixture Series', 'Fixture Book')
  const resolvedOutputDir = await realpath(outputDir)

  return {
    sourceDir,
    outputDir,
    missingDir,
    expectedDir,
    expectedFile: join(expectedDir, 'audio.mp3'),
    expectedLog: join(resolvedOutputDir, '.abook-org.log'),
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function mkFixtureRoot(): Promise<string> {
  return mkdtemp(join(tmpdir(), 'abo-web-organize-'))
}

async function pathExists(path: string): Promise<boolean> {
  try {
    await access(path)
    return true
  } catch {
    return false
  }
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
  await expect(
    page.locator(
      `.result-grid.compact >> xpath=./span[normalize-space(.)="${label}"]/following-sibling::strong[1]`,
    ),
  ).toHaveText(value)
}
