import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useDocumentStatus } from './useDocumentStatus'
import type { DocumentDetail } from '@/types/documents'
import { createElement } from 'react'
import * as documentsApi from '@/services/documentsApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('@/services/documentsApi', () => ({
  uploadDocument: vi.fn(),
  classifyDocument: vi.fn(),
  listDocuments: vi.fn(),
  getDocument: vi.fn(),
}))

const mockGetDocument = vi.mocked(documentsApi.getDocument)

// ─── helpers ─────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const pendingDoc: DocumentDetail = {
  id: 'doc-uuid-1',
  projectId: 'proj-1',
  uploadedBy: 'user-1',
  kind: 'bill',
  fileName: 'invoice.pdf',
  analysisStatus: 'pending',
  storageProvider: 'local',
  uploadedAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

const processingDoc: DocumentDetail = { ...pendingDoc, analysisStatus: 'processing' }

const analysedDoc: DocumentDetail = {
  ...pendingDoc,
  analysisStatus: 'analysed',
  billRecord: {
    id: 'bill-1',
    dueDate: '2024-02-15',
    amountDue: '1500.00',
    paymentStatus: 'unpaid',
  },
}

const failedDoc: DocumentDetail = {
  ...pendingDoc,
  analysisStatus: 'analysis_failed',
}

// ─── tests ────────────────────────────────────────────────────────────────────

describe('useDocumentStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given a document in analysed state, when hook fetches, then BillRecord is exposed', () => {
    it('returns document with billRecord after successful analysis', async () => {
      // Arrange
      mockGetDocument.mockResolvedValueOnce(analysedDoc)

      // Act
      const { result } = renderHook(
        () => useDocumentStatus('doc-uuid-1'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.document?.analysisStatus).toBe('analysed')
      expect(result.current.document?.billRecord?.id).toBe('bill-1')
      expect(result.current.document?.billRecord?.dueDate).toBe('2024-02-15')
      expect(mockGetDocument).toHaveBeenCalledWith('doc-uuid-1')
    })
  })

  describe('given a document in analysis_failed state, when hook fetches, then error state is exposed', () => {
    it('returns document with analysis_failed status and no extraction records', async () => {
      // Arrange
      mockGetDocument.mockResolvedValueOnce(failedDoc)

      // Act
      const { result } = renderHook(
        () => useDocumentStatus('doc-uuid-1'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.document?.analysisStatus).toBe('analysis_failed')
      expect(result.current.document?.billRecord).toBeUndefined()
      expect(result.current.document?.statementRecord).toBeUndefined()
    })
  })

  describe('given a pending document then analysed document, when polling, then hook transitions status', () => {
    it('returns updated data as document transitions from pending to analysed', async () => {
      // Arrange: first call returns pending, second returns analysed
      mockGetDocument
        .mockResolvedValueOnce(pendingDoc)
        .mockResolvedValueOnce(analysedDoc)

      // Act
      const { result } = renderHook(
        () => useDocumentStatus('doc-uuid-1'),
        { wrapper: makeWrapper() },
      )

      // Assert: initial state is pending
      await waitFor(() => expect(result.current.document?.analysisStatus).toBe('pending'))

      // Assert: after the refetch fires, it transitions to analysed
      result.current.refetch()
      await waitFor(() => expect(result.current.document?.analysisStatus).toBe('analysed'))
      expect(result.current.document?.billRecord).toBeDefined()
    })
  })

  describe('given a processing document, when hook fetches, then polling is active', () => {
    // Note: this test validates the initial data state for a processing document;
    // the polling interval itself is driven by refetchInterval which Vitest can
    // verify through the mock call count after a manual refetch.
    it('returns processing status without extraction records', async () => {
      // Arrange
      mockGetDocument.mockResolvedValueOnce(processingDoc)

      // Act
      const { result } = renderHook(
        () => useDocumentStatus('doc-uuid-1'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.document?.analysisStatus).toBe('processing')
      expect(result.current.isError).toBe(false)
    })
  })

  describe('given a disabled query, when documentId is empty, then no API call is made', () => {
    it('does not fetch when documentId is empty', () => {
      // Arrange
      const { result } = renderHook(
        () => useDocumentStatus(''),
        { wrapper: makeWrapper() },
      )

      // Assert: query is disabled, no loading and no API call
      expect(result.current.isLoading).toBe(false)
      expect(mockGetDocument).not.toHaveBeenCalled()
    })
  })

  describe('given an API server error, when hook fetches, then error state is returned', () => {
    it('transitions to error state when the API rejects', async () => {
      // Arrange
      const apiError = new Error('Get document failed: 404')
      mockGetDocument.mockRejectedValueOnce(apiError)

      // Act
      const { result } = renderHook(
        () => useDocumentStatus('doc-uuid-1'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toContain('404')
      expect(result.current.document).toBeUndefined()
    })
  })
})
