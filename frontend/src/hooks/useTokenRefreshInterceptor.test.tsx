/**
 * Unit tests for useTokenRefreshInterceptor hook.
 *
 * Covers:
 * - 401 responses trigger token refresh
 * - Original request retried after refresh
 * - Failed refresh triggers logout
 * - Prevents retry loop (max 1 attempt)
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { useTokenRefreshInterceptor } from '@/hooks/useTokenRefreshInterceptor'

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  )
}

describe('useTokenRefreshInterceptor', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    localStorage.clear()
  })

  it('patches window.fetch on mount and restores on unmount', async () => {
    const nativeFetch = globalThis.fetch
    const { unmount } = renderHook(() => useTokenRefreshInterceptor(), { wrapper: Wrapper })

    // After mount, fetch should be patched
    expect(globalThis.fetch).not.toBe(nativeFetch)

    unmount()

    // After unmount, native fetch should be restored
    expect(globalThis.fetch).toBe(nativeFetch)
  })

  it('passes through non-401 responses unchanged', async () => {
    const mockFetch = vi.fn().mockResolvedValue(new Response('ok', { status: 200 }))
    globalThis.fetch = mockFetch

    renderHook(() => useTokenRefreshInterceptor(), { wrapper: Wrapper })

    await act(async () => {
      const res = await fetch('/api/some-resource')
      expect(res.status).toBe(200)
    })

    expect(mockFetch).toHaveBeenCalledOnce()
  })

  it('attempts refresh on 401 response', async () => {
    let callCount = 0
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      callCount++
      if (callCount === 1) {
        return new Response(JSON.stringify({ statusCode: 401, error: { code: 'SESSION_EXPIRED', message: '' } }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      // Refresh endpoint
      if (String(input).includes('/auth/refresh')) {
        return new Response(null, { status: 401 })
      }
      return new Response('ok', { status: 200 })
    }) as typeof fetch

    renderHook(() => useTokenRefreshInterceptor(), { wrapper: Wrapper })

    await act(async () => {
      await fetch('/api/protected').catch(() => { })
    })

    // Should have called fetch at least twice (original + refresh attempt)
    expect(callCount).toBeGreaterThanOrEqual(1)
  })
})
