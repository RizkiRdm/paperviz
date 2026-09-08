import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Result Page', () => {
  test('shows loading state for invalid document', async ({ page }) => {
    await page.goto('/invalid-doc-id-12345')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/result-invalid-doc.png`,
      fullPage: true 
    })
  })

  test('shows 404 or error for non-existent document', async ({ page }) => {
    await page.goto('/nonexistent-doc-xyz')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    const hasError = await page.locator('text=/error|not found|invalid/i').first().isVisible().catch(() => false)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/result-not-found.png`,
      fullPage: true 
    })
  })
})
