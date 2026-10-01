import { test, expect } from '@playwright/test'

test.describe('Agents Page', () => {
  test('loads with heading and client tabs', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('h1')).toContainText('Add PaperViz to your agent')
    const tablist = page.locator('[role="tablist"]')
    await expect(tablist).toBeVisible()
    await expect(tablist.locator('[role="tab"]')).toHaveCount(4)
    await expect(page.locator('[role="tab"]', { hasText: 'Claude Code' })).toBeVisible()
  })

  test('publishes a stdio config, not a nonexistent hosted endpoint', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    // The repo ships a stdio MCP server (cmd/mcp). Publishing a config that
    // points at /api/mcp hands users an endpoint the router does not serve.
    await expect(page.locator('body')).not.toContainText('/api/mcp')
    await expect(page.locator('pre')).toContainText('paperviz-mcp')
    await expect(page.locator('pre')).toContainText('GEMINI_API_KEY')
  })

  test('tells the user ChatGPT has no hosted endpoint', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    await page.locator('[role="tab"]', { hasText: 'ChatGPT' }).click()
    await expect(page.locator('[role="tabpanel"]')).toContainText('does not expose a hosted MCP endpoint')
  })

  test('switching tabs updates config panel', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')
    await page.locator('[role="tab"]', { hasText: 'Cursor' }).click()
    await expect(page.locator('[role="tabpanel"]')).toBeVisible()
    await page.locator('[role="tab"]', { hasText: 'Claude Desktop' }).click()
    await expect(page.locator('[role="tabpanel"]')).toBeVisible()
  })
})
