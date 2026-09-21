import { test, expect } from '@playwright/test'

test.describe('Signup Page', () => {
  test('loads with signup form', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toContainText('Sign up')
  })

  test('shows Google OAuth button', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/api/auth/google/login"]')).toBeVisible()
    await expect(page.locator('a[href="/api/auth/google/login"]')).toContainText('Continue with Google')
  })

  test('has link to login', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('a[href="/login"]')).toBeVisible()
    await expect(page.locator('a[href="/login"]')).toContainText('Sign in')
  })

  test('shows validation on empty submit', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    await page.locator('button[type="submit"]').click()
    await expect(page.locator('text=Please fill in all fields')).toBeVisible()
  })

  test('PV logo links back to home', async ({ page }) => {
    await page.goto('/signup')
    await page.waitForLoadState('networkidle')
    await page.locator('a[href="/"]').first().click()
    await page.waitForLoadState('networkidle')
    expect(page.url()).toContain('/')
  })
})
