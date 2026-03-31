// Document types for the financial bill organizer

export type DocumentKind = 'unspecified' | 'bill' | 'statement'

export type AnalysisStatus =
  | 'pending'
  | 'processing'
  | 'analysed'
  | 'analysis_failed'

export interface Document {
  id: string
  projectId: string
  uploadedBy: string
  kind: DocumentKind
  fileName: string
  analysisStatus: AnalysisStatus
  storageProvider: string
  uploadedAt: string
  updatedAt: string
}

export interface UploadDocumentResponse {
  id: string
  projectId: string
  uploadedBy: string
  kind: DocumentKind
  fileName: string
  analysisStatus: AnalysisStatus
  storageProvider: string
  uploadedAt: string
  updatedAt: string
}

export interface ClassifyDocumentResponse {
  id: string
  projectId: string
  kind: DocumentKind
  fileName: string
  analysisStatus: AnalysisStatus
  updatedAt: string
}

export interface ListDocumentsResponse {
  documents: Document[]
  pagination: {
    nextPageToken?: string
    totalCount?: number
  }
}

export interface BillRecord {
  id: string
  dueDate: string
  amountDue: string
  pixPayload?: string
  pixQrImageRef?: string
  barcode?: string
  paymentStatus: string
  paidAt?: string
}

export interface TransactionLine {
  id: string
  transactionDate: string
  description: string
  amount: string
  direction: string
  reconciliationStatus: string
}

export interface StatementRecord {
  id: string
  bankAccountId?: string
  periodStart: string
  periodEnd: string
  lines: TransactionLine[]
}

export interface DocumentDetail extends Document {
  billRecord?: BillRecord
  statementRecord?: StatementRecord
}

export interface GetDocumentResponse extends DocumentDetail { }

