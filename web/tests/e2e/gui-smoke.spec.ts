import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { expect, test, type Page } from '@playwright/test'
import { startTestServer, type TestServer } from './server'

let server: TestServer

test.beforeAll(async () => {
  server = await startTestServer()
})

test.afterAll(async () => {
  await server?.stop()
})

test('protects authenticated API endpoints with the session token', async ({ request }) => {
  const unauthenticated = await request.get(`${server.origin}/api/config/initial`)
  expect(unauthenticated.status()).toBe(401)

  const authenticated = await request.get(`${server.origin}/api/config/initial`, {
    headers: { 'X-Audiobook-Organizer-Token': server.token },
  })
  expect(authenticated.ok()).toBe(true)

  const body = await authenticated.json()
  expect(body).toEqual(
    expect.objectContaining({
      host: '127.0.0.1',
      open: false,
    }),
  )
})

test('#187: explains how to recover when the browser opens without the web session token', async ({ page }) => {
  await page.goto(server.origin)

  await expect(
    page.getByText('This web session link is missing its token. Reopen the complete startup URL.'),
  ).toBeVisible()
})

test('renders staged workflows and connects to the local server', async ({ page }) => {
  await loadApp(page)

  await expect(page).toHaveTitle('Audiobook Organizer for Audiobookshelf')
  await expect(page.getByRole('heading', { name: 'Audiobook Organizer' })).toBeVisible()
  await expect(page.getByText(/localhost connected/)).toBeVisible()
  await expect(page.getByText(/config ready/)).toBeVisible()
  await expect(page.getByText(/options ready/)).toBeVisible()
  await expect(page.getByRole('button', { name: /Organize/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /Rename/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /Audiobookshelf/ })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Setup and preview' })).toBeVisible()
  await expect(page.getByRole('button', { name: /Run/ })).toBeDisabled()
})

test('guides an ABS organize setup into the existing advanced controls', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('button', { name: 'Guide Me' }).click()
  const guide = page.getByRole('dialog', { name: 'What would you like to do?' })
  await expect(guide).toBeVisible()
  await guide.getByRole('radio', { name: 'Organize books' }).click()
  await guide.getByRole('button', { name: 'Next' }).click()
  await expect(page.getByRole('heading', { name: 'Where should metadata come from?' })).toBeVisible()
  await page.getByRole('button', { name: 'Audiobookshelf API' }).click()

  await expect(page.getByRole('dialog')).toHaveCount(0)
  await expect(page.getByRole('radio', { name: 'Audiobookshelf metadata' })).toHaveAttribute('aria-checked', 'true')
  await expect(page.getByLabel('ABS server URL')).toBeVisible()
  await expect(page.getByText(/Next: enter the ABS server URL and token/)).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
})

test('guides an unsure local source to the safe metadata fallback', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('button', { name: 'Guide Me' }).click()
  await page.getByRole('dialog').getByRole('button', { name: 'Next' }).click()
  await page.getByRole('button', { name: 'I am not sure' }).click()
  await expect(page.getByRole('heading', { name: 'One quick question' })).toBeVisible()
  await page.getByRole('button', { name: 'No or unsure' }).click()

  await expect(page.getByRole('dialog')).toHaveCount(0)
  await expect(page.getByRole('radio', { name: 'metadata.json' })).toHaveAttribute('aria-checked', 'true')
  await expect(page.getByText(/safe preview tries metadata.json first, then embedded file metadata/)).toBeVisible()
})

test('offers Audiobookshelf metadata for guided rename setup', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('button', { name: 'Guide Me' }).click()
  await page.getByRole('dialog').getByRole('radio', { name: 'Rename files' }).click()
  await page.getByRole('dialog').getByRole('button', { name: 'Next' }).click()
  await page.getByRole('button', { name: 'Audiobookshelf API' }).click()

  await expect(page.getByRole('dialog')).toHaveCount(0)
  await expect(page.getByRole('radio', { name: 'Audiobookshelf metadata' })).toHaveAttribute('aria-checked', 'true')
  await expect(page.getByLabel('ABS server URL')).toBeVisible()
})

test('uses backend bootstrap options and offers ABS metadata for organize and rename', async ({ page }) => {
  await loadApp(page)

  await expect(page.locator('select[aria-label="Layout"] option[value="author-series-title-number"]')).toHaveCount(1)
  await expect(page.locator('select[aria-label="Layout"] option[value="custom"]')).toHaveCount(1)
  await expect(page.getByRole('radio', { name: 'Audiobookshelf metadata' })).toHaveCount(1)
  await expect(page.getByLabel('Use embedded metadata')).toHaveCount(0)
  await expect(page.getByLabel('Preview color legend')).toBeVisible()

  await page.getByRole('button', { name: /Rename/ }).click()
  await expect(page.getByRole('radio', { name: 'Audiobookshelf metadata' })).toHaveCount(1)
})

test('creates organize previews from configure and derives embedded mode from metadata source', async ({ page }) => {
  let previewBody: Record<string, any> | undefined
  const tempDir = await mkdtemp(join(tmpdir(), 'abo-config-preview-'))
  const sourceDir = join(tempDir, 'source')
  const outputDir = join(tempDir, 'output')

  await page.route('**/api/organize/preview', async (route) => {
    previewBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        summary: {
          MetadataFound: [],
          MetadataMissing: [],
          Moves: [],
          EmptyDirsRemoved: [],
        },
      }),
    })
  })

  try {
    await mkdir(sourceDir)

    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(sourceDir)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(outputDir)
    await page.getByRole('radio', { name: 'Embedded metadata by file' }).click()
    await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
    expect(previewBody).toEqual(
      expect.objectContaining({
        config: expect.objectContaining({
          base_dir: sourceDir,
          output_dir: outputDir,
          use_embedded_metadata: true,
          flat: true,
        }),
      }),
    )
  } finally {
    await rm(tempDir, { recursive: true, force: true })
  }
})

test('defaults from missing metadata.json to embedded file metadata for automatic previews', async ({ page }) => {
  const previewBodies: Record<string, any>[] = []

  await mockValidPathValidation(page)
  await page.route('**/api/organize/preview', async (route) => {
    const body = route.request().postDataJSON()
    previewBodies.push(body)
    const firstRequest = previewBodies.length === 1
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        summary: {
          MetadataFound: firstRequest ? [] : ['/repo/books-dev/input/beef/The Magician.m4b'],
          MetadataMissing: firstRequest ? ['/repo/books-dev/input/beef'] : [],
          Moves: firstRequest
            ? []
            : [
                {
                  from: '/repo/books-dev/input/beef/The Magician.m4b',
                  to: '/repo/books-dev/output/C. S. Lewis/The Chronicles of Narnia/01 - The Magician.m4b',
                },
              ],
          EmptyDirsRemoved: [],
        },
      }),
    })
  })

  await loadApp(page)
  await page.getByRole('textbox', { name: 'Source folder' }).fill('./books-dev/input')
  await page.getByRole('textbox', { name: 'Output folder' }).fill('./books-dev/output')

  await expect(page.getByRole('radio', { name: 'Embedded metadata by file' })).toHaveAttribute('aria-checked', 'true')
  await expect(page.locator('.move-list').filter({ hasText: './books-dev/input/beef/The Magician.m4b' })).toBeVisible()
  await expect(page.locator('.move-list').filter({ hasText: './books-dev/output/C. S. Lewis' })).toBeVisible()
  expect(previewBodies).toHaveLength(2)
  expect(previewBodies[0].config).toEqual(expect.objectContaining({ use_embedded_metadata: false, flat: false }))
  expect(previewBodies[1].config).toEqual(expect.objectContaining({ use_embedded_metadata: true, flat: true }))
})

test('keeps invalid configure paths on the first step before preview requests run', async ({ page }) => {
  const tempDir = await mkdtemp(join(tmpdir(), 'abo-config-invalid-'))
  const filePath = join(tempDir, 'not-a-directory')
  let previewRequested = false

  await page.route('**/api/organize/preview', async (route) => {
    previewRequested = true
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'preview should not run' }),
    })
  })

  try {
    await writeFile(filePath, 'not a directory')
    await loadApp(page)
    await page.getByRole('textbox', { name: 'Source folder' }).fill(filePath)
    await page.getByRole('textbox', { name: 'Output folder' }).fill(join(tempDir, 'output'))
    await expect(page.getByRole('heading', { name: 'Setup and preview' })).toBeVisible()
    await expect(page.locator('.configure-path-alert').filter({ hasText: 'Path is not a directory' })).toBeVisible()
    await expect(page.locator('.path-message').filter({ hasText: 'Path is not a directory' })).toBeVisible()
    expect(previewRequested).toBe(false)
  } finally {
    await rm(tempDir, { recursive: true, force: true })
  }
})

test('supports folder picker and drop affordances while preserving manual path entry', async ({ page }) => {
  await loadApp(page)

  const sourceInput = page.getByRole('textbox', { name: 'Source folder' })
  const outputInput = page.getByRole('textbox', { name: 'Output folder' })
  await sourceInput.fill('/manual/source')
  await outputInput.fill('/manual/output')

  await expect(page.getByRole('button', { name: 'Choose source folder' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Choose output folder' })).toBeVisible()

  const pickerDir = await mkdtemp(join(tmpdir(), 'abo-picker-'))
  await writeFile(join(pickerDir, 'fixture.m4b'), 'fixture')
  await page.getByLabel('Output folder directory picker').setInputFiles(pickerDir)
  await expect(
    page.locator('.path-message').filter({ hasText: 'Folder selected, but this browser did not expose a local path' }),
  ).toBeVisible()
  await expect(outputInput).toHaveValue('/manual/output')

  const dataTransfer = await page.evaluateHandle(() => {
    const transfer = new DataTransfer()
    const file = new File(['fixture'], 'book.m4b', { type: 'audio/mp4' })
    Object.defineProperty(file, 'path', { value: '/dropped/source/book.m4b' })
    transfer.items.add(file)
    return transfer
  })
  await page.locator('[data-path-field="source"]').dispatchEvent('drop', { dataTransfer })
  await expect(sourceInput).toHaveValue('/dropped/source')
  await expect(page.locator('.path-message').filter({ hasText: 'Source folder set from dropped folder.' })).toBeVisible()

  await sourceInput.fill('/typed/source')
  await expect(page.locator('.path-message').filter({ hasText: 'Source folder set from dropped folder.' })).toHaveCount(0)
})

test('keeps staged workflows usable without document overflow on narrow viewports', async ({ page }) => {
  await page.setViewportSize({ width: 393, height: 851 })
  await loadApp(page)

  await expectNoDocumentOverflow(page)
  await expect(page.getByRole('button', { name: /Organize/ })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeVisible()
  await expect(page.getByRole('textbox', { name: 'Source folder' })).toBeVisible()
  await expect(page.getByRole('textbox', { name: 'Output folder' })).toBeVisible()
  await expectNoDocumentOverflow(page)

  await page.getByRole('button', { name: /Rename/ }).click()
  await expect(page.getByRole('textbox', { name: 'Filename template' })).toBeVisible()
  await expectNoDocumentOverflow(page)

  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await expect(page.getByRole('textbox', { name: 'ABS server URL' })).toBeVisible()
  await expect(page.getByRole('textbox', { name: 'ABS path prefix' })).toBeVisible()
  await expect(page.getByText('Setup incomplete')).toBeVisible()
  await expectNoDocumentOverflow(page)

  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
  await expectNoDocumentOverflow(page)
})

test('shows bootstrap fallback state when config and options fail', async ({ page }) => {
  await page.route('**/api/config/initial', async (route) => {
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'config unavailable' }),
    })
  })
  await page.route('**/api/config/options', async (route) => {
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'options unavailable' }),
    })
  })

  await loadApp(page, { allowFailedResourceMessages: true })

  await expect(page.getByText(/config fallback/)).toBeVisible()
  await expect(page.getByText(/options fallback/)).toBeVisible()
  await expect(page.locator('.event-row').filter({ hasText: 'Config unavailable' })).toHaveCount(1)
  await expect(page.locator('.event-row').filter({ hasText: 'Options unavailable' })).toHaveCount(1)
  await expect(page.locator('select[aria-label="Layout"] option')).toHaveText(['Author / Series / Title', 'Custom'])
  await expect(page.getByRole('radio', { name: 'Audiobookshelf metadata' })).toHaveCount(1)
})

test('does not treat redacted ABS credentials as usable browser state', async ({ page }) => {
  await page.route('**/api/config/initial', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        initial: { input_dir: '', output_dir: '' },
        organizer: { layout: 'author-series-title', use_embedded_metadata: false, remove_empty: false },
        rename: {
          template: '{author} - {series} {series_number} - {title}',
          recursive: true,
          preserve_path: true,
        },
        abs: { url: 'http://localhost:13378', token: 'redacted', library_id: 'main', all_libraries: false },
      }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()

  await expect(page.locator('.deferred-state').filter({ hasText: /Saved ABS credentials are redacted/ })).toBeVisible()
})

test('contracts ABS setup controls with mocked backend responses', async ({ page }) => {
  let librariesBody: Record<string, any> | undefined
  let pathBody: Record<string, any> | undefined

  await page.route('**/api/abs/libraries', async (route) => {
    librariesBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        libraries: [
          {
            id: 'lib-audio',
            name: 'Audiobooks',
            mediaType: 'book',
            folders: [{ id: 'folder-audio', path: '/audiobooks', fullPath: '/audiobooks' }],
          },
        ],
      }),
    })
  })
  await page.route('**/api/abs/test-paths', async (route) => {
    pathBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        mappings: [{ abs_prefix: '/audiobooks', local_prefix: '/host/audiobooks' }],
      }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/host/audiobooks')
  await page.getByLabel('ABS server URL').fill('http://localhost:13378')
  await page.getByLabel('ABS API token').fill('test-token')
  await expect(page.getByLabel('ABS library')).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Validate Paths' })).toBeDisabled()
  await page.getByLabel('Local path prefix').fill('/host/audiobooks')
  await page.getByRole('button', { name: 'Test Connection' }).click()

  await expect(page.getByLabel('ABS library')).toBeVisible()
  await expect(page.getByLabel('ABS library')).toHaveValue('lib-audio')
  await expect(page.locator('.library-option.selected').filter({ hasText: 'Audiobooks' })).toBeVisible()
  expect(librariesBody).toEqual(
    expect.objectContaining({
      url: 'http://localhost:13378',
      token: 'test-token',
      library_id: '',
    }),
  )

  await page.getByRole('button', { name: 'Validate Paths' }).click()

  await expect(page.getByText('ABS libraries loaded and path mappings validated.')).toBeVisible()
  await expect(page.getByText('/host/audiobooks')).toBeVisible()
  expect(pathBody).toEqual(
    expect.objectContaining({
      input_dir: '/host/audiobooks',
      config: expect.objectContaining({
        library_id: 'lib-audio',
        path_mappings: [{ abs_prefix: '/audiobooks', local_prefix: '/host/audiobooks' }],
      }),
    }),
  )

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('heading', { name: 'Review ABS Data' })).toBeVisible()
  await expect(page.getByText('Ready for ABS operations')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeEnabled()
})

test('contracts ABS operation controls with mocked backend responses', async ({ page }) => {
  let itemsBody: Record<string, any> | undefined
  let libraryStateBody: Record<string, any> | undefined
  let scanBody: Record<string, any> | undefined
  let cleanBody: Record<string, any> | undefined

  await page.route('**/api/abs/libraries', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        libraries: [
          {
            id: 'lib-audio',
            name: 'Audiobooks',
            mediaType: 'book',
            folders: [{ id: 'folder-audio', path: '/audiobooks', fullPath: '/audiobooks' }],
          },
        ],
      }),
    })
  })
  await page.route('**/api/abs/test-paths', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        mappings: [{ abs_prefix: '/audiobooks', local_prefix: '/host/audiobooks' }],
      }),
    })
  })
  await page.route('**/api/abs/items', async (route) => {
    itemsBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          {
            title: 'Mapped Book',
            authors: ['ABS Author'],
            series: [],
            source_type: 'abs',
            source_path: '/host/audiobooks/Mapped Book',
          },
          {
            title: 'Second Book',
            authors: ['ABS Author'],
            series: [],
            source_type: 'abs',
            source_path: '/host/audiobooks/Second Book',
          },
        ],
      }),
    })
  })
  await page.route('**/api/abs/library-state', async (route) => {
    libraryStateBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        library_id: 'lib-audio',
        items: [
          {
            id: 'item-ok',
            path: '/audiobooks/Mapped Book',
            rel_path: 'Mapped Book',
            is_missing: false,
            is_invalid: false,
            media_type: 'book',
            title: 'Mapped Book',
          },
          {
            id: 'item-missing',
            path: '/audiobooks/Missing Book',
            rel_path: 'Missing Book',
            is_missing: true,
            is_invalid: true,
            media_type: 'book',
            title: 'Missing Book',
          },
        ],
      }),
    })
  })
  await page.route('**/api/abs/scan-trigger', async (route) => {
    scanBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ triggered: true, library_id: 'lib-audio' }),
    })
  })
  await page.route('**/api/abs/clean-missing', async (route) => {
    cleanBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ cleaned: true, library_id: 'lib-audio' }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/host/audiobooks')
  await page.getByLabel('ABS server URL').fill('http://localhost:13378')
  await page.getByLabel('ABS API token').fill('test-token')
  await page.getByLabel('Local path prefix').fill('/host/audiobooks')
  await page.getByRole('button', { name: 'Test Connection' }).click()
  await expect(page.getByLabel('ABS library')).toHaveValue('lib-audio')
  await page.getByRole('button', { name: 'Validate Paths' }).click()

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('button', { name: 'Load ABS Items' })).toBeEnabled()
  await expect(page.getByRole('button', { name: 'Check Library State' })).toBeEnabled()
  await page.getByRole('button', { name: 'Load ABS Items' }).click()
  await page.getByRole('button', { name: 'Check Library State' }).click()

  await expect(page.getByRole('heading', { name: 'ABS Operation Results' })).toBeVisible()
  await expect(page.getByText('/host/audiobooks/Mapped Book')).toBeVisible()
  await expect(page.locator('.move-list em').filter({ hasText: 'Missing' })).toBeVisible()
  expect(itemsBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))
  expect(libraryStateBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByText('Scan triggered for lib-audio.')).toHaveCount(0)
  await page.getByRole('button', { name: 'Trigger Scan' }).click()
  await expect(page.getByText('Scan triggered for lib-audio.')).toBeVisible()
  expect(scanBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))

  await expect(page.getByRole('button', { name: 'Clean Missing Items' })).toBeDisabled()
  expect(cleanBody).toBeUndefined()
  await page.getByLabel('I understand this removes ABS missing item records').check()
  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Clean missing ABS item records')
    await dialog.accept()
  })
  await page.getByRole('button', { name: 'Clean Missing Items' }).click()
  await expect(page.getByText('Cleanup completed for lib-audio.')).toBeVisible()
  expect(cleanBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))

  await page.getByRole('button', { name: 'Review & Run Select, execute, inspect' }).click()
  await expect(page.getByRole('heading', { name: 'ABS Operation Results' })).toBeVisible()
  await expect(page.locator('.review-layout .result-grid strong').filter({ hasText: 'lib-audio' })).toHaveCount(2)
})

test('keeps ABS later stages locked when setup requests fail', async ({ page }) => {
  await page.route('**/api/abs/libraries', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'abs connection failed' }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await page.getByLabel('ABS server URL').fill('http://localhost:13378')
  await page.getByLabel('ABS API token').fill('bad-token')
  await page.getByRole('button', { name: 'Test Connection' }).click()

  await expect(page.locator('.inline-alert').filter({ hasText: 'abs connection failed' })).toBeVisible()
  await expect(page.getByLabel('ABS library')).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Validate Paths' })).toBeDisabled()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
  await expect(page.locator('.inline-alert').filter({ hasText: 'abs connection failed' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
})

test('contracts organize preview and run UI state with mocked backend responses', async ({ page }) => {
  let previewBody: Record<string, any> | undefined
  let runBody: Record<string, any> | undefined

  await mockValidPathValidation(page)
  await page.route('**/api/organize/preview', async (route) => {
    previewBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        summary: {
          MetadataFound: ['/library/source/book/metadata.json'],
          MetadataMissing: ['/library/source/missing'],
          Moves: [{ from: '/library/source/book/audio.mp3', to: '/library/output/Author/Book/audio.mp3' }],
          EmptyDirsRemoved: [],
        },
      }),
    })
  })
  await page.route('**/api/organize/run', async (route) => {
    runBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        summary: {
          MetadataFound: ['/library/source/book/metadata.json'],
          MetadataMissing: [],
          Moves: [{ from: '/library/source/book/audio.mp3', to: '/library/output/Author/Book/audio.mp3' }],
          EmptyDirsRemoved: [],
        },
        log_path: '/library/output/.abook-org.log',
      }),
    })
  })

  await loadApp(page)
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await page.getByRole('textbox', { name: 'Output folder' }).fill('/library/output')

  await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
  await expect(page.getByText('/library/source/missing')).toBeVisible()
  await expect(page.locator('.move-list').filter({ hasText: '/library/output/Author/Book/audio.mp3' })).toBeVisible()
  expect(previewBody?.config).toEqual(
    expect.objectContaining({
      base_dir: '/library/source',
      output_dir: '/library/output',
      dry_run: true,
      layout: 'author-series-title',
    }),
  )

  await page.getByRole('combobox', { name: 'Layout' }).selectOption('custom')
  await page.getByRole('textbox', { name: 'Custom layout template' }).fill('{author}/{title}')
  await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
  expect(previewBody?.config).toEqual(expect.objectContaining({ layout_template: '{author}/{title}' }))

  await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Reviewed Organize Plan' })).toBeVisible()
  await expect(page.locator('.reviewed-plan').filter({ hasText: '/library/output/Author/Book/audio.mp3' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Reviewed Organize Plan' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run 1 Selected Move' })).toBeEnabled()

  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Run Organize will change files for 1 selected move')
    await dialog.accept()
  })
  await page.getByRole('button', { name: 'Run 1 Selected Move' }).click()

  await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
  await expect(
    page.locator('.review-layout .result-grid strong').filter({ hasText: '/library/output/.abook-org.log' }),
  ).toBeVisible()
  await expect(page.locator('.review-layout').filter({ hasText: '/library/output/Author/Book/audio.mp3' })).toBeVisible()
  expect(runBody?.config).toEqual(
    expect.objectContaining({
      dry_run: false,
      allowed_source_paths: ['/library/source/book/audio.mp3'],
    }),
  )
})

test('keeps organize run locked when preview fails', async ({ page }) => {
  let runRequested = false

  await mockValidPathValidation(page)
  await page.route('**/api/organize/preview', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'preview exploded' }),
    })
  })
  await page.route('**/api/organize/run', async (route) => {
    runRequested = true
    await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: 'run leaked' }) })
  })

  await loadApp(page)
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await page.getByRole('textbox', { name: 'Output folder' }).fill('/library/output')
  await expect(page.locator('.inline-alert').filter({ hasText: 'preview exploded' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
  expect(runRequested).toBe(false)
})

test('contracts rename preview and run UI state with mocked backend responses', async ({ page }) => {
  let previewBody: Record<string, any> | undefined
  let runBody: Record<string, any> | undefined

  await mockValidPathValidation(page)
  await page.route('**/api/rename/preview', async (route) => {
    previewBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        candidates: [
          {
            CurrentPath: '/library/source/book/audio.mp3',
            ProposedPath: '/library/source/book/Rename Author - Rename Book.mp3',
            Metadata: { title: 'Rename Book', authors: ['Rename Author'] },
            IsNoOp: false,
            IsConflict: false,
            Error: '',
          },
          {
            CurrentPath: '/library/source/book/duplicate.mp3',
            ProposedPath: '/library/source/book/Rename Author - Rename Book (1).mp3',
            Metadata: { title: 'Rename Book', authors: ['Rename Author'] },
            IsNoOp: false,
            IsConflict: true,
            Error: '',
          },
        ],
        summary: {
          FilesScanned: 3,
          FilesRenamed: 0,
          FilesSkipped: 1,
          ConflictsFound: 1,
          Errors: ['missing metadata in skipped.mp3'],
        },
      }),
    })
  })
  await page.route('**/api/rename/run', async (route) => {
    runBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        candidates: [
          {
            CurrentPath: '/library/source/book/audio.mp3',
            ProposedPath: '/library/source/book/Rename Author - Rename Book.mp3',
            Metadata: { title: 'Rename Book', authors: ['Rename Author'] },
            IsNoOp: false,
            IsConflict: false,
            Error: '',
          },
          {
            CurrentPath: '/library/source/book/duplicate.mp3',
            ProposedPath: '/library/source/book/Rename Author - Rename Book (1).mp3',
            Metadata: { title: 'Rename Book', authors: ['Rename Author'] },
            IsNoOp: false,
            IsConflict: true,
            Error: '',
          },
        ],
        summary: {
          FilesScanned: 3,
          FilesRenamed: 2,
          FilesSkipped: 1,
          ConflictsFound: 1,
          Errors: ['missing metadata in skipped.mp3'],
        },
        log_path: '/library/source/.abook-rename.log',
      }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Rename/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await page.getByRole('textbox', { name: 'Filename template' }).fill('{author} - {title}')

  await expect(page.getByRole('heading', { name: 'Rename preview ready' })).toBeVisible()
  await expect(page.locator('.move-list').filter({ hasText: '/library/source/book/Rename Author - Rename Book.mp3' })).toBeVisible()
  await expect(page.locator('.move-list em').filter({ hasText: /^Conflict$/ })).toBeVisible()
  await expect(page.getByText('missing metadata in skipped.mp3')).toBeVisible()
  expect(previewBody?.config).toEqual(
    expect.objectContaining({
      base_dir: '/library/source',
      template: '{author} - {title}',
      dry_run: true,
      recursive: true,
      preserve_path: true,
      use_embedded_metadata: false,
    }),
  )

  await page.getByRole('button', { name: 'Review & Run', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Reviewed Rename Plan' })).toBeVisible()
  await expect(page.locator('.reviewed-plan').filter({ hasText: '/library/source/book/Rename Author - Rename Book.mp3' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Review and run' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Reviewed Rename Plan' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run 2 Selected Files' })).toBeEnabled()

  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Run Rename will change 2 selected file')
    await dialog.accept()
  })
  await page.getByRole('button', { name: 'Run 2 Selected Files' }).click()

  await expect(page.getByRole('heading', { name: 'Rename Run Complete' })).toBeVisible()
  await expect(
    page.locator('.review-layout .result-grid strong').filter({ hasText: '/library/source/.abook-rename.log' }),
  ).toBeVisible()
  await expect(page.locator('.review-layout').filter({ hasText: '/library/source/book/Rename Author - Rename Book.mp3' })).toBeVisible()
  expect(runBody?.config).toEqual(
    expect.objectContaining({
      dry_run: false,
      allowed_current_paths: ['/library/source/book/audio.mp3', '/library/source/book/duplicate.mp3'],
    }),
  )
})

test('keeps rename run unavailable when preview fails', async ({ page }) => {
  let runRequested = false

  await mockValidPathValidation(page)
  await page.route('**/api/rename/preview', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'rename preview exploded' }),
    })
  })
  await page.route('**/api/rename/run', async (route) => {
    runRequested = true
    await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: 'run leaked' }) })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Rename/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await expect(page.locator('.inline-alert').filter({ hasText: 'rename preview exploded' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
  expect(runRequested).toBe(false)
})

test('separates workflow modes and gates run behind preview review', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('button', { name: /Rename/ }).click()
  await expect(page.getByRole('heading', { name: 'Setup and preview' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Rename Template' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Organization Rules' })).toHaveCount(0)

  await expect(page.getByRole('heading', { name: 'Waiting for rename inputs' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Review & Run Select, execute, inspect' })).toBeDisabled()
})

async function loadApp(page: Page, options: { allowFailedResourceMessages?: boolean } = {}): Promise<void> {
  const consoleMessages: string[] = []
  page.on('console', (message) => {
    if (['error', 'warning'].includes(message.type())) {
      const text = message.text()
      if (options.allowFailedResourceMessages && text.includes('Failed to load resource')) {
        return
      }
      consoleMessages.push(`${message.type()}: ${text}`)
    }
  })

  await page.goto(server.url)
  await expect(page.locator('#app')).not.toBeEmpty()

  expect(consoleMessages, 'No browser console warnings/errors during initial render').toEqual([])
}

async function mockValidPathValidation(page: Page): Promise<void> {
  await page.route('**/api/paths/validate', async (route) => {
    const body = route.request().postDataJSON() as { paths?: Array<{ id: string; path: string }> }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        results: (body.paths ?? []).map((item) => ({ id: item.id, path: item.path, valid: true })),
      }),
    })
  })
}

async function expectNoDocumentOverflow(page: Page): Promise<void> {
  const overflow = await page.evaluate(() => {
    const viewportWidth = document.documentElement.clientWidth
    const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth)
    const overflowers = Array.from(document.querySelectorAll<HTMLElement>('body *'))
      .filter((element) => element.scrollWidth > element.clientWidth + 1)
      .slice(0, 8)
      .map((element) => ({
        tag: element.tagName.toLowerCase(),
        className: element.className.toString(),
        ariaLabel: element.getAttribute('aria-label'),
        text: element.textContent?.replace(/\s+/g, ' ').trim().slice(0, 80),
        clientWidth: element.clientWidth,
        scrollWidth: element.scrollWidth,
      }))

    return { viewportWidth, documentWidth, overflowers }
  })

  expect(overflow.documentWidth, JSON.stringify(overflow.overflowers, null, 2)).toBeLessThanOrEqual(
    overflow.viewportWidth + 1,
  )
}
