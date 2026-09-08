import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Compare Page', () => {
  test('loads compare page', async ({ page }) => {
    await page.goto('/compare')
    await page.waitForLoadState('networkidle')
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/compare-page.png`,
      fullPage: true 
    })
  })

  test('loads with query params', async ({ page }) => {
    await page.goto('/compare?ids=doc1,doc2')
    await page.waitForLoadState('networkidle')
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/compare-with-ids.png`,
      fullPage: true 
    })
  })
})
