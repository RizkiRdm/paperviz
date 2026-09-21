import { test, expect } from '@playwright/test'

test.describe('Healthz Endpoint', () => {
  test('GET /healthz returns 200 with status ok', async ({ request }) => {
    const res = await request.get('/healthz')
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.status).toBe('ok')
  })

  test('GET /healthz is not rate limited', async ({ request }) => {
    for (let i = 0; i < 5; i++) {
      const res = await request.get('/healthz')
      expect(res.status()).toBe(200)
    }
  })
})
