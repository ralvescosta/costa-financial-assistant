import { useState } from 'react'
import { useReconciliationSummary, useCreateReconciliationLink } from '@/hooks/useReconciliation'
import type { ReconciliationEntry } from '@/types/reconciliation'

// ─── Status badge ─────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: ReconciliationEntry['reconciliationStatus'] }) {
  const styles: Record<string, string> = {
    matched_auto: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    matched_manual: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    unmatched: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    ambiguous: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
  }
  const labels: Record<string, string> = {
    matched_auto: 'Auto',
    matched_manual: 'Manual',
    unmatched: 'Unmatched',
    ambiguous: 'Ambiguous',
  }
  return (
    <span
      className={`inline-block rounded px-2 py-0.5 text-xs font-medium ${styles[status] ?? ''}`}
    >
      {labels[status] ?? status}
    </span>
  )
}

// ─── Manual link form ─────────────────────────────────────────────────────────

function ManualLinkForm({
  transactionLineId,
  onClose,
}: {
  transactionLineId: string
  onClose: () => void
}) {
  const [billRecordId, setBillRecordId] = useState('')
  const { mutate, isPending, isError, error } = useCreateReconciliationLink({
    onSuccess: () => onClose(),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (billRecordId.trim()) {
      mutate({ transactionLineId, billRecordId: billRecordId.trim() })
    }
  }

  return (
    <form onSubmit={handleSubmit} className="mt-2 flex flex-col gap-2">
      <label className="text-xs text-text-secondary">
        Bill Record ID
        <input
          type="text"
          className="mt-1 block w-full rounded border border-border bg-surface px-2 py-1 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-primary"
          placeholder="UUID of bill record"
          value={billRecordId}
          onChange={(e) => setBillRecordId(e.target.value)}
          required
        />
      </label>
      {isError && (
        <p className="text-xs text-danger">{error?.message ?? 'Failed to create link'}</p>
      )}
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={isPending || !billRecordId.trim()}
          className="rounded bg-primary px-3 py-1 text-xs font-medium text-white disabled:opacity-50"
        >
          {isPending ? 'Linking…' : 'Link'}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="rounded border border-border px-3 py-1 text-xs text-text-secondary hover:bg-surface-alt"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}

// ─── Period filter form ───────────────────────────────────────────────────────

function PeriodFilterForm({
  onApply,
}: {
  onApply: (start?: string, end?: string) => void
}) {
  const [start, setStart] = useState('')
  const [end, setEnd] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onApply(start || undefined, end || undefined)
  }

  const handleClear = () => {
    setStart('')
    setEnd('')
    onApply(undefined, undefined)
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-wrap gap-3 items-end">
      <label className="flex flex-col gap-1 text-xs text-text-secondary">
        From
        <input
          type="date"
          className="rounded border border-border bg-surface px-2 py-1 text-sm text-text-primary"
          value={start}
          onChange={(e) => setStart(e.target.value)}
        />
      </label>
      <label className="flex flex-col gap-1 text-xs text-text-secondary">
        To
        <input
          type="date"
          className="rounded border border-border bg-surface px-2 py-1 text-sm text-text-primary"
          value={end}
          onChange={(e) => setEnd(e.target.value)}
        />
      </label>
      <button
        type="submit"
        className="rounded bg-primary px-3 py-1.5 text-xs font-medium text-white"
      >
        Apply
      </button>
      <button
        type="button"
        onClick={handleClear}
        className="rounded border border-border px-3 py-1.5 text-xs text-text-secondary hover:bg-surface-alt"
      >
        Clear
      </button>
    </form>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function ReconciliationPage() {
  const [periodStart, setPeriodStart] = useState<string | undefined>()
  const [periodEnd, setPeriodEnd] = useState<string | undefined>()
  const [linkingTxId, setLinkingTxId] = useState<string | null>(null)

  const { data, isPending, isError, error } = useReconciliationSummary(periodStart, periodEnd)

  const handleApplyFilter = (start?: string, end?: string) => {
    setPeriodStart(start)
    setPeriodEnd(end)
  }

  return (
    <main className="mx-auto max-w-5xl px-4 py-8">
      <h1 className="text-2xl font-semibold text-text-primary">Reconciliation</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Match statement transactions to bills automatically or manually.
      </p>

      <section className="mt-6">
        <PeriodFilterForm onApply={handleApplyFilter} />
      </section>

      <section className="mt-6">
        {isPending && (
          <p className="text-sm text-text-secondary">Loading transactions…</p>
        )}

        {isError && (
          <p className="text-sm text-danger">{error?.message ?? 'Failed to load reconciliation data.'}</p>
        )}

        {data && data.entries.length === 0 && (
          <p className="text-sm text-text-secondary">No transaction lines found for the selected period.</p>
        )}

        {data && data.entries.length > 0 && (
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="min-w-full divide-y divide-border text-sm">
              <thead className="bg-surface-alt">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase tracking-wide">
                    Date
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase tracking-wide">
                    Description
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase tracking-wide">
                    Amount
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-text-secondary uppercase tracking-wide">
                    Status
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase tracking-wide">
                    Linked Bill
                  </th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-surface">
                {data.entries.map((entry) => (
                  <tr key={entry.transactionLineId}>
                    <td className="px-4 py-3 text-text-primary whitespace-nowrap">
                      {entry.transactionDate}
                    </td>
                    <td className="px-4 py-3 text-text-primary max-w-xs truncate">
                      {entry.description}
                    </td>
                    <td className="px-4 py-3 text-right tabular-nums text-text-primary whitespace-nowrap">
                      {entry.direction === 'debit' ? '-' : '+'}
                      {Number(entry.amount).toLocaleString('pt-BR', {
                        style: 'currency',
                        currency: 'BRL',
                      })}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <StatusBadge status={entry.reconciliationStatus} />
                    </td>
                    <td className="px-4 py-3 text-text-secondary text-xs">
                      {entry.linkedBillId ? (
                        <span title={entry.linkedBillId}>
                          Due {entry.linkedBillDueDate ?? '—'} · R${' '}
                          {entry.linkedBillAmount ?? '—'}
                        </span>
                      ) : (
                        '—'
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {(entry.reconciliationStatus === 'unmatched' ||
                        entry.reconciliationStatus === 'ambiguous') && (
                          <div>
                            {linkingTxId === entry.transactionLineId ? (
                              <ManualLinkForm
                                transactionLineId={entry.transactionLineId}
                                onClose={() => setLinkingTxId(null)}
                              />
                            ) : (
                              <button
                                className="rounded border border-border px-2 py-1 text-xs text-text-secondary hover:bg-surface-alt"
                                onClick={() => setLinkingTxId(entry.transactionLineId)}
                              >
                                Link manually
                              </button>
                            )}
                          </div>
                        )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  )
}
