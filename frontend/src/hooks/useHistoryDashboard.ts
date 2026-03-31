import { useQuery } from '@tanstack/react-query'
import {
  getHistoryCategories,
  getHistoryCompliance,
  getHistoryTimeline,
} from '@/services/historyApi'
import type {
  CategoriesResponse,
  ComplianceResponse,
  TimelineResponse,
} from '@/types/history'

export const historyQueryKeys = {
  all: ['history'] as const,
  timeline: (months: number) => ['history', 'timeline', months] as const,
  categories: (months: number) => ['history', 'categories', months] as const,
  compliance: (months: number) => ['history', 'compliance', months] as const,
}

export function useHistoryTimeline(months = 12) {
  return useQuery<TimelineResponse, Error>({
    queryKey: historyQueryKeys.timeline(months),
    queryFn: () => getHistoryTimeline(months),
  })
}

export function useHistoryCategories(months = 12) {
  return useQuery<CategoriesResponse, Error>({
    queryKey: historyQueryKeys.categories(months),
    queryFn: () => getHistoryCategories(months),
  })
}

export function useHistoryCompliance(months = 12) {
  return useQuery<ComplianceResponse, Error>({
    queryKey: historyQueryKeys.compliance(months),
    queryFn: () => getHistoryCompliance(months),
  })
}

export function useHistoryDashboard(months = 12) {
  const timelineQuery = useHistoryTimeline(months)
  const categoriesQuery = useHistoryCategories(months)
  const complianceQuery = useHistoryCompliance(months)

  return {
    months,
    timelineQuery,
    categoriesQuery,
    complianceQuery,
    isPending:
      timelineQuery.isPending || categoriesQuery.isPending || complianceQuery.isPending,
    isError: timelineQuery.isError || categoriesQuery.isError || complianceQuery.isError,
  }
}
