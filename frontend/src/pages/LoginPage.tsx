/**
 * LoginPage — secure login screen.
 *
 * Features:
 * - Default dev credentials pre-filled from env vars (never in production)
 * - Skeleton placeholder within 300ms during submission
 * - Lockout-aware error handling with countdown timer
 * - Accessible: Tab order (username → password → button), Enter submits
 * - Navigates to /dashboard on success
 */

import { type FormEvent, useEffect, useId, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthSession } from '@/hooks/useAuthSession'
import { ErrorMessage } from '@/components/ErrorMessage'
import { SkeletonPlaceholder } from '@/components/SkeletonPlaceholder'
import { authConfig } from '@/config/auth.config'

export function LoginPage() {
  const navigate = useNavigate()
  const { login, isLoading, error, lockoutUntil, isAuthenticated } = useAuthSession()

  const [username, setUsername] = useState(authConfig.defaultUsername ?? '')
  const [password, setPassword] = useState(authConfig.defaultPassword ?? '')

  const usernameId = useId()
  const passwordId = useId()

  // Navigate to dashboard when auth state becomes true
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true })
    }
  }, [isAuthenticated, navigate])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    await login(username, password)
  }

  const isLocked = lockoutUntil != null && lockoutUntil > Math.floor(Date.now() / 1000)

  return (
    <div className="flex min-h-screen items-center justify-center bg-[color:var(--color-surface)] p-4">
      <div className="w-full max-w-sm rounded-xl bg-[color:var(--color-surface-raised)] p-8 shadow-md">
        <h1 className="mb-6 text-2xl font-bold text-[color:var(--color-text-primary)]">
          Sign in
        </h1>

        {/* Error / lockout message */}
        {error && <div className="mb-4"><ErrorMessage message={error} lockoutUntil={lockoutUntil} /></div>}

        <form onSubmit={handleSubmit} noValidate>
          {/* Username */}
          <div className="mb-4">
            <label
              htmlFor={usernameId}
              className="mb-1 block text-sm font-medium text-[color:var(--color-text-secondary)]"
            >
              Username
            </label>
            {isLoading ? (
              <SkeletonPlaceholder height="h-10" />
            ) : (
              <input
                id={usernameId}
                type="text"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={isLocked || isLoading}
                required
                className="
                  w-full rounded-md border border-[color:var(--color-border)]
                  bg-[color:var(--color-surface)] px-3 py-2 text-sm
                  text-[color:var(--color-text-primary)]
                  focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary)]
                  disabled:cursor-not-allowed disabled:opacity-50
                "
              />
            )}
          </div>

          {/* Password */}
          <div className="mb-6">
            <label
              htmlFor={passwordId}
              className="mb-1 block text-sm font-medium text-[color:var(--color-text-secondary)]"
            >
              Password
            </label>
            {isLoading ? (
              <SkeletonPlaceholder height="h-10" />
            ) : (
              <input
                id={passwordId}
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isLocked || isLoading}
                required
                className="
                  w-full rounded-md border border-[color:var(--color-border)]
                  bg-[color:var(--color-surface)] px-3 py-2 text-sm
                  text-[color:var(--color-text-primary)]
                  focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary)]
                  disabled:cursor-not-allowed disabled:opacity-50
                "
              />
            )}
          </div>

          {/* Submit */}
          <button
            type="submit"
            disabled={isLoading || isLocked}
            className="
              flex w-full items-center justify-center gap-2
              rounded-md bg-[color:var(--color-primary-action-bg)] px-4 py-2
              text-sm font-semibold text-[color:var(--color-primary-action-fg)]
              hover:bg-[color:var(--color-primary-action-hover)]
              active:bg-[color:var(--color-primary-action-hover)]
              focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary-action-focus)] focus:ring-offset-2
              focus:ring-offset-[color:var(--color-surface-raised)]
              disabled:cursor-not-allowed disabled:bg-[color:var(--color-primary-action-disabled-bg)]
              disabled:text-[color:var(--color-primary-action-disabled-fg)]
              disabled:hover:bg-[color:var(--color-primary-action-disabled-bg)] disabled:opacity-100
              transition-colors duration-150
            "
          >
            {isLoading ? (
              <>
                <span
                  className="h-4 w-4 animate-spin rounded-full border-2 border-[color:var(--color-primary-action-fg)] border-t-transparent"
                  aria-hidden="true"
                />
                Signing in…
              </>
            ) : (
              'Sign in'
            )}
          </button>
        </form>
      </div>
    </div>
  )
}
