import { test, expect } from '@playwright/test'

test.describe('Account Page', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/account')
    await page.waitForLoadState('networkidle')
    expect(page.url()).toContain('/login')
  })

  test('shows account when authenticated', async ({ page, request }) => {
    const email = `e2e-account-${Date.now()}@example.com`
    const password = 'TestPassword123!'
    const signupRes = await request.post('/api/auth/signup', { data: { email, password } })
    expect([200, 201]).toContain(signupRes.status())
    const _cookies = signupRes.headersArray().filter(h => h.name === 'set-cookie')
    // reuse session via storage state: login then go to account
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.locator('#email').fill(email)
    await page.locator('#password').fill(password)
    await page.locator('button[type="submit"]').click()
    await page.waitForURL('**/account', { timeout: 10000 })
    await expect(page.locator('h1')).toContainText('Account')
    await expect(page.locator(`text=${email}`)).toBeVisible()
  })
})
