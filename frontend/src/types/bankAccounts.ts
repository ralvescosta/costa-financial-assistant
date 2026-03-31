// Bank account types for the financial bill organizer settings

export interface BankAccount {
  id: string
  projectId: string
  label: string
  createdBy?: string
  createdAt: string
  updatedAt: string
}

export interface CreateBankAccountRequest {
  label: string
}

export interface CreateBankAccountResponse extends BankAccount { }

export interface ListBankAccountsResponse {
  items: BankAccount[]
}
