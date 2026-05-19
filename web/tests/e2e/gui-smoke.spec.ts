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
  await expect(page.getByRole('heading', { name: 'Configure and scan setup' })).toBeVisible()
  await expect(page.getByRole('button', { name: /Run/ })).toBeDisabled()
})

test('uses backend bootstrap options and scopes ABS scan mode to ABS workflow', async ({ page }) => {
  await loadApp(page)

  await expect(page.locator('select[aria-label="Layout"] option[value="author-series-title-number"]')).toHaveCount(1)
  await expect(page.locator('select[aria-label="Metadata source"] option[value="abs"]')).toHaveCount(0)

  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await expect(page.locator('select[aria-label="Metadata source"] option[value="abs"]')).toHaveCount(1)
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
  await expect(page.locator('select[aria-label="Layout"] option')).toHaveText(['Author / Series / Title'])
  await expect(page.locator('select[aria-label="Metadata source"] option[value="abs"]')).toHaveCount(0)
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
  await page.getByLabel('ABS library ID').fill('pending-library')
  await page.getByLabel('Local path prefix').fill('/host/audiobooks')
  await page.getByRole('button', { name: 'Load Libraries' }).click()

  await expect(page.getByRole('button', { name: /Audiobooks lib-audio/ })).toBeVisible()
  await page.getByRole('button', { name: /Audiobooks lib-audio/ }).click()
  await expect(page.getByLabel('ABS library ID')).toHaveValue('lib-audio')
  expect(librariesBody).toEqual(
    expect.objectContaining({
      url: 'http://localhost:13378',
      token: 'test-token',
      library_id: 'pending-library',
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

  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await expect(page.getByRole('heading', { name: 'ABS Operation Summary' })).toBeVisible()
  await expect(page.getByText('Ready for ABS operations')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeEnabled()
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
  await page.getByLabel('ABS library ID').fill('lib-audio')
  await page.getByLabel('Local path prefix').fill('/host/audiobooks')
  await page.getByRole('button', { name: 'Load Libraries' }).click()
  await page.getByRole('button', { name: 'Validate Paths' }).click()

  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await expect(page.getByRole('button', { name: 'Load ABS Items' })).toBeEnabled()
  await expect(page.getByRole('button', { name: 'Check Library State' })).toBeEnabled()
  await page.getByRole('button', { name: 'Load ABS Items' }).click()
  await page.getByRole('button', { name: 'Check Library State' }).click()

  await expect(page.getByRole('heading', { name: 'ABS operation data ready' })).toBeVisible()
  await expect(page.getByText('/host/audiobooks/Mapped Book')).toBeVisible()
  await expect(page.locator('.move-list em').filter({ hasText: 'Missing' })).toBeVisible()
  expect(itemsBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))
  expect(libraryStateBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))

  await page.getByRole('button', { name: 'Run Execute after review' }).click()
  await expect(page.getByText('Scan triggered for lib-audio.')).toHaveCount(0)
  await page.getByRole('button', { name: 'Trigger Scan' }).click()
  await expect(page.getByText('Scan triggered for lib-audio.')).toBeVisible()
  expect(scanBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))

  await expect(page.getByRole('button', { name: 'Clean Missing Items' })).toBeDisabled()
  await page.getByLabel('I understand this removes ABS missing item records').check()
  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Clean missing ABS item records')
    await dialog.accept()
  })
  await page.getByRole('button', { name: 'Clean Missing Items' }).click()
  await expect(page.getByText('Cleanup completed for lib-audio.')).toBeVisible()
  expect(cleanBody?.config).toEqual(expect.objectContaining({ library_id: 'lib-audio' }))

  await page.getByRole('button', { name: 'Review Inspect backend results' }).click()
  await expect(page.getByRole('heading', { name: 'ABS Operation Results' })).toBeVisible()
  await expect(page.locator('.review-layout .result-grid strong').filter({ hasText: 'lib-audio' })).toHaveCount(2)
})

test('keeps ABS later stages locked when setup requests fail', async ({ page }) => {
  await page.route('**/api/abs/libraries', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'abs token is required' }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Audiobookshelf/ }).click()
  await page.getByLabel('ABS server URL').fill('http://localhost:13378')
  await page.getByRole('button', { name: 'Load Libraries' }).click()

  await expect(page.locator('.inline-alert').filter({ hasText: 'abs token is required' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
  await page.getByRole('button', { name: 'Review Inspect backend results' }).click()
  await expect(page.getByRole('heading', { name: 'ABS Results Need Attention' })).toBeVisible()
  await expect(page.locator('.review-layout .error-list').getByText('ABS libraries: abs token is required')).toBeVisible()
  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await expect(page.locator('.inline-alert').filter({ hasText: 'ABS setup must load libraries' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
})

test('contracts organize preview and run UI state with mocked backend responses', async ({ page }) => {
  let previewBody: Record<string, any> | undefined
  let runBody: Record<string, any> | undefined

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
  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()

  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
  await page.getByRole('button', { name: 'Create Dry-run Preview' }).click()

  await expect(page.getByRole('heading', { name: 'Organize preview ready' })).toBeVisible()
  await expect(page.getByText('/library/source/missing')).toBeVisible()
  await expect(page.getByText('/library/output/Author/Book/audio.mp3')).toBeVisible()
  expect(previewBody?.config).toEqual(
    expect.objectContaining({
      base_dir: '/library/source',
      output_dir: '/library/output',
      dry_run: true,
      layout: 'author-series-title',
    }),
  )

  await page.getByRole('button', { name: 'Review Preview & Continue' }).click()
  await expect(page.getByRole('heading', { name: 'Execute the reviewed plan' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Organize' })).toBeEnabled()

  page.once('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Run Organize will change files')
    await dialog.accept()
  })
  await page.getByRole('button', { name: 'Run Organize' }).click()

  await expect(page.getByRole('heading', { name: 'Organize Run Complete' })).toBeVisible()
  await expect(
    page.locator('.review-layout .result-grid strong').filter({ hasText: '/library/output/.abook-org.log' }),
  ).toBeVisible()
  expect(runBody?.config).toEqual(expect.objectContaining({ dry_run: false }))
})

test('keeps organize run locked when preview fails', async ({ page }) => {
  await page.route('**/api/organize/preview', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'preview exploded' }),
    })
  })

  await loadApp(page)
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await page.getByRole('textbox', { name: 'Output folder' }).fill('/library/output')
  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await page.getByRole('button', { name: 'Create Dry-run Preview' }).click()

  await expect(page.locator('.inline-alert').filter({ hasText: 'preview exploded' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
  await page.getByRole('button', { name: 'Review Inspect backend results' }).click()
  await expect(page.getByRole('heading', { name: 'Organize Results Need Attention' })).toBeVisible()
  await expect(page.locator('.review-layout .error-list').getByText('Organize preview: preview exploded')).toBeVisible()
})

test('contracts rename preview UI state with mocked backend responses', async ({ page }) => {
  let previewBody: Record<string, any> | undefined
  let renameRunRequested = false

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
    renameRunRequested = true
    await route.fulfill({ status: 404, contentType: 'application/json', body: JSON.stringify({ error: 'not found' }) })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Rename/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await page.getByRole('textbox', { name: 'Rename template' }).fill('{author} - {title}')
  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()

  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
  await page.getByRole('button', { name: 'Create Rename Preview' }).click()

  await expect(page.getByRole('heading', { name: 'Rename preview ready' })).toBeVisible()
  await expect(page.getByText('/library/source/book/Rename Author - Rename Book.mp3')).toBeVisible()
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

  await page.getByRole('button', { name: 'Review Candidates & Continue' }).click()
  await expect(page.getByRole('heading', { name: 'Rename Execution Deferred' })).toBeVisible()
  await expect(page.getByText(/Rename execution is deferred/)).toBeVisible()
  await expect(page.getByRole('button', { name: 'Rename Execution Deferred' })).toBeDisabled()
  expect(renameRunRequested).toBe(false)
})

test('keeps rename run unavailable when preview fails', async ({ page }) => {
  await page.route('**/api/rename/preview', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'rename preview exploded' }),
    })
  })

  await loadApp(page)
  await page.getByRole('button', { name: /Rename/ }).click()
  await page.getByRole('textbox', { name: 'Source folder' }).fill('/library/source')
  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await page.getByRole('button', { name: 'Create Rename Preview' }).click()

  await expect(page.locator('.inline-alert').filter({ hasText: 'rename preview exploded' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
  await page.getByRole('button', { name: 'Review Inspect backend results' }).click()
  await expect(page.getByRole('heading', { name: 'Rename Results Need Attention' })).toBeVisible()
  await expect(
    page.locator('.review-layout .error-list').getByText('Rename preview: rename preview exploded'),
  ).toBeVisible()
})

test('separates workflow modes and gates run behind preview review', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('button', { name: /Rename/ }).click()
  await expect(page.getByRole('heading', { name: 'Configure and scan setup' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Rename Template' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Organization Rules' })).toHaveCount(0)

  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await expect(page.getByRole('heading', { name: 'Review a dry-run preview' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Create a rename preview' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Create Rename Preview' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Execute after review' })).toBeDisabled()
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
