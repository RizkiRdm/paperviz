import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Landing / Upload Page', () => {
  test('loads with correct title and heading', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await expect(page).toHaveTitle(/PaperViz/)
    await expect(page.locator('h1')).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/landing-page.png`,
      fullPage: true 
    })
  })

  test('shows upload dropzone', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const dropzone = page.locator('[data-testid="upload-dropzone"], [role="button"]').first()
    await expect(dropzone).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/landing-dropzone.png`,
      fullPage: true 
    })
  })

  test('shows paste text option', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const pasteButton = page.locator('text=/paste|text/i').first()
    await expect(pasteButton).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/landing-paste-option.png`,
      fullPage: true 
    })
  })

  test('navigation links are visible', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const nav = page.locator('header, nav').first()
    await expect(nav).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/landing-nav.png`,
      fullPage: true 
    })
  })

  test('clicking login navigates to login page', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const loginLink = page.locator('a[href="/login"], button:has-text("Log in"), button:has-text("Sign in")').first()
    if (await loginLink.isVisible()) {
      await loginLink.click()
      await page.waitForLoadState('networkidle')
      expect(page.url()).toContain('/login')
      
      await page.screenshot({ 
        path: `${SCREENSHOT_DIR}/nav-to-login.png`,
        fullPage: true 
      })
    }
  })

  test('clicking signup navigates to signup page', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const signupLink = page.locator('a[href="/signup"], button:has-text("Sign up"), button:has-text("Register")').first()
    if (await signupLink.isVisible()) {
      await signupLink.click()
      await page.waitForLoadState('networkidle')
      expect(page.url()).toContain('/signup')
      
      await page.screenshot({ 
        path: `${SCREENSHOT_DIR}/nav-to-signup.png`,
        fullPage: true 
      })
    }
  })
})
