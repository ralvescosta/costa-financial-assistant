/**
 * useAuthRefresh — schedules a silent token refresh at 75% of session lifetime.
 *
 * Reads refreshAtTimestamp from auth context and sets a timer. When the timer
 * fires it calls refreshAccessToken(), which rotates the HTTP-only session
 * cookies via the BFF. On unmount or logout the timer is cancelled.
 */

import { useEffect, useRef } from 'react'
import { useAuthContext } from '@/hooks/useAuthContext'

export function useAuthRefresh(): void {
  const { isAuthenticated, refreshAtTimestamp, refreshAccessToken } = useAuthContext()
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!isAuthenticated || refreshAtTimestamp == null) return

    const nowSec = Math.floor(Date.now() / 1000)
    const delayMs = Math.max(0, (refreshAtTimestamp - nowSec) * 1000)

    timerRef.current = setTimeout(() => {
      void refreshAccessToken()
    }, delayMs)

    return () => {
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current)
        timerRef.current = null
      }
    }
  }, [isAuthenticated, refreshAtTimestamp, refreshAccessToken])
}
