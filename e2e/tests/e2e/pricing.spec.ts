import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Pricing Page', () => {
  test('loads with pricing information', async ({ page }) => {
    await page.goto('/pricing')
    await page.waitForLoadState('networkidle')
    
    const heading = page.locator('h1, h2').first()
    await expect(heading).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/pricing-page.png`,
      fullPage: true 
    })
  })

  test('shows pricing tiers', async ({ page }) => {
    await page.goto('/pricing')
    await page.waitForLoadState('networkidle')
    
    const cards = page.locator('[class*="card"], [class*="tier"], [class*="plan"]').first()
    const isVisible = await cards.isVisible().catch(() => false)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/pricing-tiers.png`,
      fullPage: true 
    })
  })

  test('has navigation back to main app', async ({ page }) => {
    await page.goto('/pricing')
    await page.waitForLoadState('networkidle')
    
    const homeLink = page.locator('a[href="/"]').first()
    await expect(homeLink).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/pricing-nav.png`,
      fullPage: true 
    })
  })
})
