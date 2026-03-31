import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement } from 'react'
import {
  usePaymentDashboard,
  useMarkBillPaid,
  usePreferredDay,
  useSetPreferredDay,
} from './usePaymentDashboard'
import type {
  PaymentDashboardResponse,
  MarkBillPaidResponse,
  CyclePreference,
} from '@/types/payments'
import * as paymentsApi from '@/services/paymentsApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('@/services/paymentsApi', () => ({
  getPaymentDashboard: vi.fn(),
  markBillPaid: vi.fn(),
  getPreferredDay: vi.fn(),
  setPreferredDay: vi.fn(),
}))

const mockGetDashboard = vi.mocked(paymentsApi.getPaymentDashboard)
const mockMarkBillPaid = vi.mocked(paymentsApi.markBillPaid)
const mockGetPreferredDay = vi.mocked(paymentsApi.getPreferredDay)
const mockSetPreferredDay = vi.mocked(paymentsApi.setPreferredDay)

// ─── Helpers ──────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const unpaidBill: PaymentDashboardResponse = {
  entries: [
    {
      bill: {
        id: 'bill-uuid-1',
        projectId: 'proj-uuid-1',
        documentId: 'doc-uuid-1',
        dueDate: '2024-02-10',
        amountDue: '150.00',
        paymentStatus: 'PAYMENT_STATUS_UNPAID',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
      billType: { id: 'bt-1', projectId: 'proj-uuid-1', name: 'Electricity' },
      isOverdue: false,
      daysUntilDue: 5,
    },
  ],
  nextPageToken: undefined,
}

const overdueDashboard: PaymentDashboardResponse = {
  entries: [
    {
      bill: {
        id: 'bill-uuid-2',
        projectId: 'proj-uuid-1',
        documentId: 'doc-uuid-2',
        dueDate: '2024-01-01',
        amountDue: '200.00',
        paymentStatus: 'PAYMENT_STATUS_OVERDUE',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
      isOverdue: true,
      daysUntilDue: -3,
    },
  ],
}

const markedPaidResponse: MarkBillPaidResponse = {
  bill: {
    id: 'bill-uuid-1',
    projectId: 'proj-uuid-1',
    documentId: 'doc-uuid-1',
    dueDate: '2024-02-10',
    amountDue: '150.00',
    paymentStatus: 'PAYMENT_STATUS_PAID',
    paidAt: '2024-01-25T10:00:00Z',
    markedPaidBy: 'user-uuid-1',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-25T10:00:00Z',
  },
}

const cyclePreference: CyclePreference = {
  projectId: 'proj-uuid-1',
  preferredDayOfMonth: 10,
  updatedAt: '2024-01-01T00:00:00Z',
}

// ─── Tests ────────────────────────────────────────────────────────────────────

describe('usePaymentDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given no cycle filter, when the hook fetches, then it returns dashboard entries', () => {
    it('resolves with bill entries on success', async () => {
      // Arrange
      mockGetDashboard.mockResolvedValueOnce(unpaidBill)
      const { result } = renderHook(() => usePaymentDashboard(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.entries).toHaveLength(1)
      expect(result.current.data?.entries[0].bill.id).toBe('bill-uuid-1')
      expect(result.current.data?.entries[0].isOverdue).toBe(false)
      expect(mockGetDashboard).toHaveBeenCalledWith(undefined, undefined)
    })
  })

  describe('given cycle date filters, when the hook fetches, then it passes them to the API', () => {
    it('forwards cycleStart and cycleEnd to getPaymentDashboard', async () => {
      // Arrange
      mockGetDashboard.mockResolvedValueOnce(unpaidBill)
      const { result } = renderHook(
        () => usePaymentDashboard('2024-02-01', '2024-02-28'),
        { wrapper: makeWrapper() },
      )

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(mockGetDashboard).toHaveBeenCalledWith('2024-02-01', '2024-02-28')
    })
  })

  describe('given an overdue bill, when the hook fetches, then isOverdue and daysUntilDue are negative', () => {
    it('returns overdue flag and negative daysUntilDue', async () => {
      // Arrange
      mockGetDashboard.mockResolvedValueOnce(overdueDashboard)
      const { result } = renderHook(() => usePaymentDashboard(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      const entry = result.current.data!.entries[0]
      expect(entry.isOverdue).toBe(true)
      expect(entry.daysUntilDue).toBeLessThan(0)
    })
  })

  describe('given an API error, when the hook fetches, then it exposes the error', () => {
    it('sets isError on fetch failure', async () => {
      // Arrange
      mockGetDashboard.mockRejectedValueOnce(new Error('fetch failed'))
      const { result } = renderHook(() => usePaymentDashboard(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toBe('fetch failed')
    })
  })
})

describe('useMarkBillPaid', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given an unpaid bill id, when mutate is called, then the bill is marked paid', () => {
    it('resolves with the paid bill record', async () => {
      // Arrange
      mockMarkBillPaid.mockResolvedValueOnce(markedPaidResponse)
      const { result } = renderHook(() => useMarkBillPaid(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate('bill-uuid-1')

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.bill.paymentStatus).toBe('PAYMENT_STATUS_PAID')
      expect(mockMarkBillPaid).toHaveBeenCalledWith('bill-uuid-1')
    })
  })

  describe('given an onSuccess callback, when mutate succeeds, then the callback is invoked', () => {
    it('calls the onSuccess option with response data', async () => {
      // Arrange
      mockMarkBillPaid.mockResolvedValueOnce(markedPaidResponse)
      const onSuccess = vi.fn()
      const { result } = renderHook(() => useMarkBillPaid({ onSuccess }), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate('bill-uuid-1')

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(onSuccess).toHaveBeenCalledWith(markedPaidResponse)
    })
  })

  describe('given an API error, when mutate is called, then it exposes the error', () => {
    it('sets isError on mutation failure', async () => {
      // Arrange
      mockMarkBillPaid.mockRejectedValueOnce(new Error('network error'))
      const { result } = renderHook(() => useMarkBillPaid(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate('invalid-id')

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toBe('network error')
    })
  })
})

describe('usePreferredDay', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given a project with a preferred day, when the hook fetches, then it returns the preference', () => {
    it('resolves with cycle preference on success', async () => {
      // Arrange
      mockGetPreferredDay.mockResolvedValueOnce(cyclePreference)
      const { result } = renderHook(() => usePreferredDay(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.preferredDayOfMonth).toBe(10)
      expect(mockGetPreferredDay).toHaveBeenCalledOnce()
    })
  })

  describe('given an API error, when the hook fetches, then it exposes the error', () => {
    it('sets isError on fetch failure', async () => {
      // Arrange
      mockGetPreferredDay.mockRejectedValueOnce(new Error('not found'))
      const { result } = renderHook(() => usePreferredDay(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toBe('not found')
    })
  })
})

describe('useSetPreferredDay', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given a preferred day of month, when mutate is called, then the preference is saved', () => {
    it('resolves with updated cycle preference', async () => {
      // Arrange
      mockSetPreferredDay.mockResolvedValueOnce(cyclePreference)
      const { result } = renderHook(() => useSetPreferredDay(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ preferredDayOfMonth: 10 })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(result.current.data?.preferredDayOfMonth).toBe(10)
      expect(mockSetPreferredDay).toHaveBeenCalledWith({ preferredDayOfMonth: 10 })
    })
  })

  describe('given an onSuccess callback, when mutate succeeds, then the callback is invoked', () => {
    it('calls the onSuccess option with response data', async () => {
      // Arrange
      mockSetPreferredDay.mockResolvedValueOnce(cyclePreference)
      const onSuccess = vi.fn()
      const { result } = renderHook(() => useSetPreferredDay({ onSuccess }), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ preferredDayOfMonth: 10 })

      // Assert
      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(onSuccess).toHaveBeenCalledWith(cyclePreference)
    })
  })

  describe('given an API error, when mutate is called, then it exposes the error', () => {
    it('sets isError on mutation failure', async () => {
      // Arrange
      mockSetPreferredDay.mockRejectedValueOnce(new Error('server error'))
      const { result } = renderHook(() => useSetPreferredDay(), {
        wrapper: makeWrapper(),
      })

      // Act
      result.current.mutate({ preferredDayOfMonth: 29 })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toBe('server error')
    })
  })
})
