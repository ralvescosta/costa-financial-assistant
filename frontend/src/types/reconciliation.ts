// Reconciliation types for the financial bill organizer

export type ReconciliationStatus =
  | 'unmatched'
  | 'matched_auto'
  | 'matched_manual'
  | 'ambiguous'

export type ReconciliationLinkType = 'auto' | 'manual'

export interface ReconciliationEntry {
  transactionLineId: string
  transactionDate: string
  description: string
  amount: string
  direction: string
  reconciliationStatus: ReconciliationStatus
  linkedBillId?: string
  linkedBillDueDate?: string
  linkedBillAmount?: string
  linkType?: ReconciliationLinkType
}

export interface ReconciliationSummary {
  projectId: string
  periodStart?: string
  periodEnd?: string
  entries: ReconciliationEntry[]
}

export interface ReconciliationLink {
  id: string
  projectId: string
  transactionLineId: string
  billRecordId: string
  linkType: ReconciliationLinkType
  linkedBy?: string
  createdAt: string
}

export interface CreateReconciliationLinkRequest {
  transactionLineId: string
  billRecordId: string
}
