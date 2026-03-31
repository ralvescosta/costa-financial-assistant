import { useState } from 'react'
import { useBankAccounts } from '@/hooks/useBankAccounts'

/** SettingsPage — bank account label management for statement attribution. */
export default function SettingsPage() {
  const [newLabel, setNewLabel] = useState('')
  const [feedback, setFeedback] = useState<{ msg: string; isError: boolean } | null>(null)

  const { bankAccounts, isLoading, isError, create, remove, isCreating, isDeleting } =
    useBankAccounts()

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = newLabel.trim()
    if (!trimmed) return
    setFeedback(null)
    try {
      await create({ label: trimmed })
      setNewLabel('')
      setFeedback({ msg: `Bank account "${trimmed}" added.`, isError: false })
    } catch (err) {
      setFeedback({
        msg: err instanceof Error ? err.message : 'Failed to add bank account.',
        isError: true,
      })
    }
  }

  async function handleDelete(id: string, label: string) {
    setFeedback(null)
    try {
      await remove(id)
      setFeedback({ msg: `Removed "${label}".`, isError: false })
    } catch (err) {
      setFeedback({
        msg: err instanceof Error ? err.message : 'Failed to remove bank account.',
        isError: true,
      })
    }
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-semibold text-text-primary">Settings</h1>

      {/* ── Bank Accounts ────────────────────────────────────────────── */}
      <section aria-label="Bank account labels" className="rounded-lg border border-border bg-surface p-6">
        <h2 className="mb-1 text-lg font-medium text-text-primary">Bank Accounts</h2>
        <p className="mb-4 text-sm text-text-secondary">
          Manage bank account labels used to attribute bank statements.
        </p>

        {/* ── Add form ──────────────────────────────────────────────── */}
        <form onSubmit={(e) => void handleCreate(e)} className="mb-6 flex gap-2">
          <input
            type="text"
            value={newLabel}
            onChange={(e) => setNewLabel(e.target.value)}
            placeholder="e.g. Checking Account"
            maxLength={100}
            className="flex-1 rounded border border-border bg-surface px-3 py-1.5 text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-primary"
            aria-label="Bank account label"
          />
          <button
            type="submit"
            disabled={isCreating || !newLabel.trim()}
            className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isCreating ? 'Adding…' : 'Add'}
          </button>
        </form>

        {/* ── Feedback ──────────────────────────────────────────────── */}
        {feedback && (
          <p
            role="status"
            className={`mb-4 rounded px-3 py-2 text-sm ${feedback.isError ? 'bg-danger/10 text-danger' : 'bg-success/10 text-success'}`}
          >
            {feedback.msg}
          </p>
        )}

        {/* ── List ──────────────────────────────────────────────────── */}
        {isLoading && (
          <p className="text-sm text-text-secondary">Loading bank accounts…</p>
        )}
        {isError && (
          <p className="text-sm text-danger">Failed to load bank accounts.</p>
        )}
        {!isLoading && !isError && bankAccounts.length === 0 && (
          <p className="text-sm text-text-secondary">No bank accounts added yet.</p>
        )}
        {bankAccounts.length > 0 && (
          <ul className="divide-y divide-border rounded border border-border" role="list">
            {bankAccounts.map((account) => (
              <li
                key={account.id}
                className="flex items-center justify-between px-4 py-3"
              >
                <span className="text-sm text-text-primary">{account.label}</span>
                <button
                  onClick={() => void handleDelete(account.id, account.label)}
                  disabled={isDeleting}
                  className="rounded px-2 py-1 text-xs text-danger hover:bg-danger/10 disabled:cursor-not-allowed disabled:opacity-50"
                  aria-label={`Remove ${account.label}`}
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </main>
  )
}

