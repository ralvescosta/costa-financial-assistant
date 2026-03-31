import { useState } from 'react'
import {
  usePaymentDashboard,
  useMarkBillPaid,
  usePreferredDay,
  useSetPreferredDay,
} from '@/hooks/usePaymentDashboard'
import type { PaymentDashboardEntry } from '@/types/payments'

// ─── Sub-components ───────────────────────────────────────────────────────────

function DueBadge({ isOverdue, daysUntilDue }: { isOverdue: boolean; daysUntilDue: number }) {
  if (isOverdue) {
    return (
      <span className="inline-block rounded-full bg-danger/20 px-2 py-0.5 text-xs font-medium text-danger">
        Overdue ({Math.abs(daysUntilDue)}d ago)
      </span>
    )
  }
  if (daysUntilDue <= 3) {
    return (
      <span className="inline-block rounded-full bg-warning/20 px-2 py-0.5 text-xs font-medium text-warning">
        Due in {daysUntilDue}d
      </span>
    )
  }
  return (
    <span className="inline-block rounded-full bg-surface-secondary px-2 py-0.5 text-xs font-medium text-text-secondary">
      Due in {daysUntilDue}d
    </span>
  )
}

function BillRow({
  entry,
  onMarkPaid,
  marking,
}: {
  entry: PaymentDashboardEntry
  onMarkPaid: (id: string) => void
  marking: boolean
}) {
  const { bill, billType, isOverdue, daysUntilDue } = entry
  return (
    <tr className="border-b border-border/50 last:border-0">
      <td className="py-3 pr-4">
        <p className="font-medium text-text-primary">{billType?.name ?? '—'}</p>
        <p className="text-xs text-text-secondary">{bill.dueDate}</p>
      </td>
      <td className="py-3 pr-4 text-right font-mono text-text-primary">
        R$ {bill.amountDue}
      </td>
      <td className="py-3 pr-4">
        <DueBadge isOverdue={isOverdue} daysUntilDue={daysUntilDue} />
      </td>
      <td className="py-3 text-right">
        <button
          type="button"
          onClick={() => onMarkPaid(bill.id)}
          disabled={marking}
          className="rounded-md bg-primary px-3 py-1 text-xs font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          aria-label={`Mark bill ${bill.id} as paid`}
        >
          Mark paid
        </button>
      </td>
    </tr>
  )
}

function PreferredDayForm() {
  const { data: pref, isPending: loading } = usePreferredDay()
  const { mutate: save, isPending: saving } = useSetPreferredDay()
  const [day, setDay] = useState<number>(pref?.preferredDayOfMonth ?? 1)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    save({ preferredDayOfMonth: day })
  }

  if (loading) {
    return <p className="text-sm text-text-secondary">Loading preference…</p>
  }

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-3">
      <label htmlFor="preferred-day" className="text-sm text-text-secondary">
        Preferred payment day
      </label>
      <select
        id="preferred-day"
        value={day}
        onChange={(e) => setDay(Number(e.target.value))}
        className="rounded-md border border-border bg-surface px-2 py-1 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-primary"
      >
        {Array.from({ length: 28 }, (_, i) => i + 1).map((d) => (
          <option key={d} value={d}>
            {d}
          </option>
        ))}
      </select>
      <button
        type="submit"
        disabled={saving}
        className="rounded-md bg-primary px-3 py-1 text-xs font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {saving ? 'Saving…' : 'Save'}
      </button>
    </form>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────

/** PaymentDashboardPage — lists outstanding bills with overdue highlighting and mark-paid actions. */
export default function PaymentDashboardPage() {
  const { data, isPending, isError } = usePaymentDashboard()
  const { mutate: markPaid, isPending: marking } = useMarkBillPaid()

  return (
    <main className="mx-auto max-w-4xl px-4 py-8">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl font-semibold text-text-primary">Payment Dashboard</h1>
        <PreferredDayForm />
      </div>

      {isPending && (
        <p className="text-sm text-text-secondary" role="status">
          Loading bills…
        </p>
      )}

      {isError && (
        <p className="text-sm text-danger" role="alert">
          Failed to load payment dashboard. Please try again.
        </p>
      )}

      {!isPending && !isError && data?.entries.length === 0 && (
        <p className="text-sm text-text-secondary">No outstanding bills for this cycle.</p>
      )}

      {!isPending && !isError && data && data.entries.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="w-full text-left text-sm" aria-label="Outstanding bills">
            <thead className="bg-surface-secondary">
              <tr>
                <th className="px-4 py-3 font-medium text-text-secondary">Bill</th>
                <th className="px-4 py-3 text-right font-medium text-text-secondary">Amount</th>
                <th className="px-4 py-3 font-medium text-text-secondary">Status</th>
                <th className="px-4 py-3 text-right font-medium text-text-secondary">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50 px-4">
              {data.entries.map((entry) => (
                <BillRow
                  key={entry.bill.id}
                  entry={entry}
                  onMarkPaid={markPaid}
                  marking={marking}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
  )
}
