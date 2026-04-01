/**
 * useAuthContext — central authentication state management hook.
 *
 * Provides login, logout, token refresh, and active project switching.
 * Auth tokens are managed exclusively via HTTP-only cookies set by the BFF;
 * the hook only stores non-sensitive session metadata.
 */

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import type { AuthenticationContext, User, ActiveProject } from '@/types/auth'
import { apiPost, apiPostEmpty, ApiError } from '@/services/api.client'
import {
  LoginSuccessResponseSchema,
  RefreshSuccessResponseSchema,
} from '@/types/auth-response.schema'
import {
  LockoutErrorResponseSchema,
  calcLockoutRemainingSeconds,
} from '@/types/lockout.schema'
import {
  SessionMetadataSchema,
  type SessionMetadata,
} from '@/types/session.schema'

const SESSION_STORAGE_KEY = 'cfa:session'

function loadPersistedSession(): SessionMetadata | null {
  try {
    const raw = localStorage.getItem(SESSION_STORAGE_KEY)
    if (!raw) return null
    const parsed = SessionMetadataSchema.safeParse(JSON.parse(raw))
    if (!parsed.success) return null
    // Discard expired sessions
    if (parsed.data.expiryTimestamp < Math.floor(Date.now() / 1000)) {
      localStorage.removeItem(SESSION_STORAGE_KEY)
      return null
    }
    return parsed.data
  } catch {
    return null
  }
}

function persistSession(
  user: User,
  expiryTimestamp: number,
  refreshAtTimestamp: number,
  activeProjectId?: string,
): void {
  const meta: SessionMetadata = {
    userId: user.id,
    username: user.username,
    expiryTimestamp,
    refreshAtTimestamp,
    activeProjectId,
  }
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(meta))
}

function clearPersistedSession(): void {
  localStorage.removeItem(SESSION_STORAGE_KEY)
}

// ─── Context ────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthenticationContext | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const persisted = loadPersistedSession()

  const [isAuthenticated, setIsAuthenticated] = useState(persisted !== null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | undefined>(undefined)
  const [user, setUser] = useState<User | undefined>(
    persisted ? { id: persisted.userId, username: persisted.username } : undefined,
  )
  const [activeProject, setActiveProjectState] = useState<ActiveProject | undefined>(undefined)
  const [expiresIn, setExpiresIn] = useState<number | undefined>(undefined)
  const [expiryTimestamp, setExpiryTimestamp] = useState<number | undefined>(
    persisted?.expiryTimestamp,
  )
  const [refreshAtTimestamp, setRefreshAtTimestamp] = useState<number | undefined>(
    persisted?.refreshAtTimestamp,
  )
  const [csrfToken, setCsrfToken] = useState<string | undefined>(undefined)
  const [lockoutUntil, setLockoutUntil] = useState<number | undefined>(undefined)

  // Restore active project from persisted session on mount
  useEffect(() => {
    if (persisted?.activeProjectId) {
      setActiveProjectState({ id: persisted.activeProjectId, name: '', role: 'read_only' })
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const login = useCallback(async (username: string, password: string) => {
    setIsLoading(true)
    setError(undefined)
    setLockoutUntil(undefined)
    try {
      const raw = await apiPost('/auth/login', { username, password })
      const parsed = LoginSuccessResponseSchema.safeParse(raw)
      if (!parsed.success) {
        throw new Error('Unexpected server response format')
      }
      const { data } = parsed.data
      const nowSec = Math.floor(Date.now() / 1000)
      const expiry = nowSec + data.expiresIn
      const refreshAt = nowSec + data.refreshAt

      setUser(data.user)
      setExpiresIn(data.expiresIn)
      setExpiryTimestamp(expiry)
      setRefreshAtTimestamp(refreshAt)
      setCsrfToken(data.csrfToken)
      if (data.activeProject) setActiveProjectState(data.activeProject)
      setIsAuthenticated(true)
      persistSession(data.user, expiry, refreshAt, data.activeProject?.id)
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 429) {
          const lockout = LockoutErrorResponseSchema.safeParse(err.body)
          if (lockout.success) {
            const remaining = calcLockoutRemainingSeconds(
              lockout.data.error.lockoutUntil,
            )
            setLockoutUntil(Math.floor(Date.now() / 1000) + remaining)
            setError(lockout.data.error.message)
            return
          }
        }
        setError(err.message)
      } else {
        setError('Login failed. Please try again.')
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const logout = useCallback(() => {
    clearPersistedSession()
    setIsAuthenticated(false)
    setUser(undefined)
    setActiveProjectState(undefined)
    setExpiresIn(undefined)
    setExpiryTimestamp(undefined)
    setRefreshAtTimestamp(undefined)
    setCsrfToken(undefined)
    setLockoutUntil(undefined)
    setError(undefined)
  }, [])

  const refreshAccessToken = useCallback(async () => {
    try {
      const raw = await apiPostEmpty('/auth/refresh')
      const parsed = RefreshSuccessResponseSchema.safeParse(raw)
      if (!parsed.success) return
      const { data } = parsed.data
      const nowSec = Math.floor(Date.now() / 1000)
      const expiry = nowSec + data.expiresIn
      const refreshAt = nowSec + data.refreshAt

      setExpiresIn(data.expiresIn)
      setExpiryTimestamp(expiry)
      setRefreshAtTimestamp(refreshAt)
      setCsrfToken(data.csrfToken)
      if (user) persistSession(user, expiry, refreshAt, activeProject?.id)
    } catch {
      // Refresh failure triggers re-login via interceptor
      logout()
    }
  }, [user, activeProject, logout])

  const setActiveProject = useCallback((projectId: string) => {
    setActiveProjectState((prev) =>
      prev ? { ...prev, id: projectId } : { id: projectId, name: '', role: 'read_only' },
    )
  }, [])

  const value = useMemo<AuthenticationContext>(
    () => ({
      isAuthenticated,
      isLoading,
      error,
      expiresIn,
      expiryTimestamp,
      refreshAtTimestamp,
      csrfToken,
      lockoutUntil,
      user,
      activeProject,
      login,
      logout,
      refreshAccessToken,
      setActiveProject,
    }),
    [
      isAuthenticated,
      isLoading,
      error,
      expiresIn,
      expiryTimestamp,
      refreshAtTimestamp,
      csrfToken,
      lockoutUntil,
      user,
      activeProject,
      login,
      logout,
      refreshAccessToken,
      setActiveProject,
    ],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

/**
 * Hook to consume the AuthContext.
 * Must be used within an <AuthProvider> subtree.
 */
export function useAuthContext(): AuthenticationContext {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuthContext must be used within an AuthProvider')
  }
  return ctx
}
