import type {
  ReconciliationSummary,
  ReconciliationLink,
  CreateReconciliationLinkRequest,
} from '@/types/reconciliation'

const BASE = '/api/v1'

/**
 * getReconciliationSummary — fetches the project's reconciliation summary.
 */
export async function getReconciliationSummary(
  periodStart?: string,
  periodEnd?: string,
): Promise<ReconciliationSummary> {
  const params = new URLSearchParams()
  if (periodStart) params.set('periodStart', periodStart)
  if (periodEnd) params.set('periodEnd', periodEnd)

  const qs = params.toString()
  const url = `${BASE}/reconciliation/summary${qs ? `?${qs}` : ''}`

  const res = await fetch(url)
  if (!res.ok) {
    throw new Error(`getReconciliationSummary: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<ReconciliationSummary>
}

/**
 * createReconciliationLink — manually links a transaction line to a bill record.
 */
export async function createReconciliationLink(
  data: CreateReconciliationLinkRequest,
): Promise<ReconciliationLink> {
  const res = await fetch(`${BASE}/reconciliation/links`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) {
    throw new Error(`createReconciliationLink: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<ReconciliationLink>
}
