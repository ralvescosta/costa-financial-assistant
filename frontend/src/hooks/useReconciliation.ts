import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getReconciliationSummary,
  createReconciliationLink,
} from '../services/reconciliationApi'
import type {
  ReconciliationSummary,
  ReconciliationLink,
  CreateReconciliationLinkRequest,
} from '../types/reconciliation'

// ─── Query keys ───────────────────────────────────────────────────────────────

export const reconciliationQueryKeys = {
  all: ['reconciliation'] as const,
  summary: (periodStart?: string, periodEnd?: string) =>
    ['reconciliation', 'summary', periodStart, periodEnd] as const,
}

// ─── Hooks ────────────────────────────────────────────────────────────────────

/**
 * useReconciliationSummary — fetches the transaction-to-bill reconciliation status.
 *
 * Usage:
 *   const { data, isPending } = useReconciliationSummary('2024-01-01', '2024-01-31')
 */
export function useReconciliationSummary(periodStart?: string, periodEnd?: string) {
  return useQuery<ReconciliationSummary, Error>({
    queryKey: reconciliationQueryKeys.summary(periodStart, periodEnd),
    queryFn: () => getReconciliationSummary(periodStart, periodEnd),
  })
}

/**
 * useCreateReconciliationLink — mutation that manually links a transaction to a bill.
 * Invalidates reconciliation queries on success.
 *
 * Usage:
 *   const { mutate } = useCreateReconciliationLink()
 *   mutate({ transactionLineId: '...', billRecordId: '...' })
 */
export function useCreateReconciliationLink(
  options: {
    onSuccess?: (data: ReconciliationLink) => void
    onError?: (err: Error) => void
  } = {},
) {
  const queryClient = useQueryClient()

  return useMutation<ReconciliationLink, Error, CreateReconciliationLinkRequest>({
    mutationFn: (data: CreateReconciliationLinkRequest) => createReconciliationLink(data),
    onSuccess: (data) => {
      void queryClient.invalidateQueries({ queryKey: reconciliationQueryKeys.all })
      options.onSuccess?.(data)
    },
    onError: (err) => {
      options.onError?.(err)
    },
  })
}
