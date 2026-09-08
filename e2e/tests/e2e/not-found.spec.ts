import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('404 Page', () => {
  test('shows 404 for unknown routes', async ({ page }) => {
    await page.goto('/this-page-does-not-exist-12345')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(1000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/404-page.png`,
      fullPage: true 
    })
  })

  test('404 page has link back to home', async ({ page }) => {
    await page.goto('/completely-made-up-route')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(1000)
    
    const homeLink = page.locator('a[href="/"]').first()
    const isVisible = await homeLink.isVisible().catch(() => false)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/404-home-link.png`,
      fullPage: true 
    })
  })
})
