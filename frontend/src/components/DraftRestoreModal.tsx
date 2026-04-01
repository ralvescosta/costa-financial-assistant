/**
 * DraftRestoreModal — prompts the user to restore previously saved draft work
 * after forced re-login due to session expiry.
 *
 * Auto-dismisses if the draft TTL has already expired.
 */

import { useEffect, useState } from 'react'

interface DraftRestoreModalProps {
  draftKey: string
  onRestore: (draftData: unknown) => void
  onDiscard: () => void
  /** Time remaining in seconds before the draft expires */
  remainingSeconds?: number
}

export function DraftRestoreModal({
  onRestore,
  onDiscard,
  remainingSeconds,
}: DraftRestoreModalProps) {
  const [secondsLeft, setSecondsLeft] = useState(remainingSeconds)

  useEffect(() => {
    if (secondsLeft == null) return
    if (secondsLeft <= 0) {
      onDiscard()
      return
    }

    const id = setInterval(() => {
      setSecondsLeft((prev) => {
        if (prev == null || prev <= 1) {
          clearInterval(id)
          onDiscard()
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(id)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="draft-restore-title"
      className="
        fixed inset-0 z-50 flex items-center justify-center
        bg-[color:var(--color-overlay)]
      "
    >
      <div
        className="
          w-full max-w-sm rounded-xl
          bg-[color:var(--color-draft-restore-modal)]
          p-6 shadow-xl
        "
      >
        <h2
          id="draft-restore-title"
          className="text-lg font-semibold text-[color:var(--color-text-primary)]"
        >
          Restore your work?
        </h2>
        <p className="mt-2 text-sm text-[color:var(--color-text-secondary)]">
          You have unsaved work from before your session expired.
          {secondsLeft != null && secondsLeft > 0 && (
            <> Available for <span className="font-medium">{secondsLeft}s</span>.</>
          )}
        </p>

        <div className="mt-5 flex gap-3 justify-end">
          <button
            type="button"
            onClick={onDiscard}
            className="
              rounded-md px-4 py-2 text-sm font-medium
              text-[color:var(--color-text-secondary)]
              hover:bg-surface-raised
              focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary)]
            "
          >
            Discard
          </button>
          <button
            type="button"
            onClick={() => onRestore(null)}
            className="
              rounded-md bg-[color:var(--color-primary)] px-4 py-2
              text-sm font-semibold text-white
              hover:bg-[color:var(--color-primary-hover)]
              focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary)] focus:ring-offset-2
            "
          >
            Restore
          </button>
        </div>
      </div>
    </div>
  )
}
