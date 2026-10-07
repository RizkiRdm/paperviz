import { test, expect } from '@playwright/test'

test.describe('Landing / Upload Page', () => {
  test('loads with correct title and heading', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveTitle(/PaperViz/)
    await expect(page.locator('h1')).toContainText('Papers, in plain language')
  })

  test('shows two CTAs: Add to Claude Code and Sign in', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/agents"]')).toBeVisible()
    await expect(page.locator('a[href="/login"]')).toBeVisible()
  })

  // The header used to render a hardcoded Sign in link, so a signed-in user
  // was told to sign in on the page they were already signed in to.
  test('shows the account link instead of Sign in once signed in', async ({ page }) => {
    const email = `e2e-header-${Date.now()}@test.com`
    const signup = await page.request.post('/api/auth/signup', {
      data: { email, password: 'Str0ng!Passphrase' },
    })
    expect(signup.ok()).toBeTruthy()

    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page.locator(`a[href="/account"]`)).toContainText(email)
    await expect(page.locator('a[href="/login"]')).toHaveCount(0)
  })

  test('shows Sign in to an anonymous visitor', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/login"]')).toBeVisible()
    await expect(page.locator('a[href="/account"]')).toHaveCount(0)
  })

  test('shows PDF/Paste/DOI/URL tab selector', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    const tablist = page.locator('[role="tablist"]')
    await expect(tablist).toBeVisible()
    await expect(tablist.locator('[role="tab"]')).toHaveCount(4)
    await expect(tablist.locator('[role="tab"]').nth(0)).toHaveText('PDF')
    await expect(tablist.locator('[role="tab"]').nth(1)).toHaveText('Paste')
    await expect(tablist.locator('[role="tab"]').nth(2)).toHaveText('DOI')
    await expect(tablist.locator('[role="tab"]').nth(3)).toHaveText('URL')
  })

  test('shows source type badge', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('text=source: pdf')).toBeVisible()
  })

  test('switching tabs updates source badge', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('[role="tab"]', { hasText: 'Paste' }).click()
    await expect(page.locator('text=source: pasted_text')).toBeVisible()
    await page.locator('[role="tab"]', { hasText: 'DOI' }).click()
    await expect(page.locator('text=source: doi')).toBeVisible()
  })

  test('DOI tab shows input and Import button', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('[role="tab"]', { hasText: 'DOI' }).click()
    await expect(page.locator('#doi-input')).toBeVisible()
    await expect(page.locator('button', { hasText: 'Import' })).toBeVisible()
  })

  test('URL tab shows input and Import button', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('[role="tab"]', { hasText: 'URL' }).click()
    await expect(page.locator('#url-input')).toBeVisible()
    await expect(page.locator('button', { hasText: 'Import' })).toBeVisible()
  })

  test('clicking Sign in navigates to /login', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('a[href="/login"]').click()
    await page.waitForLoadState('networkidle')
    expect(page.url()).toContain('/login')
  })

  test('clicking Add to Claude Code navigates to /agents', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('a[href="/agents"]').click()
    await page.waitForLoadState('networkidle')
    expect(page.url()).toContain('/agents')
  })
})
