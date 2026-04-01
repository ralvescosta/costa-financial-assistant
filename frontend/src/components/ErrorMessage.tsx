/**
 * ErrorMessage — displays auth error with optional lockout countdown.
 */

import { useEffect, useState } from 'react'

interface ErrorMessageProps {
  message: string
  /** Unix timestamp (seconds) until the lockout expires. */
  lockoutUntil?: number
}

export function ErrorMessage({ message, lockoutUntil }: ErrorMessageProps) {
  const [remaining, setRemaining] = useState<number | null>(null)

  useEffect(() => {
    if (lockoutUntil == null) {
      setRemaining(null)
      return
    }

    function update() {
      const secs = Math.max(0, lockoutUntil! - Math.floor(Date.now() / 1000))
      setRemaining(secs)
    }

    update()
    const id = setInterval(update, 1000)
    return () => clearInterval(id)
  }, [lockoutUntil])

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="rounded-md bg-danger-bg px-4 py-3 text-sm text-[color:var(--color-input-error)]"
    >
      <p>{message}</p>
      {remaining != null && remaining > 0 && (
        <p className="mt-1 font-medium text-[color:var(--color-lockout-warning)]">
          Try again in {remaining}s
        </p>
      )}
    </div>
  )
}
