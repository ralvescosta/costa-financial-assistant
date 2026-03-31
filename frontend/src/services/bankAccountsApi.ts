import type {
  CreateBankAccountRequest,
  CreateBankAccountResponse,
  ListBankAccountsResponse,
} from '@/types/bankAccounts'

const BASE = '/api/v1'

/**
 * Lists all project-scoped bank account labels.
 */
export async function listBankAccounts(): Promise<ListBankAccountsResponse> {
  const res = await fetch(`${BASE}/bank-accounts`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `List bank accounts failed: ${res.status}`)
  }
  return res.json() as Promise<ListBankAccountsResponse>
}

/**
 * Creates a new project-scoped bank account label.
 */
export async function createBankAccount(
  req: CreateBankAccountRequest,
): Promise<CreateBankAccountResponse> {
  const res = await fetch(`${BASE}/bank-accounts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Create bank account failed: ${res.status}`)
  }
  return res.json() as Promise<CreateBankAccountResponse>
}

/**
 * Deletes a project-scoped bank account label by ID.
 */
export async function deleteBankAccount(bankAccountId: string): Promise<void> {
  const res = await fetch(
    `${BASE}/bank-accounts/${encodeURIComponent(bankAccountId)}`,
    { method: 'DELETE' },
  )
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Delete bank account failed: ${res.status}`)
  }
}
