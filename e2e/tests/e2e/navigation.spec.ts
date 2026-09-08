import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('Navigation Flow', () => {
  test('full navigation flow: landing -> login -> signup -> landing', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await page.screenshot({ 
      path: `${SCREENSHOT_DIR}/flow-01-landing.png`,
      fullPage: true 
    })
    
    const loginLink = page.locator('a[href="/login"], button:has-text("Log in"), button:has-text("Sign in")').first()
    if (await loginLink.isVisible()) {
      await loginLink.click()
      await page.waitForLoadState('networkidle')
      
      await page.screenshot({ 
        path: `${SCREENSHOT_DIR}/flow-02-login.png`,
        fullPage: true 
      })
      
      const signupLink = page.locator('a[href="/signup"]').first()
      if (await signupLink.isVisible()) {
        await signupLink.click()
        await page.waitForLoadState('networkidle')
        
        await page.screenshot({ 
          path: `${SCREENSHOT_DIR}/flow-03-signup.png`,
          fullPage: true 
        })
      }
    }
  })

  test('pricing page navigation', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const pricingLink = page.locator('a[href="/pricing"]').first()
    if (await pricingLink.isVisible()) {
      await pricingLink.click()
      await page.waitForLoadState('networkidle')
      
      await page.screenshot({ 
        path: `${SCREENSHOT_DIR}/flow-pricing.png`,
        fullPage: true 
      })
    }
  })

  test('compare page navigation', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const compareLink = page.locator('a[href="/compare"]').first()
    if (await compareLink.isVisible()) {
      await compareLink.click()
      await page.waitForLoadState('networkidle')
      
      await page.screenshot({ 
        path: `${SCREENSHOT_DIR}/flow-compare.png`,
        fullPage: true 
      })
    }
  })
})
