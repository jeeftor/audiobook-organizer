import { expect, test, type Page } from '@playwright/test'
import { access, copyFile, mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises'
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

test('previews and executes real rename candidates through the web UI', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createRenameFixture()
  try {
    const renameRequests: string[] = []
    page.on('request', (request) => {
      const url = new URL(request.url())
      if (url.pathname.startsWith('/api/rename/')) {
        renameRequests.push(url.pathname)
      }
    })

    await loadApp(page)
    await page.getByRole('button', { name: /Rename/ }).click()
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Rename template' }).fill('{author} - {title}')
    await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()

    await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
    await page.getByRole('button', { name: 'Create Rename Preview' }).click()

    await expect(page.getByRole('heading', { name: 'Rename preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Files scanned', '4')
    await expectSummaryValue(page, 'Candidates', '4')
    await expectSummaryValue(page, 'Conflicts', '1')
    await expectSummaryValue(page, 'Skipped', '2')
    await expectSummaryValue(page, 'Errors', '1')
    await expect(page.getByText(fixture.firstProposedPath)).toBeVisible()
    await expect(page.getByText(fixture.conflictProposedPath)).toBeVisible()
    await expect(page.locator('.move-list em').filter({ hasText: /^Conflict$/ })).toBeVisible()
    await expect(page.locator('.move-list em').filter({ hasText: 'Skipped: unchanged' })).toBeVisible()
    await expect(page.locator('.warning-list li').filter({ hasText: /Failed to extract metadata/ })).toBeVisible()
    await page.getByRole('checkbox', { name: `Select rename candidate ${fixture.conflictOriginalPath}` }).uncheck()
    await expectSummaryValue(page, 'Selected files', '1')
    expect(renameRequests).toContain('/api/rename/preview')
    await expectPathExists(fixture.firstOriginalPath)
    await expectPathExists(fixture.conflictOriginalPath)
    await expectPathMissing(fixture.firstProposedPath)
    await expectPathMissing(fixture.conflictProposedPath)

    await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
    await page.getByRole('button', { name: 'Review Candidates & Continue' }).click()
    await expect(page.getByRole('heading', { name: 'Execute the reviewed plan' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Reviewed Rename Plan' })).toBeVisible()
    await expect(page.locator('.reviewed-plan').getByText(fixture.firstProposedPath)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Run 1 Selected File' })).toBeEnabled()

    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Rename will change 1 selected file')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected File' }).click()

    await expect(page.getByRole('heading', { name: 'Rename Run Complete' })).toBeVisible()
    await expectReviewSummaryValue(page, 'Job status', 'Complete')
    await expectReviewSummaryValue(page, 'Files scanned', '1')
    await expectReviewSummaryValue(page, 'Files renamed', '1')
    await expectReviewSummaryValue(page, 'Conflicts', '0')
    await expectReviewSummaryValue(page, 'Skipped', '0')
    await expectReviewSummaryValue(page, 'Errors', '0')
    await expectReviewSummaryValue(page, 'Undo log', fixture.logPath)
    await expect(page.locator('.review-layout .recovery-note')).toContainText(fixture.logPath)
    await expect(page.locator('.event-row').filter({ hasText: 'Request started: Rename preview' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Request succeeded: Rename preview' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Local review: Rename candidates accepted' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Request started: Rename run' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Request succeeded: Rename run' })).toHaveCount(1)
    expect(renameRequests).toContain('/api/rename/run')
    await expectPathMissing(fixture.firstOriginalPath)
    await expectPathExists(fixture.conflictOriginalPath)
    await expectPathExists(fixture.firstProposedPath)
    await expectPathMissing(fixture.conflictProposedPath)
    await expectPathExists(fixture.noopPath)
    await expectPathExists(fixture.brokenPath)
    const log = await readFile(fixture.logPath, 'utf8')
    expect(log).toContain(fixture.firstOriginalPath)
    expect(log).not.toContain(fixture.conflictOriginalPath)
  } finally {
    await fixture.cleanup()
  }
})

type RenameFixture = {
  sourceDir: string
  firstOriginalPath: string
  conflictOriginalPath: string
  firstProposedPath: string
  conflictProposedPath: string
  noopPath: string
  brokenPath: string
  logPath: string
  cleanup: () => Promise<void>
}

async function createRenameFixture(): Promise<RenameFixture> {
  const root = await mkdtemp(join(tmpdir(), 'abo-web-rename-'))
  const sourceAudio = join(repoRoot(), 'testdata', 'mp3flat', 'charlesdexterward_01_lovecraft_64kb.mp3')
  const conflictADir = join(root, '01-conflict-a')
  const conflictBDir = join(root, '02-conflict-b')
  const noopDir = join(root, '03-noop')
  const brokenDir = join(root, '04-broken')

  await createRenameBook(root, '01-conflict-a', 'original-a.mp3', sourceAudio, {
    title: 'Conflict Book',
    authors: ['Conflict Author'],
    series: ['Rename Series #1'],
  })
  await createRenameBook(root, '02-conflict-b', 'original-b.mp3', sourceAudio, {
    title: 'Conflict Book',
    authors: ['Conflict Author'],
    series: ['Rename Series #1'],
  })
  await createRenameBook(root, '03-noop', 'Noop Author - Noop Book.mp3', sourceAudio, {
    title: 'Noop Book',
    authors: ['Noop Author'],
  })

  await mkdir(brokenDir, { recursive: true })
  await writeFile(join(brokenDir, 'broken.mp3'), 'not audio')

  return {
    sourceDir: root,
    firstOriginalPath: join(conflictADir, 'original-a.mp3'),
    conflictOriginalPath: join(conflictBDir, 'original-b.mp3'),
    firstProposedPath: join(conflictADir, 'Conflict Author - Conflict Book.mp3'),
    conflictProposedPath: join(conflictBDir, 'Conflict Author - Conflict Book (2).mp3'),
    noopPath: join(noopDir, 'Noop Author - Noop Book.mp3'),
    brokenPath: join(brokenDir, 'broken.mp3'),
    logPath: join(root, '.abook-rename.log'),
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function createRenameBook(
  root: string,
  dirName: string,
  audioName: string,
  sourceAudio: string,
  metadata: Record<string, unknown>,
): Promise<void> {
  const bookDir = join(root, dirName)
  await mkdir(bookDir, { recursive: true })
  await copyFile(sourceAudio, join(bookDir, audioName))
  await writeFile(join(bookDir, 'metadata.json'), JSON.stringify(metadata))
}

function repoRoot(): string {
  return new URL('../../..', import.meta.url).pathname
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

async function expectReviewSummaryValue(page: Page, label: string, value: string): Promise<void> {
  await expect(
    page.locator(`.review-layout .result-grid >> xpath=./span[normalize-space(.)="${label}"]/following-sibling::strong[1]`),
  ).toHaveText(value)
}

async function expectPathExists(path: string): Promise<void> {
  expect(await pathExists(path)).toBe(true)
}

async function expectPathMissing(path: string): Promise<void> {
  expect(await pathExists(path)).toBe(false)
}

async function pathExists(path: string): Promise<boolean> {
  return access(path).then(
    () => true,
    () => false,
  )
}
