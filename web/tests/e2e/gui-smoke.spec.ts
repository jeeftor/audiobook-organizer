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

test('renders the dashboard and connects to the local server', async ({ page }) => {
  await loadApp(page)

  await expect(page).toHaveTitle('Audiobook Organizer')
  await expect(page.getByRole('heading', { name: 'Audiobook Organizer' })).toBeVisible()
  await expect(page.getByText('localhost connected')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Source & Output' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Audiobookshelf Connection' })).toBeVisible()
  await expect(page.getByRole('cell', { name: 'Project Hail Mary' })).toBeVisible()
})

test('updates visible state from table and scan interactions', async ({ page }) => {
  await loadApp(page)

  await page.getByRole('cell', { name: 'Dune' }).click()
  await expect(page.getByRole('heading', { name: 'Selected: Dune' })).toBeVisible()

  await page.getByRole('button', { name: 'Scan Library' }).click()
  await expect(page.getByText('Scan requested')).toBeVisible()
  await expect(page.getByText('/Volumes/Media/Audiobooks/Unsorted')).toBeVisible()
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
