import { beforeEach, describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement } from 'react'

import {
  useHistoryCategories,
  useHistoryCompliance,
  useHistoryDashboard,
  useHistoryTimeline,
} from './useHistoryDashboard'

import * as historyApi from '@/services/historyApi'
import type {
  CategoriesResponse,
  ComplianceResponse,
  TimelineResponse,
} from '@/types/history'

vi.mock('@/services/historyApi', () => ({
  getHistoryTimeline: vi.fn(),
  getHistoryCategories: vi.fn(),
  getHistoryCompliance: vi.fn(),
}))

const mockGetHistoryTimeline = vi.mocked(historyApi.getHistoryTimeline)
const mockGetHistoryCategories = vi.mocked(historyApi.getHistoryCategories)
const mockGetHistoryCompliance = vi.mocked(historyApi.getHistoryCompliance)

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })

  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const timelineFixture: TimelineResponse = {
  projectId: 'proj-1',
  months: 12,
  timeline: [{ month: '2026-03-01', totalAmount: '1200.00', billCount: 5 }],
}

const categoriesFixture: CategoriesResponse = {
  projectId: 'proj-1',
  months: 12,
  categories: [
    {
      month: '2026-03-01',
      billTypeName: 'Energy',
      totalAmount: '400.00',
      billCount: 2,
    },
  ],
}

const complianceFixture: ComplianceResponse = {
  projectId: 'proj-1',
  months: 12,
  compliance: [
    {
      month: '2026-03-01',
      totalBills: 5,
      paidOnTime: 4,
      overdue: 1,
      complianceRate: '80.00',
    },
  ],
}

describe('useHistoryDashboard hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('given months filter when loading timeline then forwards filter to API', async () => {
    mockGetHistoryTimeline.mockResolvedValueOnce(timelineFixture)

    const { result } = renderHook(() => useHistoryTimeline(6), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockGetHistoryTimeline).toHaveBeenCalledWith(6)
    expect(result.current.data?.timeline[0].billCount).toBe(5)
  })

  it('given months filter when loading categories then forwards filter to API', async () => {
    mockGetHistoryCategories.mockResolvedValueOnce(categoriesFixture)

    const { result } = renderHook(() => useHistoryCategories(3), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockGetHistoryCategories).toHaveBeenCalledWith(3)
    expect(result.current.data?.categories[0].billTypeName).toBe('Energy')
  })

  it('given months filter when loading compliance then forwards filter to API', async () => {
    mockGetHistoryCompliance.mockResolvedValueOnce(complianceFixture)

    const { result } = renderHook(() => useHistoryCompliance(9), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockGetHistoryCompliance).toHaveBeenCalledWith(9)
    expect(result.current.data?.compliance[0].complianceRate).toBe('80.00')
  })

  it('given dashboard hook when all queries succeed then exposes aggregated success state', async () => {
    mockGetHistoryTimeline.mockResolvedValueOnce(timelineFixture)
    mockGetHistoryCategories.mockResolvedValueOnce(categoriesFixture)
    mockGetHistoryCompliance.mockResolvedValueOnce(complianceFixture)

    const { result } = renderHook(() => useHistoryDashboard(12), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.isError).toBe(false)
    expect(result.current.timelineQuery.data?.timeline.length).toBe(1)
    expect(result.current.categoriesQuery.data?.categories.length).toBe(1)
    expect(result.current.complianceQuery.data?.compliance.length).toBe(1)
  })

  it('given one failing query when dashboard hook resolves then exposes aggregated error state', async () => {
    mockGetHistoryTimeline.mockRejectedValueOnce(new Error('timeline failed'))
    mockGetHistoryCategories.mockResolvedValueOnce(categoriesFixture)
    mockGetHistoryCompliance.mockResolvedValueOnce(complianceFixture)

    const { result } = renderHook(() => useHistoryDashboard(12), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.timelineQuery.error?.message).toBe('timeline failed')
  })
})
