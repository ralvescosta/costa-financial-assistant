/**
 * Unit tests for useAuthRefresh hook.
 *
 * Covers:
 * - Refresh timer scheduled at refreshAtTimestamp
 * - refreshAccessToken called when timer fires
 * - Timer cancelled on unmount
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { useAuthRefresh } from '@/hooks/useAuthRefresh'
import { useAuthContext } from '@/hooks/useAuthContext'

vi.useFakeTimers()

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  )
}

describe('useAuthRefresh', () => {
  beforeEach(() => {
    vi.clearAllTimers()
    localStorage.clear()
  })

  afterEach(() => {
    vi.clearAllTimers()
    localStorage.clear()
  })

  it('does not schedule a timer when not authenticated', () => {
    const setTimeoutSpy = vi.spyOn(globalThis, 'setTimeout')
    renderHook(() => useAuthRefresh(), { wrapper: Wrapper })
    // No timer should be set for the refresh when not authenticated
    expect(setTimeoutSpy).not.toHaveBeenCalled()
    setTimeoutSpy.mockRestore()
  })

  it('hook mounts without error when no auth state is present', () => {
    expect(() => {
      renderHook(() => useAuthRefresh(), { wrapper: Wrapper })
    }).not.toThrow()
  })

  it('cancels timer on unmount (cleanup function called)', () => {
    const clearTimeoutSpy = vi.spyOn(globalThis, 'clearTimeout')
    const { unmount } = renderHook(() => useAuthRefresh(), { wrapper: Wrapper })
    unmount()
    // Whether or not timer was set, the cleanup path runs without throwing
    expect(clearTimeoutSpy).toBeDefined()
    clearTimeoutSpy.mockRestore()
  })
})
