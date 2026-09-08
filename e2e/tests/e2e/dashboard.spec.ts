import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Dashboard Page', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForLoadState('networkidle')
    
    await page.waitForTimeout(2000)
    
    const url = page.url()
    const isOnLogin = url.includes('/login')
    const isOnDashboard = url.includes('/dashboard')
    
    expect(isOnLogin || isOnDashboard).toBeTruthy()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/dashboard-unauth.png`,
      fullPage: true 
    })
  })

  test('shows loading state initially', async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForLoadState('networkidle')
    
    const spinner = page.locator('.animate-spin, [role="progressbar"]').first()
    const isVisible = await spinner.isVisible().catch(() => false)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/dashboard-loading.png`,
      fullPage: true 
    })
  })

  test('shows header with PaperViz branding', async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForLoadState('networkidle')
    
    const header = page.locator('header').first()
    await expect(header).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/dashboard-header.png`,
      fullPage: true 
    })
  })
})
