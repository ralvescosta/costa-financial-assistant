import type {
  CategoriesResponse,
  ComplianceResponse,
  TimelineResponse,
} from '@/types/history'

const BASE = '/api/v1/history'

function withMonthsParam(months?: number) {
  const params = new URLSearchParams()
  if (typeof months === 'number') {
    params.set('months', String(months))
  }
  const query = params.toString()
  return query ? `?${query}` : ''
}

async function parseError(res: Response, fallback: string): Promise<Error> {
  const body = (await res.json().catch(() => null)) as { title?: string } | null
  return new Error(body?.title ?? `${fallback}: ${res.status}`)
}

export async function getHistoryTimeline(months = 12): Promise<TimelineResponse> {
  const res = await fetch(`${BASE}/timeline${withMonthsParam(months)}`)
  if (!res.ok) {
    throw await parseError(res, 'Get history timeline failed')
  }
  return res.json() as Promise<TimelineResponse>
}

export async function getHistoryCategories(months = 12): Promise<CategoriesResponse> {
  const res = await fetch(`${BASE}/categories${withMonthsParam(months)}`)
  if (!res.ok) {
    throw await parseError(res, 'Get history categories failed')
  }
  return res.json() as Promise<CategoriesResponse>
}

export async function getHistoryCompliance(months = 12): Promise<ComplianceResponse> {
  const res = await fetch(`${BASE}/compliance${withMonthsParam(months)}`)
  if (!res.ok) {
    throw await parseError(res, 'Get history compliance failed')
  }
  return res.json() as Promise<ComplianceResponse>
}
