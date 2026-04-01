/**
 * usePersistentSession — restores persisted session on app mount.
 *
 * The AuthContext already serialises session metadata to localStorage on
 * login and reads it on initialisation. This hook exposes whether the
 * persisted session was found (for loading state in the app root).
 *
 * Token transport (HTTP-only cookies) is handled by the BFF — this hook
 * only manages non-sensitive session metadata used for page-reload recovery.
 */

import { useMemo } from 'react'
import { useAuthContext } from '@/hooks/useAuthContext'

export interface UsePersistentSessionResult {
  hasPersistedSession: boolean
  isAuthenticated: boolean
}

export function usePersistentSession(): UsePersistentSessionResult {
  const { isAuthenticated } = useAuthContext()

  const hasPersistedSession = useMemo(() => {
    try {
      return localStorage.getItem('cfa:session') !== null
    } catch {
      return false
    }
  }, [])

  return { hasPersistedSession, isAuthenticated }
}
