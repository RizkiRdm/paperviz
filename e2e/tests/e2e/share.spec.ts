import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Share Pages', () => {
  test('share figure page loads with invalid token', async ({ page }) => {
    await page.goto('/share/fig/invalid-token-12345')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/share-figure-invalid.png`,
      fullPage: true 
    })
  })

  test('share paper page loads with invalid token', async ({ page }) => {
    await page.goto('/share/doc/invalid-token-12345')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/share-paper-invalid.png`,
      fullPage: true 
    })
  })

  test('share figure page shows error for non-existent token', async ({ page }) => {
    await page.goto('/share/fig/nonexistent')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/share-figure-not-found.png`,
      fullPage: true 
    })
  })

  test('share paper page shows error for non-existent token', async ({ page }) => {
    await page.goto('/share/doc/nonexistent')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/share-paper-not-found.png`,
      fullPage: true 
    })
  })
})
