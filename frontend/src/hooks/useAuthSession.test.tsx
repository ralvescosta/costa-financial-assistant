/**
 * Contract test: BFF POST /api/auth/login response envelope validation.
 *
 * Uses MSW to mock the BFF and verifies that the useAuthSession hook
 * correctly validates and maps login/error/lockout responses.
 */

import { describe, it, expect, beforeAll, beforeEach, afterEach, afterAll } from 'vitest'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { renderHook, act } from '@testing-library/react'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { useAuthSession } from '@/hooks/useAuthSession'

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  )
}

const server = setupServer()

beforeAll(() => server.listen())
beforeEach(() => localStorage.clear())
afterEach(() => { server.resetHandlers(); localStorage.clear() })
afterAll(() => server.close())

describe('useAuthSession — BFF /api/auth/login contract', () => {
  it('maps a valid 200 response to authenticated state', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json({
          statusCode: 200,
          data: {
            expiresIn: 3600,
            refreshAt: 2700,
            csrfToken: 'csrf-abc',
            user: { id: 'u1', username: 'demo' },
          },
        }),
      ),
    )

    const { result } = renderHook(() => useAuthSession(), { wrapper: Wrapper })
    await act(() => result.current.login('demo', 'secret'))

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.error).toBeUndefined()
  })

  it('maps a 401 error to error state', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json(
          { statusCode: 401, error: { code: 'INVALID_CREDENTIALS', message: 'Bad credentials' } },
          { status: 401 },
        ),
      ),
    )

    const { result } = renderHook(() => useAuthSession(), { wrapper: Wrapper })
    await act(() => result.current.login('demo', 'wrong'))

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.error).toBeTruthy()
  })

  it('maps a 429 lockout response to lockout state', async () => {
    const lockoutUntil = new Date(Date.now() + 300_000).toISOString()
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json(
          {
            statusCode: 429,
            error: {
              code: 'AUTH_LOCKED',
              message: 'Too many attempts',
              lockoutUntil,
              remainingSeconds: 300,
            },
          },
          { status: 429 },
        ),
      ),
    )

    const { result } = renderHook(() => useAuthSession(), { wrapper: Wrapper })
    await act(() => result.current.login('demo', 'wrong'))

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.lockoutUntil).toBeDefined()
    expect(result.current.lockoutUntil).toBeGreaterThan(0)
  })
})
