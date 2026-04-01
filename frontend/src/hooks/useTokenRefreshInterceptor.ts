/**
 * useTokenRefreshInterceptor — intercepts 401 API errors, attempts a token
 * refresh, and retries the original request once. A second 401 after the
 * retry triggers logout to prevent an infinite refresh loop.
 *
 * Usage: mount this hook once inside AuthProvider-wrapped components.
 * It patches `window.fetch` for the lifetime of the component.
 */

import { useEffect, useRef } from 'react'
import { useAuthContext } from '@/hooks/useAuthContext'
import { ApiError } from '@/services/api.client'

export function useTokenRefreshInterceptor(): void {
  const { refreshAccessToken, logout } = useAuthContext()
  const isRefreshing = useRef(false)
  const originalFetch = useRef(window.fetch)

  useEffect(() => {
    const nativeFetch = originalFetch.current

    window.fetch = async (input, init) => {
      const res = await nativeFetch(input, init)

      if (res.status !== 401) return res

      // Prevent retry loop: if already refreshing, log out
      if (isRefreshing.current) {
        logout()
        return res
      }

      isRefreshing.current = true
      try {
        await refreshAccessToken()
      } catch {
        logout()
        isRefreshing.current = false
        return res
      }
      isRefreshing.current = false

      // Retry the original request once with the new session cookie
      return nativeFetch(input, init)
    }

    return () => {
      window.fetch = nativeFetch
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshAccessToken, logout])
}

// Re-export ApiError so consumers can reference it without importing from api.client
export { ApiError }
