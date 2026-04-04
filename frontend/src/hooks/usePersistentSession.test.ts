/**
 * Unit tests for usePersistentSession.
 *
 * Covers:
 * - persisted session metadata is detected for the seeded owner flow
 * - missing storage keeps the hook unauthenticated
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { usePersistentSession } from '@/hooks/usePersistentSession'

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function Wrapper({ children }: { children: ReactNode }) {
  return createElement(
    QueryClientProvider,
    { client: queryClient },
    createElement(AuthProvider, null, children),
  )
}

describe('usePersistentSession', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('detects the persisted seeded-owner session metadata', () => {
    localStorage.setItem(
      'cfa:session',
      JSON.stringify({
        userId: '00000000-0000-0000-0000-000000000001',
        username: 'ralvescosta',
        expiryTimestamp: Math.floor(Date.now() / 1000) + 3600,
        refreshAtTimestamp: Math.floor(Date.now() / 1000) + 1800,
        activeProjectId: '00000000-0000-0000-0000-000000000010',
      }),
    )

    const { result } = renderHook(() => usePersistentSession(), { wrapper: Wrapper })

    expect(result.current.hasPersistedSession).toBe(true)
    expect(result.current.isAuthenticated).toBe(true)
  })

  it('returns false when no session metadata exists', () => {
    const { result } = renderHook(() => usePersistentSession(), { wrapper: Wrapper })

    expect(result.current.hasPersistedSession).toBe(false)
  })
})
