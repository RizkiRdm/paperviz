import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Signup Page', () => {
  test('loads with signup form', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    
    const emailInput = page.locator('input[type="email"], input[name="email"], input[placeholder*="email" i]').first()
    const passwordInput = page.locator('input[type="password"], input[name="password"]').first()
    const submitButton = page.locator('button[type="submit"], button:has-text("Sign up"), button:has-text("Create")').first()
    
    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
    await expect(submitButton).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/signup-page.png`,
      fullPage: true 
    })
  })

  test('shows validation on empty submit', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    
    const submitButton = page.locator('button[type="submit"], button:has-text("Sign up"), button:has-text("Create")').first()
    await submitButton.click()
    
    await page.waitForTimeout(500)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/signup-validation.png`,
      fullPage: true 
    })
  })

  test('shows error on duplicate email', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    
    const emailInput = page.locator('input[type="email"], input[name="email"], input[placeholder*="email" i]').first()
    const passwordInput = page.locator('input[type="password"], input[name="password"]').first()
    const submitButton = page.locator('button[type="submit"], button:has-text("Sign up"), button:has-text("Create")').first()
    
    await emailInput.fill('existing@example.com')
    await passwordInput.fill('Password123!')
    await submitButton.click()
    
    await page.waitForTimeout(2000)
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/signup-duplicate-email.png`,
      fullPage: true 
    })
  })

  test('has link to login', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    
    const loginLink = page.locator('a[href="/login"]').first()
    await expect(loginLink).toBeVisible()
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/signup-to-login-link.png`,
      fullPage: true 
    })
  })
})
