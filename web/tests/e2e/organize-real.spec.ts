import { expect, test, type Page } from '@playwright/test'
import { access, copyFile, mkdir, mkdtemp, realpath, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { startTestServer, type TestServer } from './server'

const repoRoot = new URL('../../..', import.meta.url).pathname

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
    const organizeRequests = collectOrganizeRequests(page)

    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Metadata found', '1')
    await expectSummaryValue(page, 'Planned moves', '1')
    await expectSummaryValue(page, 'Warnings', '2')
    await expect(page.getByText(fixture.missingDir)).toBeVisible()
    await expect(page.getByText(fixture.expectedDir)).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(false)
    expect(organizeRequests).toContain('/api/organize/preview')

    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Reviewed Organize Plan' })).toBeVisible()
    await expect(page.locator('.reviewed-plan').getByText(fixture.expectedDir)).toBeVisible()
    await expect(page.locator('.reviewed-plan .warning-list').getByText(fixture.missingDir)).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Run 1 Selected Move' })).toBeEnabled()

    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.dismiss()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()
    await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(false)
    await expect(page.locator('.event-row').filter({ hasText: 'Request started: Organize run' })).toHaveCount(0)
    expect(organizeRequests).not.toContain('/api/organize/run')

    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect(page.locator('.review-layout .result-grid strong').filter({ hasText: fixture.expectedLog })).toBeVisible()
    await expect(page.getByText(`Undo details are available in the backend log at ${fixture.expectedLog}.`)).toBeVisible()
    await expectReviewSummaryValue(page, 'Warnings', '0')
    await expect(page.locator('.event-row').filter({ hasText: 'Request started: Organize preview' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Request succeeded: Organize preview' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Request started: Organize run' })).toHaveCount(1)
    await expect(page.locator('.event-row').filter({ hasText: 'Request succeeded: Organize run' })).toHaveCount(1)
    await expect(page.getByText('Waiting for run')).toHaveCount(0)
    await expect(page.getByText('Not created')).toHaveCount(0)
    await expect(page.getByText('None yet')).toHaveCount(0)
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(true)
    await expect.poll(() => pathExists(fixture.expectedLog)).toBe(true)
    expect(organizeRequests).toContain('/api/organize/run')
  } finally {
    await fixture.cleanup()
  }
})

test('runs only selected organize preview rows against real filesystem fixtures', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createOrganizeSelectionFixture()
  try {
    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Planned moves', '2')
    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    await page.locator('.selectable-list input[type="checkbox"]').nth(1).uncheck()
    await expectSummaryValue(page, 'Selected moves', '1')
    await expect(page.getByRole('button', { name: 'Run 1 Selected Move' })).toBeEnabled()
    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect.poll(() => pathExists(fixture.selectedOutputFile)).toBe(true)
    await expect.poll(() => pathExists(fixture.unselectedInputFile)).toBe(true)
    await expect.poll(() => pathExists(fixture.unselectedOutputFile)).toBe(false)
  } finally {
    await fixture.cleanup()
  }
})

test('organizes a real EPUB fixture through embedded metadata mode', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createEmbeddedEPUBFixture()
  try {
    const organizeRequests = collectOrganizeRequests(page)

    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)
    await page.getByRole('radio', { name: 'Embedded metadata by directory' }).click()

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Metadata found', '0')
    await expectSummaryValue(page, 'Planned moves', '1')
    await expectSummaryValue(page, 'Warnings', '1')
    await expect(page.getByText(fixture.expectedDir)).toBeVisible()
    await expect(page.locator('.warning-list').getByText(fixture.sourceDir)).toBeVisible()
    await expect.poll(() => pathExists(fixture.mappedDir)).toBe(false)
    await expect.poll(() => pathExists(fixture.sourceFile)).toBe(true)
    expect(organizeRequests).toContain('/api/organize/preview')

    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect.poll(() => pathExists(fixture.mappedDir)).toBe(true)
    await expect.poll(() => pathExists(fixture.sourceFile)).toBe(false)
    await expect.poll(() => pathExists(fixture.expectedLog)).toBe(true)
    expect(organizeRequests).toContain('/api/organize/run')
  } finally {
    await fixture.cleanup()
  }
})

test('uses numbered layout and removes empty source folders after organize run', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createNumberedLayoutFixture()
  try {
    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)
    await page.getByRole('combobox', { name: 'Layout' }).selectOption('author-series-title-number')
    await page.getByLabel('Remove empty source folders after run').check()

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Metadata found', '1')
    await expectSummaryValue(page, 'Planned moves', '1')
    await expect(page.getByText(fixture.expectedDir)).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(false)
    await expect.poll(() => pathExists(fixture.bookDir)).toBe(true)

    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect(page.locator('.review-layout .result-grid strong').filter({ hasText: fixture.expectedLog })).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(true)
    await expect.poll(() => pathExists(fixture.expectedLog)).toBe(true)
    await expect.poll(() => pathExists(fixture.bookDir)).toBe(false)
  } finally {
    await fixture.cleanup()
  }
})

test('uses a custom metadata field mapping in a real organize preview and execution', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createFieldMappingFixture()
  try {
    await loadApp(page)
    await page.getByRole('combobox', { name: 'Title field mapping' }).fill('alternate_title')
    await page.getByRole('combobox', { name: 'Author field mapping' }).fill('alternate_authors')
    await page.getByRole('combobox', { name: 'Series field mapping' }).fill('alternate_series')
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expect(page.getByText(fixture.mappedDir)).toBeVisible()
    await expect.poll(() => pathExists(fixture.mappedDir)).toBe(false)

    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect.poll(() => pathExists(fixture.mappedDir)).toBe(true)
    await expect.poll(() => pathExists(fixture.sourceFile)).toBe(false)
  } finally {
    await fixture.cleanup()
  }
})

test('uses a custom layout template for real organize preview and execution', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createCustomLayoutTemplateFixture()
  try {
    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)
    await page.getByRole('combobox', { name: 'Layout' }).selectOption('custom')
    await page
      .getByRole('textbox', { name: 'Custom layout template' })
      .fill('{author}/{series}/{series-count} - {title} ({narrator})')

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expectSummaryValue(page, 'Metadata found', '1')
    await expectSummaryValue(page, 'Planned moves', '1')
    await expect(page.getByText(fixture.expectedDir)).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(false)

    await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
      await dialog.accept()
    })
    await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

    await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(true)
    await expect.poll(() => pathExists(fixture.expectedLog)).toBe(true)
  } finally {
    await fixture.cleanup()
  }
})

test('reports real backend path validation errors for organize preview', async ({ page }) => {
  test.setTimeout(60_000)

  const fixture = await createPathErrorFixture()
  try {
    const organizeRequests = collectOrganizeRequests(page)

    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.missingSourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)

    await expect(page.locator('.inline-alert').filter({ hasText: 'Directory does not exist' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()

    await page.getByRole('textbox', { name: 'Source folder' }).fill(fixture.sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.missingOutputDir)

    await expect(page.locator('.inline-alert').filter({ hasText: 'error resolving output directory path' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()

    await page.getByRole('textbox', { name: 'Output folder' }).fill(fixture.outputDir)

    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    await expect(page.locator('.inline-alert')).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeEnabled()
    await expect.poll(() => pathExists(fixture.expectedFile)).toBe(false)
    expect(organizeRequests.filter((path) => path === '/api/organize/preview')).toHaveLength(2)
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

type OrganizeSelectionFixture = {
  sourceDir: string
  outputDir: string
  selectedBookDir: string
  unselectedBookDir: string
  selectedOutputFile: string
  unselectedInputFile: string
  unselectedOutputFile: string
  cleanup: () => Promise<void>
}

type EmbeddedEPUBFixture = OrganizeFixture & {
  sourceFile: string
}

type NumberedLayoutFixture = OrganizeFixture & {
  bookDir: string
}

type PathErrorFixture = {
  sourceDir: string
  outputDir: string
  expectedFile: string
  missingSourceDir: string
  missingOutputDir: string
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

async function createOrganizeSelectionFixture(): Promise<OrganizeSelectionFixture> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'source')
  const outputDir = join(root, 'output')
  const selectedBookDir = join(sourceDir, 'selected-book')
  const unselectedBookDir = join(sourceDir, 'unselected-book')

  await mkdir(selectedBookDir, { recursive: true })
  await mkdir(unselectedBookDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await writeFile(
    join(selectedBookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Selected Book',
      authors: ['Selection Author'],
    }),
  )
  await writeFile(
    join(unselectedBookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Unselected Book',
      authors: ['Selection Author'],
    }),
  )
  await writeFile(join(selectedBookDir, 'selected.mp3'), 'fake audio data')
  await writeFile(join(unselectedBookDir, 'unselected.mp3'), 'fake audio data')

  return {
    sourceDir,
    outputDir,
    selectedBookDir,
    unselectedBookDir,
    selectedOutputFile: join(outputDir, 'Selection Author', 'Selected Book', 'selected.mp3'),
    unselectedInputFile: join(unselectedBookDir, 'unselected.mp3'),
    unselectedOutputFile: join(outputDir, 'Selection Author', 'Unselected Book', 'unselected.mp3'),
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function createEmbeddedEPUBFixture(): Promise<EmbeddedEPUBFixture> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'source')
  const outputDir = join(root, 'output')
  const bookDir = join(sourceDir, 'embedded-book')
  const sourceFile = join(bookDir, 'title-author-series1.epub')

  await mkdir(bookDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await copyFile(join(repoRoot, 'testdata', 'epub', 'title-author-series1.epub'), sourceFile)

  const expectedDir = join(outputDir, 'Jeef of Github,Some random guy', 'Test Books', 'First book of testing knowledge')
  const resolvedOutputDir = await realpath(outputDir)

  return {
    sourceDir,
    outputDir,
    missingDir: '',
    expectedDir,
    expectedFile: join(expectedDir, 'title-author-series1.epub'),
    expectedLog: join(resolvedOutputDir, '.abook-org.log'),
    sourceFile,
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function createNumberedLayoutFixture(): Promise<NumberedLayoutFixture> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'source')
  const outputDir = join(root, 'output')
  const bookDir = join(sourceDir, 'numbered-layout-book')

  await mkdir(bookDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await writeFile(
    join(bookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Numbered Layout Book',
      authors: ['Layout Author'],
      series: ['Layout Series'],
      series_index: 3,
    }),
  )
  await writeFile(join(bookDir, 'audio.mp3'), 'fake audio data')

  const expectedDir = join(outputDir, 'Layout Author', 'Layout Series', '#3 - Numbered Layout Book')
  const resolvedOutputDir = await realpath(outputDir)

  return {
    sourceDir,
    outputDir,
    missingDir: '',
    expectedDir,
    expectedFile: join(expectedDir, 'audio.mp3'),
    expectedLog: join(resolvedOutputDir, '.abook-org.log'),
    bookDir,
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function createCustomLayoutTemplateFixture(): Promise<OrganizeFixture> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'source')
  const outputDir = join(root, 'output')
  const bookDir = join(sourceDir, 'custom-layout-book')

  await mkdir(bookDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await writeFile(
    join(bookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Template Book',
      authors: ['Template Author'],
      series: ['Template Series'],
      series_index: 4,
      narrator: 'Template Narrator',
    }),
  )
  await writeFile(join(bookDir, 'audio.mp3'), 'fake audio data')

  const expectedDir = join(outputDir, 'Template Author', 'Template Series', '4 - Template Book (Template Narrator)')
  const resolvedOutputDir = await realpath(outputDir)

  return {
    sourceDir,
    outputDir,
    missingDir: '',
    expectedDir,
    expectedFile: join(expectedDir, 'audio.mp3'),
    expectedLog: join(resolvedOutputDir, '.abook-org.log'),
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function createPathErrorFixture(): Promise<PathErrorFixture> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'source')
  const outputDir = join(root, 'output')
  const bookDir = join(sourceDir, 'valid-book')

  await mkdir(bookDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await writeFile(
    join(bookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Valid Book',
      authors: ['Valid Author'],
    }),
  )
  await writeFile(join(bookDir, 'audio.mp3'), 'fake audio data')

  return {
    sourceDir,
    outputDir,
    expectedFile: join(outputDir, 'Valid Author', 'Valid Book', 'audio.mp3'),
    missingSourceDir: join(root, 'missing-source'),
    missingOutputDir: join(root, 'missing-output'),
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
}

async function mkFixtureRoot(): Promise<string> {
  return mkdtemp(join(tmpdir(), 'abo-web-organize-'))
}

async function createFieldMappingFixture(): Promise<{
  sourceDir: string
  outputDir: string
  defaultDir: string
  mappedDir: string
  sourceFile: string
  cleanup: () => Promise<void>
}> {
  const root = await mkFixtureRoot()
  const sourceDir = join(root, 'input')
  const outputDir = join(root, 'output')
  const bookDir = join(sourceDir, 'mapped-book')
  const sourceAudio = join(repoRoot, 'testdata', 'mp3flat', 'charlesdexterward_01_lovecraft_64kb.mp3')
  await mkdir(bookDir, { recursive: true })
  await mkdir(outputDir, { recursive: true })
  await copyFile(sourceAudio, join(bookDir, 'book.mp3'))
  await writeFile(
    join(bookDir, 'metadata.json'),
    JSON.stringify({
      title: 'Default Title',
      authors: ['Default Author'],
      series: ['Default Series #1'],
      alternate_title: 'Mapped Title',
      alternate_authors: ['Mapped Author'],
      alternate_series: 'Mapped Series #2',
    }),
  )
  return {
    sourceDir,
    outputDir,
    defaultDir: join(outputDir, 'Default Author', 'Default Series', 'Default Title'),
    mappedDir: join(outputDir, 'Mapped Author', 'Mapped Series', 'Mapped Title'),
    sourceFile: join(bookDir, 'book.mp3'),
    cleanup: () => rm(root, { recursive: true, force: true }),
  }
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

function collectOrganizeRequests(page: Page): string[] {
  const organizeRequests: string[] = []
  page.on('request', (request) => {
    const url = new URL(request.url())
    if (url.pathname.startsWith('/api/organize/')) {
      organizeRequests.push(url.pathname)
    }
  })
  return organizeRequests
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
