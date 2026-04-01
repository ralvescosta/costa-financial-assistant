/**
 * TypeScript interfaces for the authentication and session domain.
 *
 * All token transport is handled by HTTP-only cookies set by the BFF.
 * The frontend never reads or writes auth tokens directly.
 */

export interface User {
  id: string
  username: string
  email?: string
  fullName?: string
  avatar?: string
}

export interface ActiveProject {
  id: string
  name: string
  role: 'read_only' | 'update' | 'write'
}

export interface AuthenticationContext {
  // Authentication status
  isAuthenticated: boolean
  isLoading: boolean
  error?: string

  // Session metadata (tokens are HTTP-only cookies managed by BFF)
  expiresIn?: number
  expiryTimestamp?: number
  refreshAtTimestamp?: number
  csrfToken?: string
  lockoutUntil?: number

  // User information
  user?: User

  // Active project context (multi-tenancy)
  activeProject?: ActiveProject

  // Actions
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  refreshAccessToken: () => Promise<void>
  setActiveProject: (projectId: string) => void
}

export interface Session {
  expiresIn: number
  expiryTimestamp: number
  refreshAtTimestamp: number
  csrfToken: string
  user: User
  activeProject?: ActiveProject
}
