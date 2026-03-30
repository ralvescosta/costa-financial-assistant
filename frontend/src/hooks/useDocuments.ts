import { useQuery } from '@tanstack/react-query'
import { listDocuments } from '@/services/documentsApi'
import type { Document } from '@/types/documents'

export interface UseDocumentsOptions {
  /** Number of items per page. Defaults to 20. */
  pageSize?: number
  /** Opaque token returned from the previous page response. */
  pageToken?: string
  /** Set to false to keep the query paused (e.g. while auth is not ready). */
  enabled?: boolean
}

export interface UseDocumentsResult {
  documents: Document[]
  nextPageToken?: string
  totalCount?: number
  isLoading: boolean
  isError: boolean
  error: Error | null
  refetch: () => void
}

/**
 * useDocuments — query hook that returns a paginated list of project-scoped
 * documents. Uses keyset pagination via `pageToken`.
 *
 * Usage:
 *   const { documents, isLoading } = useDocuments()
 */
export function useDocuments(options: UseDocumentsOptions = {}): UseDocumentsResult {
  const { pageSize = 20, pageToken, enabled = true } = options

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['documents', pageSize, pageToken ?? ''],
    queryFn: () => listDocuments(pageSize, pageToken),
    enabled,
  })

  return {
    documents: data?.documents ?? [],
    nextPageToken: data?.pagination?.nextPageToken,
    totalCount: data?.pagination?.totalCount,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  }
}
