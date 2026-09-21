import { test, expect } from '@playwright/test'

test.describe('Agents Page', () => {
  test('loads with heading and client tabs', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('h1')).toContainText('Add PaperViz to your agent')
    const tablist = page.locator('[role="tablist"]')
    await expect(tablist).toBeVisible()
    await expect(tablist.locator('[role="tab"]')).toHaveCount(4)
    await expect(page.locator('[role="tab"]', { hasText: 'Claude Code' })).toBeVisible()
  })

  test('shows sign-in prompt when not authenticated', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('text=Sign in to get your API key')).toBeVisible()
    await expect(page.locator('a[href="/login"]')).toBeVisible()
  })

  test('switching tabs updates config panel', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    await page.locator('[role="tab"]', { hasText: 'Cursor' }).click()
    await expect(page.locator('[role="tabpanel"]')).toBeVisible()
    await page.locator('[role="tab"]', { hasText: 'Claude Desktop' }).click()
    await expect(page.locator('[role="tabpanel"]')).toBeVisible()
  })
})
