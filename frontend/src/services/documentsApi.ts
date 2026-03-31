import type {
  UploadDocumentResponse,
  ClassifyDocumentResponse,
  ListDocumentsResponse,
  GetDocumentResponse,
  DocumentKind,
} from '@/types/documents'

const BASE = '/api/v1'

/**
 * Uploads a PDF file and creates a pending document record.
 * Sends the file as raw binary with Content-Type: application/pdf.
 */
export async function uploadDocument(
  file: File,
): Promise<UploadDocumentResponse> {
  const bytes = await file.arrayBuffer()
  const res = await fetch(
    `${BASE}/documents/upload?fileName=${encodeURIComponent(file.name)}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/pdf' },
      body: bytes,
    },
  )
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Upload failed: ${res.status}`)
  }
  return res.json() as Promise<UploadDocumentResponse>
}

/**
 * Classifies an existing document as a bill or statement.
 */
export async function classifyDocument(
  documentId: string,
  kind: DocumentKind,
): Promise<ClassifyDocumentResponse> {
  const res = await fetch(`${BASE}/documents/${documentId}/classify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ kind }),
  })
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Classify failed: ${res.status}`)
  }
  return res.json() as Promise<ClassifyDocumentResponse>
}

/**
 * Lists project-scoped documents with optional pagination.
 */
export async function listDocuments(
  pageSize = 20,
  pageToken?: string,
): Promise<ListDocumentsResponse> {
  const params = new URLSearchParams({ pageSize: String(pageSize) })
  if (pageToken) params.set('pageToken', pageToken)
  const res = await fetch(`${BASE}/documents?${params.toString()}`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `List documents failed: ${res.status}`)
  }
  return res.json() as Promise<ListDocumentsResponse>
}

/**
 * Fetches full document detail including extracted bill or statement data.
 */
export async function getDocument(documentId: string): Promise<GetDocumentResponse> {
  const res = await fetch(`${BASE}/documents/${encodeURIComponent(documentId)}`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Get document failed: ${res.status}`)
  }
  return res.json() as Promise<GetDocumentResponse>
}
