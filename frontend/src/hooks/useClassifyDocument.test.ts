import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useClassifyDocument } from './useClassifyDocument'
import type { ClassifyDocumentResponse } from '@/types/documents'
import { createElement } from 'react'
import * as documentsApi from '@/services/documentsApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('@/services/documentsApi', () => ({
  uploadDocument: vi.fn(),
  classifyDocument: vi.fn(),
  listDocuments: vi.fn(),
}))

const mockClassify = vi.mocked(documentsApi.classifyDocument)

// ─── helpers ─────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const classifiedDoc: ClassifyDocumentResponse = {
  id: 'doc-uuid-1',
  projectId: 'proj-1',
  kind: 'bill',
  fileName: 'invoice.pdf',
  analysisStatus: 'pending',
  updatedAt: '2024-01-02T00:00:00Z',
}

// ─── tests ────────────────────────────────────────────────────────────────────

describe('useClassifyDocument', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given an existing document, when classified as bill, then kind is updated', () => {
    it('resolves with updated document on success', async () => {
      // Arrange
      mockClassify.mockResolvedValueOnce(classifiedDoc)
      const { result } = renderHook(() => useClassifyDocument(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ documentId: 'doc-uuid-1', kind: 'bill' })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.kind).toBe('bill')
      expect(mockClassify).toHaveBeenCalledWith('doc-uuid-1', 'bill')
    })
  })

  describe('given an existing document, when classified as statement, then kind is updated', () => {
    it('resolves with kind=statement', async () => {
      // Arrange
      mockClassify.mockResolvedValueOnce({ ...classifiedDoc, kind: 'statement' })
      const { result } = renderHook(() => useClassifyDocument(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ documentId: 'doc-uuid-1', kind: 'statement' })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.kind).toBe('statement')
      expect(mockClassify).toHaveBeenCalledWith('doc-uuid-1', 'statement')
    })
  })

  describe('given a missing document, when classify is called, then a not-found error is surfaced', () => {
    it('transitions to error state when the API rejects with 404', async () => {
      // Arrange
      mockClassify.mockRejectedValueOnce(new Error('Classify failed: 404'))
      const { result } = renderHook(() => useClassifyDocument(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ documentId: 'missing-doc', kind: 'bill' })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toContain('404')
    })
  })

  describe('given an API server error, when classify is called, then a 500 error is surfaced', () => {
    it('transitions to error state when the API rejects with 500', async () => {
      // Arrange
      mockClassify.mockRejectedValueOnce(new Error('Classify failed: 500'))
      const { result } = renderHook(() => useClassifyDocument(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ documentId: 'doc-uuid-1', kind: 'bill' })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toContain('500')
    })
  })
})
