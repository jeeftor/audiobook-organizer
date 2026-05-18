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
  await expect(page.getByText('localhost connected')).toBeVisible()
  await expect(page.getByRole('button', { name: /Organize/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /Rename/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /Audiobookshelf/ })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Configure and scan setup' })).toBeVisible()
  await expect(page.getByRole('button', { name: /Run/ })).toBeDisabled()
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
  await expect(page.getByText('Preview reviewed')).toBeVisible()
})

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
