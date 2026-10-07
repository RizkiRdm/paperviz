import { test, expect } from '@playwright/test'

test.describe('Landing / Upload Page', () => {
  test('loads with correct title and heading', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveTitle(/PaperViz/)
    await expect(page.locator('h1')).toContainText('evidence you can inspect')
  })

  test('states the free + bring-your-own-key positioning', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('body')).toContainText('Free to use')
    await expect(page.locator('body')).toContainText('Bring your own AI API key')
  })

  test('shows no freemium or verified-research claims', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    const body = (await page.locator('body').innerText()).toLowerCase()
    for (const banned of ['upgrade', 'pricing', 'pro plan', 'research plan', 'subscription', 'billing']) {
      expect(body, `landing page must not mention "${banned}"`).not.toContain(banned)
    }
  })

  test('leads with the evidence workflow, not agent access', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    for (const step of ['Understand', 'Inspect', 'Trace', 'Reuse']) {
      await expect(page.locator('body')).toContainText(step)
    }
    // Primary CTA is the paper workflow; the MCP link is secondary.
    await expect(page.locator('a[href="#ingest"]')).toBeVisible()
    await expect(page.locator('a[href="/login"]')).toBeVisible()
    await expect(page.locator('a[href="/agents"]')).toBeVisible()
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
