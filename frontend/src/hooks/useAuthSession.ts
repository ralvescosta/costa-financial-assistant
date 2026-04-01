/**
 * useAuthSession — login mutation hook for US1.
 *
 * Wraps the login action from AuthContext and exposes isAuthenticated,
 * isLoading, error, and lockoutUntil for LoginPage consumption.
 */

import { useAuthContext } from '@/hooks/useAuthContext'

export interface UseAuthSessionResult {
  isAuthenticated: boolean
  isLoading: boolean
  error?: string
  lockoutUntil?: number
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

export function useAuthSession(): UseAuthSessionResult {
  const { isAuthenticated, isLoading, error, lockoutUntil, login, logout } =
    useAuthContext()

  return { isAuthenticated, isLoading, error, lockoutUntil, login, logout }
}
