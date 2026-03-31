// Payments types for the financial bill organizer

export type PaymentStatus =
  | 'PAYMENT_STATUS_UNSPECIFIED'
  | 'PAYMENT_STATUS_UNPAID'
  | 'PAYMENT_STATUS_PAID'
  | 'PAYMENT_STATUS_OVERDUE'

export interface BillRecord {
  id: string
  projectId: string
  documentId: string
  billTypeId?: string
  dueDate: string
  amountDue: string
  pixPayload?: string
  pixQrImageRef?: string
  barcode?: string
  paymentStatus: PaymentStatus
  paidAt?: string
  markedPaidBy?: string
  createdAt: string
  updatedAt: string
}

export interface BillType {
  id: string
  projectId: string
  name: string
}

export interface PaymentDashboardEntry {
  bill: BillRecord
  billType?: BillType
  isOverdue: boolean
  daysUntilDue: number
}

export interface PaymentDashboardResponse {
  entries: PaymentDashboardEntry[]
  nextPageToken?: string
}

export interface MarkBillPaidResponse {
  bill: BillRecord
}

export interface CyclePreference {
  projectId: string
  preferredDayOfMonth: number
  updatedAt: string
}

export interface SetPreferredDayRequest {
  preferredDayOfMonth: number
}
