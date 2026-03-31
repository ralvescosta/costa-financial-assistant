import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement } from 'react'
import {
  useReconciliationSummary,
  useCreateReconciliationLink,
} from './useReconciliation'
import type {
  ReconciliationSummary,
  ReconciliationLink,
  ReconciliationEntry,
} from '../types/reconciliation'
import * as reconciliationApi from '../services/reconciliationApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('../services/reconciliationApi', () => ({
  getReconciliationSummary: vi.fn(),
  createReconciliationLink: vi.fn(),
}))

const mockGetSummary = vi.mocked(reconciliationApi.getReconciliationSummary)
const mockCreateLink = vi.mocked(reconciliationApi.createReconciliationLink)

// ─── Helpers ──────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const unmatchedSummary: ReconciliationSummary = {
  projectId: 'proj-uuid-1',
  entries: [
    {
      transactionLineId: 'tx-uuid-1',
      transactionDate: '2024-01-15',
      description: 'Electric company',
      amount: '150.00',
      direction: 'debit',
      reconciliationStatus: 'unmatched',
    },
  ],
}

const mixedSummary: ReconciliationSummary = {
  projectId: 'proj-uuid-1',
  periodStart: '2024-01-01',
  periodEnd: '2024-01-31',
  entries: [
    {
      transactionLineId: 'tx-uuid-2',
      transactionDate: '2024-01-10',
      description: 'Internet bill',
      amount: '89.90',
      direction: 'debit',
      reconciliationStatus: 'matched_auto',
      linkedBillId: 'bill-uuid-1',
      linkedBillDueDate: '2024-01-12',
      linkedBillAmount: '89.90',
      linkType: 'auto',
    },
    {
      transactionLineId: 'tx-uuid-3',
      transactionDate: '2024-01-20',
      description: 'Unknown debit',
      amount: '200.00',
      direction: 'debit',
      reconciliationStatus: 'ambiguous',
    },
  ],
}

const createdLink: ReconciliationLink = {
  id: 'link-uuid-1',
  projectId: 'proj-uuid-1',
  transactionLineId: 'tx-uuid-1',
  billRecordId: 'bill-uuid-1',
  linkType: 'manual',
  linkedBy: 'user-uuid-1',
  createdAt: '2024-01-25T12:00:00Z',
}

// ─── Tests ────────────────────────────────────────────────────────────────────

describe('useReconciliationSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given no period filter, when the hook fetches, then it returns all entries', () => {
    it('resolves with reconciliation entries on success', async () => {
      // Arrange
      mockGetSummary.mockResolvedValueOnce(unmatchedSummary)
      const { result } = renderHook(() => useReconciliationSummary(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.entries).toHaveLength(1)
      expect(result.current.data?.entries[0].reconciliationStatus).toBe('unmatched')
      expect(mockGetSummary).toHaveBeenCalledWith(undefined, undefined)
    })
  })

  describe('given period date filters, when the hook fetches, then it passes them to the API', () => {
    it('forwards periodStart and periodEnd to getReconciliationSummary', async () => {
      // Arrange
      mockGetSummary.mockResolvedValueOnce(mixedSummary)
      const { result } = renderHook(
        () => useReconciliationSummary('2024-01-01', '2024-01-31'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(mockGetSummary).toHaveBeenCalledWith('2024-01-01', '2024-01-31')
      expect(result.current.data?.entries).toHaveLength(2)
    })
  })

  describe('given a matched entry, when the hook fetches, then linked bill details are present', () => {
    it('returns linkedBillId and linkType for matched entries', async () => {
      // Arrange
      mockGetSummary.mockResolvedValueOnce(mixedSummary)
      const { result } = renderHook(
        () => useReconciliationSummary('2024-01-01', '2024-01-31'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      const matchedEntry = result.current.data!.entries.find(
        (e: ReconciliationEntry) => e.reconciliationStatus === 'matched_auto',
      )
      expect(matchedEntry?.linkedBillId).toBe('bill-uuid-1')
      expect(matchedEntry?.linkType).toBe('auto')
    })
  })

  describe('given an API error, when the hook fetches, then it exposes the error', () => {
    it('sets isError on fetch failure', async () => {
      // Arrange
      mockGetSummary.mockRejectedValueOnce(new Error('server unavailable'))
      const { result } = renderHook(() => useReconciliationSummary(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toBe('server unavailable')
    })
  })
})

describe('useCreateReconciliationLink', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given valid transaction and bill IDs, when mutate is called, then the link is created', () => {
    it('resolves with the created reconciliation link', async () => {
      // Arrange
      mockCreateLink.mockResolvedValueOnce(createdLink)
      const { result } = renderHook(() => useCreateReconciliationLink(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ transactionLineId: 'tx-uuid-1', billRecordId: 'bill-uuid-1' })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.id).toBe('link-uuid-1')
      expect(result.current.data?.linkType).toBe('manual')
      expect(mockCreateLink).toHaveBeenCalledWith({
        transactionLineId: 'tx-uuid-1',
        billRecordId: 'bill-uuid-1',
      })
    })
  })

  describe('given an onSuccess callback, when mutate succeeds, then the callback is invoked', () => {
    it('calls the onSuccess option with response data', async () => {
      // Arrange
      mockCreateLink.mockResolvedValueOnce(createdLink)
      const onSuccess = vi.fn()
      const { result } = renderHook(() => useCreateReconciliationLink({ onSuccess }), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ transactionLineId: 'tx-uuid-1', billRecordId: 'bill-uuid-1' })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(onSuccess).toHaveBeenCalledWith(createdLink)
    })
  })

  describe('given an API error, when mutate is called, then it exposes the error', () => {
    it('sets isError on mutation failure', async () => {
      // Arrange
      mockCreateLink.mockRejectedValueOnce(new Error('conflict'))
      const { result } = renderHook(() => useCreateReconciliationLink(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ transactionLineId: 'tx-uuid-1', billRecordId: 'bill-uuid-1' })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toBe('conflict')
    })
  })
})
