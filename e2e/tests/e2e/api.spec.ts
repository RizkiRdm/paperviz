import { test, expect } from '@playwright/test'

// The Go service is an API only — the SPA has its own origin, so API specs
// must not resolve against the UI baseURL.
test.use({ baseURL: process.env.API_URL || 'http://localhost:8080' })

test.describe('API Endpoints', () => {
  test('GET /api/auth/me returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('/api/auth/me')
    expect(response.status()).toBe(401)
  })

  test('GET /api/documents returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('/api/documents')
    expect(response.status()).toBe(401)
  })

  test('POST /api/documents returns 400 when not authenticated', async ({ request }) => {
    const response = await request.post('/api/documents', { data: {} })
    expect([400, 401]).toContain(response.status())
  })

  test('GET /api/documents/stats returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('/api/documents/stats')
    expect(response.status()).toBe(401)
  })

  test('GET /api/collections returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('/api/collections')
    expect(response.status()).toBe(401)
  })

  // Clerk owns signup and login. These endpoints must stay absent: a local
  // credential path is one Clerk cannot revoke, which is the exact failure the
  // Clerk migration closed.
  test('POST /api/auth/signup does not exist', async ({ request }) => {
    const response = await request.post('/api/auth/signup', {
      data: { email: `test-${Date.now()}@example.com`, password: 'TestPassword123!' }
    })
    expect(response.status()).toBe(404)
  })

  test('POST /api/auth/login does not exist', async ({ request }) => {
    const response = await request.post('/api/auth/login', {
      data: { email: 'nonexistent@example.com', password: 'wrongpassword' }
    })
    expect(response.status()).toBe(404)
  })

  test('GET /healthz returns 200', async ({ request }) => {
    const response = await request.get('/healthz')
    expect(response.status()).toBe(200)
    const body = await response.json()
    expect(body.status).toBe('ok')
  })
})
