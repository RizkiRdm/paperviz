import { test, expect } from '@playwright/test'

test.describe('404 Page', () => {
  test('shows not-found page for unknown routes', async ({ page }) => {
    await page.goto('/this-page-does-not-exist-12345')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('text=/404|not found/i').first()).toBeVisible({ timeout: 10000 })
  })

  test('has link back to home', async ({ page }) => {
    await page.goto('/completely-made-up-route')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/"]').first()).toBeVisible()
  })
})
