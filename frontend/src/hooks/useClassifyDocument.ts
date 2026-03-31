import { useMutation, useQueryClient } from '@tanstack/react-query'
import { classifyDocument } from '@/services/documentsApi'
import type { ClassifyDocumentResponse, DocumentKind } from '@/types/documents'

export interface ClassifyDocumentInput {
  documentId: string
  kind: DocumentKind
}

/**
 * useClassifyDocument — mutation hook that classifies an existing document
 * as a bill or statement. Invalidates the documents list on success.
 *
 * Usage:
 *   const { mutate, isPending } = useClassifyDocument()
 *   mutate({ documentId: 'doc-1', kind: 'bill' })
 */
export function useClassifyDocument() {
  const queryClient = useQueryClient()

  return useMutation<ClassifyDocumentResponse, Error, ClassifyDocumentInput>({
    mutationFn: ({ documentId, kind }) => classifyDocument(documentId, kind),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['documents'] })
    },
  })
}
