import { test, expect } from '@playwright/test'

test.describe('Result Page', () => {
  test('shows error for non-existent document', async ({ page }) => {
    await page.goto('/nonexistent-doc-xyz')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('text=/error|not found|invalid/i').first()).toBeVisible({ timeout: 10000 })
  })
})
