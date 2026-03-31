import type {
  PaymentDashboardResponse,
  MarkBillPaidResponse,
  CyclePreference,
  SetPreferredDayRequest,
} from '@/types/payments'

const BASE = '/api/v1'

/**
 * Fetches outstanding and overdue bills for the project's payment cycle.
 */
export async function getPaymentDashboard(
  cycleStart?: string,
  cycleEnd?: string,
  pageSize = 20,
  pageToken?: string,
): Promise<PaymentDashboardResponse> {
  const params = new URLSearchParams({ pageSize: String(pageSize) })
  if (cycleStart) params.set('cycleStart', cycleStart)
  if (cycleEnd) params.set('cycleEnd', cycleEnd)
  if (pageToken) params.set('pageToken', pageToken)

  const res = await fetch(`${BASE}/bills/payment-dashboard?${params.toString()}`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Get payment dashboard failed: ${res.status}`)
  }
  return res.json() as Promise<PaymentDashboardResponse>
}

/**
 * Idempotently marks a bill as paid.
 */
export async function markBillPaid(billId: string): Promise<MarkBillPaidResponse> {
  const res = await fetch(`${BASE}/bills/${encodeURIComponent(billId)}/mark-paid`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  })
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Mark bill paid failed: ${res.status}`)
  }
  return res.json() as Promise<MarkBillPaidResponse>
}

/**
 * Returns the project's preferred payment day of month.
 */
export async function getPreferredDay(): Promise<CyclePreference> {
  const res = await fetch(`${BASE}/payment-cycle/preferred-day`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Get preferred day failed: ${res.status}`)
  }
  return res.json() as Promise<CyclePreference>
}

/**
 * Creates or updates the project's preferred payment day of month.
 */
export async function setPreferredDay(
  data: SetPreferredDayRequest,
): Promise<CyclePreference> {
  const res = await fetch(`${BASE}/payment-cycle/preferred-day`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Set preferred day failed: ${res.status}`)
  }
  return res.json() as Promise<CyclePreference>
}
