import { test, expect } from '@playwright/test'

test.describe('Account Page', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/account')
    // The redirect happens in an effect after fetch(/api/auth/me) settles, so
    // waiting on networkidle can assert the URL before the navigation occurs.
    await page.waitForURL('**/login')
    expect(page.url()).toContain('/login')
  })

  // Needs a Clerk session; see the note in landing.spec.ts.
  test.skip('shows account when authenticated', async () => {})
})
