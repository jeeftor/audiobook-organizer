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

  await expect(page.getByText(/Saved ABS credentials are redacted/)).toBeVisible()
})

test('separates workflow modes and gates run behind preview review', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('button', { name: /Rename/ }).click()
  await expect(page.getByRole('heading', { name: 'Configure and scan setup' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Rename Template' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Organization Rules' })).toHaveCount(0)

  await page.getByRole('button', { name: 'Preview Review dry-run output' }).click()
  await expect(page.getByRole('heading', { name: 'Review a dry-run preview' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Dry-run preview first' })).toBeVisible()

  await page.getByRole('button', { name: /Mark Preview Reviewed/ }).click()
  await expect(page.getByRole('heading', { name: 'Execute the reviewed plan' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Run Rename' })).toBeEnabled()
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
