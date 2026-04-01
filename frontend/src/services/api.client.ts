/**
 * Base HTTP client for all BFF API requests.
 *
 * All requests include `credentials: 'include'` so HTTP-only session cookies
 * are sent automatically. The BFF sets cookies with SameSite=Strict and
 * HttpOnly — the frontend never accesses them via JavaScript.
 *
 * CSRF tokens (when required) are read from the auth context and injected
 * via the X-CSRF-Token header by callers.
 */

export interface ApiFetchOptions extends Omit<RequestInit, 'body'> {
  /** When provided the body is serialised to JSON and Content-Type is set. */
  json?: unknown
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public readonly body: any = null,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

const BASE = '/api'

/**
 * Perform an authenticated fetch against the BFF.
 * Throws `ApiError` for non-2xx responses.
 */
export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const { json, headers, ...rest } = options

  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    headers: {
      ...(json !== undefined ? { 'Content-Type': 'application/json' } : {}),
      ...headers,
    },
    body: json !== undefined ? JSON.stringify(json) : undefined,
    ...rest,
  })

  if (!res.ok) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: any = await res.json().catch(() => null)
    const message: string =
      body?.error?.message ?? body?.title ?? `Request failed: ${res.status}`
    throw new ApiError(message, res.status, body)
  }

  return res.json() as Promise<T>
}

/**
 * Convenience wrapper for POST requests with a JSON body.
 */
export function apiPost<T>(
  path: string,
  body: unknown,
  options: Omit<ApiFetchOptions, 'json' | 'method'> = {},
): Promise<T> {
  return apiFetch<T>(path, { method: 'POST', json: body, ...options })
}

/**
 * Convenience wrapper for POST requests with no body (e.g. logout, refresh).
 */
export function apiPostEmpty<T>(
  path: string,
  options: Omit<ApiFetchOptions, 'method'> = {},
): Promise<T> {
  return apiFetch<T>(path, { method: 'POST', ...options })
}
