import { useState } from 'react'

import { useHistoryDashboard } from '@/hooks/useHistoryDashboard'

export default function HistoryDashboardPage() {
  const [months, setMonths] = useState(12)
  const { timelineQuery, categoriesQuery, complianceQuery, isPending, isError } =
    useHistoryDashboard(months)

  return (
    <main className="mx-auto max-w-6xl px-4 py-8">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-text-primary">Financial History</h1>
          <p className="mt-1 text-sm text-text-secondary">
            Timeline, category mix, and payment compliance for your active project.
          </p>
        </div>
        <label className="flex items-center gap-2 text-sm text-text-secondary">
          Look-back
          <select
            value={months}
            onChange={(e) => setMonths(Number(e.target.value))}
            className="rounded-md border border-border bg-surface px-2 py-1 text-sm text-text-primary"
            aria-label="History look-back window"
          >
            <option value={3}>Last 3 months</option>
            <option value={6}>Last 6 months</option>
            <option value={12}>Last 12 months</option>
            <option value={0}>All history</option>
          </select>
        </label>
      </div>

      {isPending && (
        <p className="mt-6 text-sm text-text-secondary" role="status">
          Loading history metrics...
        </p>
      )}

      {isError && (
        <p className="mt-6 text-sm text-danger" role="alert">
          Unable to load one or more history panels. Try refreshing the page.
        </p>
      )}

      {!isPending && !isError && (
        <section className="mt-6 grid gap-6 lg:grid-cols-2">
          <article className="rounded-lg border border-border bg-surface p-4">
            <h2 className="text-lg font-medium text-text-primary">Monthly Timeline</h2>
            <div className="mt-3 overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-text-secondary">
                  <tr>
                    <th className="py-2">Month</th>
                    <th className="py-2 text-right">Bills</th>
                    <th className="py-2 text-right">Total</th>
                  </tr>
                </thead>
                <tbody>
                  {(timelineQuery.data?.timeline ?? []).map((row) => (
                    <tr key={row.month} className="border-t border-border/60">
                      <td className="py-2 text-text-primary">{row.month}</td>
                      <td className="py-2 text-right text-text-primary">{row.billCount}</td>
                      <td className="py-2 text-right font-mono text-text-primary">
                        R$ {row.totalAmount}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </article>

          <article className="rounded-lg border border-border bg-surface p-4">
            <h2 className="text-lg font-medium text-text-primary">Category Breakdown</h2>
            <div className="mt-3 overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-text-secondary">
                  <tr>
                    <th className="py-2">Month</th>
                    <th className="py-2">Category</th>
                    <th className="py-2 text-right">Total</th>
                  </tr>
                </thead>
                <tbody>
                  {(categoriesQuery.data?.categories ?? []).map((row) => (
                    <tr key={`${row.month}-${row.billTypeName}`} className="border-t border-border/60">
                      <td className="py-2 text-text-primary">{row.month}</td>
                      <td className="py-2 text-text-primary">{row.billTypeName}</td>
                      <td className="py-2 text-right font-mono text-text-primary">
                        R$ {row.totalAmount}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </article>

          <article className="rounded-lg border border-border bg-surface p-4 lg:col-span-2">
            <h2 className="text-lg font-medium text-text-primary">Payment Compliance</h2>
            <div className="mt-3 overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-text-secondary">
                  <tr>
                    <th className="py-2">Month</th>
                    <th className="py-2 text-right">On-time</th>
                    <th className="py-2 text-right">Overdue</th>
                    <th className="py-2 text-right">Rate</th>
                  </tr>
                </thead>
                <tbody>
                  {(complianceQuery.data?.compliance ?? []).map((row) => (
                    <tr key={row.month} className="border-t border-border/60">
                      <td className="py-2 text-text-primary">{row.month}</td>
                      <td className="py-2 text-right text-success">{row.paidOnTime}</td>
                      <td className="py-2 text-right text-danger">{row.overdue}</td>
                      <td className="py-2 text-right font-mono text-text-primary">
                        {row.complianceRate}%
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </article>
        </section>
      )}
    </main>
  )
}
