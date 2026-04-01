/**
 * Contract test: BFF POST /api/auth/refresh endpoint.
 *
 * Uses MSW to verify that the auth API refresh call validates and maps
 * the server response correctly.
 */

import { describe, it, expect, beforeAll, afterEach, afterAll } from 'vitest'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { apiPostEmpty } from '@/services/api.client'
import { RefreshSuccessResponseSchema } from '@/types/auth-response.schema'

const server = setupServer()

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe('auth.api — POST /api/auth/refresh contract', () => {
  it('returns a valid refresh response that passes schema validation', async () => {
    server.use(
      http.post('/api/auth/refresh', () =>
        HttpResponse.json({
          statusCode: 200,
          data: {
            expiresIn: 3600,
            refreshAt: 2700,
            csrfToken: 'new-csrf-token',
          },
        }),
      ),
    )

    const raw = await apiPostEmpty('/auth/refresh')
    const parsed = RefreshSuccessResponseSchema.safeParse(raw)
    expect(parsed.success).toBe(true)
    if (parsed.success) {
      expect(parsed.data.data.expiresIn).toBe(3600)
      expect(parsed.data.data.refreshAt).toBe(2700)
      expect(parsed.data.data.csrfToken).toBe('new-csrf-token')
    }
  })

  it('throws ApiError on 401 refresh failure', async () => {
    server.use(
      http.post('/api/auth/refresh', () =>
        HttpResponse.json(
          { statusCode: 401, error: { code: 'SESSION_EXPIRED', message: 'Session expired' } },
          { status: 401 },
        ),
      ),
    )

    await expect(apiPostEmpty('/auth/refresh')).rejects.toThrow()
  })
})
