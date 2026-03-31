import { useParams, Link } from 'react-router-dom'
import { useDocumentStatus } from '@/hooks/useDocumentStatus'
import type { BillRecord, StatementRecord } from '@/types/documents'

// ─── sub-components ───────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: string }) {
  const color: Record<string, string> = {
    pending: 'bg-surface-secondary text-text-secondary',
    processing: 'bg-warning/20 text-warning',
    analysed: 'bg-success/20 text-success',
    analysis_failed: 'bg-danger/20 text-danger',
  }
  return (
    <span
      className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${color[status] ?? 'bg-surface-secondary text-text-secondary'}`}
    >
      {status}
    </span>
  )
}

function BillDetail({ record }: { record: BillRecord }) {
  return (
    <section aria-label="Extracted bill data">
      <h2 className="mb-3 text-lg font-semibold text-text-primary">Bill Details</h2>
      <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
        <dt className="text-text-secondary">Due Date</dt>
        <dd className="text-text-primary">{record.dueDate || '—'}</dd>

        <dt className="text-text-secondary">Amount Due</dt>
        <dd className="text-text-primary">{record.amountDue || '—'}</dd>

        <dt className="text-text-secondary">Payment Status</dt>
        <dd className="text-text-primary capitalize">{record.paymentStatus}</dd>

        {record.paidAt && (
          <>
            <dt className="text-text-secondary">Paid At</dt>
            <dd className="text-text-primary">{record.paidAt}</dd>
          </>
        )}

        {record.barcode && (
          <>
            <dt className="text-text-secondary">Barcode</dt>
            <dd className="break-all font-mono text-xs text-text-primary">{record.barcode}</dd>
          </>
        )}

        {record.pixPayload && (
          <>
            <dt className="text-text-secondary">Pix Payload</dt>
            <dd className="break-all font-mono text-xs text-text-primary">{record.pixPayload}</dd>
          </>
        )}
      </dl>
    </section>
  )
}

function StatementDetail({ record }: { record: StatementRecord }) {
  return (
    <section aria-label="Extracted statement data">
      <h2 className="mb-3 text-lg font-semibold text-text-primary">Statement Details</h2>
      <dl className="mb-4 grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
        <dt className="text-text-secondary">Period Start</dt>
        <dd className="text-text-primary">{record.periodStart}</dd>

        <dt className="text-text-secondary">Period End</dt>
        <dd className="text-text-primary">{record.periodEnd}</dd>

        {record.bankAccountId && (
          <>
            <dt className="text-text-secondary">Bank Account</dt>
            <dd className="text-text-primary">{record.bankAccountId}</dd>
          </>
        )}
      </dl>

      <h3 className="mb-2 text-sm font-semibold text-text-primary">
        Transactions ({record.lines.length})
      </h3>

      {record.lines.length === 0 ? (
        <p className="text-sm text-text-secondary">No transactions found.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="pb-2 pr-4 font-medium text-text-secondary">Date</th>
                <th className="pb-2 pr-4 font-medium text-text-secondary">Description</th>
                <th className="pb-2 pr-4 text-right font-medium text-text-secondary">Amount</th>
                <th className="pb-2 font-medium text-text-secondary">Direction</th>
              </tr>
            </thead>
            <tbody>
              {record.lines.map((line) => (
                <tr key={line.id} className="border-b border-border/50">
                  <td className="py-1.5 pr-4 text-text-secondary">{line.transactionDate}</td>
                  <td className="py-1.5 pr-4 text-text-primary">{line.description}</td>
                  <td className="py-1.5 pr-4 text-right font-mono text-text-primary">
                    {line.amount}
                  </td>
                  <td className="py-1.5 capitalize text-text-secondary">{line.direction}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}

// ─── page ─────────────────────────────────────────────────────────────────────

/** DocumentDetailPage — shows document metadata, analysis status, and extracted data. */
export default function DocumentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { document, isLoading, isError, error } = useDocumentStatus(id ?? '')

  if (!id) {
    return (
      <main className="mx-auto max-w-3xl px-4 py-8">
        <p className="text-sm text-danger">No document ID provided.</p>
      </main>
    )
  }

  if (isLoading) {
    return (
      <main className="mx-auto max-w-3xl px-4 py-8">
        <p className="text-sm text-text-secondary">Loading document…</p>
      </main>
    )
  }

  if (isError || !document) {
    return (
      <main className="mx-auto max-w-3xl px-4 py-8">
        <p className="text-sm text-danger">
          {error?.message ?? 'Document not found.'}
        </p>
        <Link to="/upload" className="mt-4 inline-block text-sm text-primary underline">
          Back to documents
        </Link>
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-8">
      {/* ── Header ──────────────────────────────────────────────────── */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-text-primary">{document.fileName}</h1>
        <Link to="/upload" className="text-sm text-primary underline">
          Back
        </Link>
      </div>

      {/* ── Metadata ────────────────────────────────────────────────── */}
      <section aria-label="Document metadata" className="mb-6">
        <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
          <dt className="text-text-secondary">ID</dt>
          <dd className="font-mono text-xs text-text-primary">{document.id}</dd>

          <dt className="text-text-secondary">Kind</dt>
          <dd className="capitalize text-text-primary">{document.kind}</dd>

          <dt className="text-text-secondary">Analysis Status</dt>
          <dd>
            <StatusBadge status={document.analysisStatus} />
          </dd>

          <dt className="text-text-secondary">Uploaded At</dt>
          <dd className="text-text-primary">{document.uploadedAt}</dd>
        </dl>
      </section>

      <hr className="mb-6 border-border" />

      {/* ── Extraction data ─────────────────────────────────────────── */}
      {document.analysisStatus === 'analysed' && document.billRecord && (
        <BillDetail record={document.billRecord} />
      )}

      {document.analysisStatus === 'analysed' && document.statementRecord && (
        <StatementDetail record={document.statementRecord} />
      )}

      {document.analysisStatus === 'pending' && (
        <p className="text-sm text-text-secondary">
          Document is queued for analysis. This page will refresh automatically.
        </p>
      )}

      {document.analysisStatus === 'processing' && (
        <p className="text-sm text-text-secondary">
          Analysis in progress… This page will refresh automatically.
        </p>
      )}

      {document.analysisStatus === 'analysis_failed' && (
        <p className="text-sm text-danger">
          Analysis failed. Please re-upload the document or contact support.
        </p>
      )}
    </main>
  )
}
