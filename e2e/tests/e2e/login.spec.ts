import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Login Page', () => {
  test('loads with login form', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    
    const emailInput = page.locator('input[type="email"], input[name="email"], input[placeholder*="email" i]').first()
    const passwordInput = page.locator('input[type="password"], input[name="password"]').first()
    const submitButton = page.locator('button[type="submit"], button:has-text("Log in"), button:has-text("Sign in")').first()
    
    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
    await expect(submitButton).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/login-page.png`,
      fullPage: true 
    })
  })

  test('shows validation on empty submit', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    
    const submitButton = page.locator('button[type="submit"], button:has-text("Log in"), button:has-text("Sign in")').first()
    await submitButton.click()
    
    await page.waitForTimeout(500)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/login-validation.png`,
      fullPage: true 
    })
  })

  test('shows error on invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    
    const emailInput = page.locator('input[type="email"], input[name="email"], input[placeholder*="email" i]').first()
    const passwordInput = page.locator('input[type="password"], input[name="password"]').first()
    const submitButton = page.locator('button[type="submit"], button:has-text("Log in"), button:has-text("Sign in")').first()
    
    await emailInput.fill('test@example.com')
    await passwordInput.fill('wrongpassword')
    await submitButton.click()
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/login-invalid-creds.png`,
      fullPage: true 
    })
  })

  test('has link to signup', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    
    const signupLink = page.locator('a[href="/signup"]').first()
    await expect(signupLink).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/login-to-signup-link.png`,
      fullPage: true 
    })
  })
})
