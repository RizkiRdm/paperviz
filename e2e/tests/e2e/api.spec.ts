import { test, expect } from '@playwright/test'

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

  test('POST /api/auth/signup accepts valid payload', async ({ request }) => {
    const response = await request.post('/api/auth/signup', {
      data: { email: `test-${Date.now()}@example.com`, password: 'TestPassword123!' }
    })
    expect([200, 201, 409]).toContain(response.status())
  })

  test('POST /api/auth/login rejects invalid credentials', async ({ request }) => {
    const response = await request.post('/api/auth/login', {
      data: { email: 'nonexistent@example.com', password: 'wrongpassword' }
    })
    expect([400, 401, 403]).toContain(response.status())
  })

  test('GET /healthz returns 200', async ({ request }) => {
    const response = await request.get('/healthz')
    expect(response.status()).toBe(200)
    const body = await response.json()
    expect(body.status).toBe('ok')
  })
})
