import { useMutation, useQueryClient } from '@tanstack/react-query'
import { uploadDocument, classifyDocument } from '@/services/documentsApi'
import type { DocumentKind, UploadDocumentResponse } from '@/types/documents'

export interface UseUploadDocumentOptions {
  /**
   * Kind to apply immediately after upload. When provided the hook will
   * run classify as a chained mutation step.
   */
  kind?: DocumentKind
  onSuccess?: (doc: UploadDocumentResponse) => void
  onError?: (err: Error) => void
}

/**
 * useUploadDocument — mutation hook that uploads a PDF and optionally
 * auto-classifies it in a single user action.
 *
 * Usage:
 *   const { mutate, isPending } = useUploadDocument({ kind: 'bill' })
 *   mutate(file)
 */
export function useUploadDocument(options: UseUploadDocumentOptions = {}) {
  const queryClient = useQueryClient()

  return useMutation<UploadDocumentResponse, Error, File>({
    mutationFn: async (file: File) => {
      const uploaded = await uploadDocument(file)

      if (options.kind && options.kind !== 'unspecified') {
        await classifyDocument(uploaded.id, options.kind)
      }

      return uploaded
    },
    onSuccess: (doc) => {
      // Invalidate documents list so any open list view refreshes automatically.
      void queryClient.invalidateQueries({ queryKey: ['documents'] })
      options.onSuccess?.(doc)
    },
    onError: (err) => {
      options.onError?.(err)
    },
  })
}
