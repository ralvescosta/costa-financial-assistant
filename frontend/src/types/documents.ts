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
