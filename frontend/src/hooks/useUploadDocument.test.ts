import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useUploadDocument } from './useUploadDocument'
import type { UploadDocumentResponse } from '@/types/documents'
import { createElement } from 'react'
import * as documentsApi from '@/services/documentsApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('@/services/documentsApi', () => ({
  uploadDocument: vi.fn(),
  classifyDocument: vi.fn(),
  listDocuments: vi.fn(),
}))

const mockUpload = vi.mocked(documentsApi.uploadDocument)
const mockClassify = vi.mocked(documentsApi.classifyDocument)

// ─── helpers ──────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

function makePdfFile(name = 'test.pdf') {
  return new File(['%PDF-1.4'], name, { type: 'application/pdf' })
}

const uploadedDoc: UploadDocumentResponse = {
  id: 'doc-uuid-1',
  projectId: 'proj-1',
  uploadedBy: 'user-1',
  kind: 'unspecified',
  fileName: 'test.pdf',
  analysisStatus: 'pending',
  storageProvider: 'local',
  uploadedAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

// ─── tests ────────────────────────────────────────────────────────────────────

describe('useUploadDocument', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given a valid PDF file, when upload is called, then the document is returned', () => {
    it('resolves with the uploaded document on success', async () => {
      // Arrange
      mockUpload.mockResolvedValueOnce(uploadedDoc)
      const { result } = renderHook(() => useUploadDocument(), {
        wrapper: makeWrapper(),
      })
      const file = makePdfFile()

      // Act
      result.current.mutate(file)

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.id).toBe('doc-uuid-1')
      expect(result.current.data?.analysisStatus).toBe('pending')
      expect(mockUpload).toHaveBeenCalledWith(file)
    })
  })

  describe('given a PDF and a kind option, when upload is called, then classify is also invoked', () => {
    it('calls classify after upload when kind is provided', async () => {
      // Arrange
      mockUpload.mockResolvedValueOnce(uploadedDoc)
      mockClassify.mockResolvedValueOnce({ ...uploadedDoc, kind: 'bill' })
      const { result } = renderHook(
        () => useUploadDocument({ kind: 'bill' }),
        { wrapper: makeWrapper() },
      )
      const file = makePdfFile()

      // Act
      result.current.mutate(file)

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(mockClassify).toHaveBeenCalledWith('doc-uuid-1', 'bill')
    })
  })

  describe('given a server error, when upload is called, then the error is surfaced', () => {
    it('transitions to error state when the API rejects', async () => {
      // Arrange
      const uploadError = new Error('Upload failed: 500')
      mockUpload.mockRejectedValueOnce(uploadError)
      const onError = vi.fn()
      const { result } = renderHook(
        () => useUploadDocument({ onError }),
        { wrapper: makeWrapper() },
      )
      const file = makePdfFile()

      // Act
      result.current.mutate(file)

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toContain('Upload failed')
      expect(onError).toHaveBeenCalledOnce()
    })
  })

  describe('given kind is unspecified, when upload is called, then classify is NOT invoked', () => {
    it('skips classify step when kind is unspecified', async () => {
      // Arrange
      mockUpload.mockResolvedValueOnce(uploadedDoc)
      const { result } = renderHook(
        () => useUploadDocument({ kind: 'unspecified' }),
        { wrapper: makeWrapper() },
      )
      const file = makePdfFile()

      // Act
      result.current.mutate(file)

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(mockClassify).not.toHaveBeenCalled()
    })
  })

  describe('given a duplicate document, when upload is called, then a conflict error is surfaced', () => {
    it('transitions to error state on 409 Conflict', async () => {
      // Arrange
      mockUpload.mockRejectedValueOnce(new Error('Conflict: document already uploaded'))
      const { result } = renderHook(() => useUploadDocument(), {
        wrapper: makeWrapper(),
      })
      const file = makePdfFile()

      // Act
      result.current.mutate(file)

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toContain('Conflict')
    })
  })
})

