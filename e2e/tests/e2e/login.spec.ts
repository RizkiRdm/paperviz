import { test, expect } from '@playwright/test'

test.describe('Login Page', () => {
  test('loads with login form', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('h1')).toContainText('Welcome back')
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toContainText('Sign in')
  })

  // PaperViz holds no Google account, so the OAuth affordance must stay gone.
  test('has no Google OAuth affordance', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/api/auth/google/login"]')).toHaveCount(0)
    await expect(page.locator('text=Continue with Google')).toHaveCount(0)
  })

  test('has link to signup', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/signup"]')).toBeVisible()
    await expect(page.locator('a[href="/signup"]')).toContainText('Sign up')
  })

  test('shows validation on empty submit', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.locator('button[type="submit"]').click()
    await expect(page.locator('text=Please fill in all fields')).toBeVisible()
  })

  test('shows error on invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.locator('#email').fill('nonexistent@test.com')
    await page.locator('#password').fill('wrongpassword')
    await page.locator('button[type="submit"]').click()
    await expect(page.locator('text=Invalid email or password')).toBeVisible({ timeout: 10000 })
  })

  test('PV logo links back to home', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.locator('a[href="/"]').first().click()
    await page.waitForLoadState('networkidle')
    expect(page.url()).toContain('/')
  })
})
