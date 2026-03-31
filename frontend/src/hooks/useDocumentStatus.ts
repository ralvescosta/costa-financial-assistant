import { useQuery } from '@tanstack/react-query'
import { getDocument } from '@/services/documentsApi'
import type { DocumentDetail } from '@/types/documents'

/** Interval (ms) to poll while a document is still being processed. */
const POLLING_INTERVAL_MS = 2_000

export interface UseDocumentStatusOptions {
  /** Set to false to keep the query disabled (e.g. while auth is not ready). */
  enabled?: boolean
}

export interface UseDocumentStatusResult {
  document: DocumentDetail | undefined
  isLoading: boolean
  isError: boolean
  error: Error | null
  refetch: () => void
}

/**
 * useDocumentStatus — query hook that returns document detail and polls while
 * the analysis is in progress (`pending` or `processing` state).
 *
 * Polling stops automatically once the document reaches a terminal state
 * (`analysed` or `analysis_failed`).
 *
 * Usage:
 *   const { document, isLoading } = useDocumentStatus(documentId)
 */
export function useDocumentStatus(
  documentId: string,
  options: UseDocumentStatusOptions = {},
): UseDocumentStatusResult {
  const { enabled = true } = options

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['document', documentId],
    queryFn: () => getDocument(documentId),
    enabled: enabled && Boolean(documentId),
    // Poll while the document is in a transient state; stop on terminal states.
    refetchInterval: (query) => {
      const status = query.state.data?.analysisStatus
      if (status === 'pending' || status === 'processing') {
        return POLLING_INTERVAL_MS
      }
      return false
    },
  })

  return {
    document: data,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  }
}
