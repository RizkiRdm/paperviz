import { test, expect } from '@playwright/test'

const SCREENSHOT_DIR = '../assets'

test.describe('API Endpoints', () => {
  test('GET /api/auth/me returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('http://localhost:6767/api/auth/me')
    expect(response.status()).toBe(401)
  })

  test('GET /api/documents returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('http://localhost:6767/api/documents')
    expect(response.status()).toBe(401)
  })

  test('POST /api/documents returns 401 or 400 when not authenticated', async ({ request }) => {
    const response = await request.post('http://localhost:6767/api/documents', {
      data: {}
    })
    expect([400, 401]).toContain(response.status())
  })

  test('GET /api/documents/stats returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('http://localhost:6767/api/documents/stats')
    expect(response.status()).toBe(401)
  })

  test('GET /api/collections returns 401 when not authenticated', async ({ request }) => {
    const response = await request.get('http://localhost:6767/api/collections')
    expect(response.status()).toBe(401)
  })

  test('POST /api/auth/signup accepts valid payload', async ({ request }) => {
    const response = await request.post('http://localhost:6767/api/auth/signup', {
      data: {
        email: `test-${Date.now()}@example.com`,
        password: 'TestPassword123!'
      }
    })
    expect([200, 201, 409]).toContain(response.status())
  })

  test('POST /api/auth/login rejects invalid credentials', async ({ request }) => {
    const response = await request.post('http://localhost:6767/api/auth/login', {
      data: {
        email: 'nonexistent@example.com',
        password: 'wrongpassword'
      }
    })
    expect([400, 401, 403]).toContain(response.status())
  })

  test('GET /share/fig/:token returns error for invalid token', async ({ request }) => {
    const response = await request.get('http://localhost:6767/api/share/fig/invalid-token')
    expect([400, 404]).toContain(response.status())
  })

  test('GET /share/doc/:token returns error for invalid token', async ({ request }) => {
    const response = await request.get('http://localhost:6767/api/share/doc/invalid-token')
    expect([400, 404]).toContain(response.status())
  })
})
