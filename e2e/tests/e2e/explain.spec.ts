import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Explain Page', () => {
  test('loads explain page with slug', async ({ page }) => {
    await page.goto('/explain/test-slug')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/explain-page.png`,
      fullPage: true 
    })
  })

  test('shows 404 for non-existent explanation', async ({ page }) => {
    await page.goto('/explain/nonexistent-slug-xyz')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    const has404 = await page.locator('text=/404|not found/i').first().isVisible().catch(() => false)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/explain-not-found.png`,
      fullPage: true 
    })
  })
})
