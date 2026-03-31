import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getPaymentDashboard,
  markBillPaid,
  getPreferredDay,
  setPreferredDay,
} from '@/services/paymentsApi'
import type {
  PaymentDashboardResponse,
  MarkBillPaidResponse,
  CyclePreference,
  SetPreferredDayRequest,
} from '@/types/payments'

// ─── Query keys ───────────────────────────────────────────────────────────────

export const paymentQueryKeys = {
  all: ['payments'] as const,
  dashboard: (cycleStart?: string, cycleEnd?: string) =>
    ['payments', 'dashboard', cycleStart, cycleEnd] as const,
  preferredDay: () => ['payments', 'preferred-day'] as const,
}

// ─── Hooks ────────────────────────────────────────────────────────────────────

/**
 * usePaymentDashboard — fetches outstanding and overdue bills for a cycle range.
 *
 * Usage:
 *   const { data, isPending } = usePaymentDashboard('2024-01-01', '2024-01-31')
 */
export function usePaymentDashboard(cycleStart?: string, cycleEnd?: string) {
  return useQuery<PaymentDashboardResponse, Error>({
    queryKey: paymentQueryKeys.dashboard(cycleStart, cycleEnd),
    queryFn: () => getPaymentDashboard(cycleStart, cycleEnd),
  })
}

/**
 * useMarkBillPaid — mutation that idempotently marks a bill as paid.
 * Invalidates the dashboard query on success.
 *
 * Usage:
 *   const { mutate, isPending } = useMarkBillPaid()
 *   mutate('bill-uuid')
 */
export function useMarkBillPaid(
  options: {
    onSuccess?: (data: MarkBillPaidResponse) => void
    onError?: (err: Error) => void
  } = {},
) {
  const queryClient = useQueryClient()

  return useMutation<MarkBillPaidResponse, Error, string>({
    mutationFn: (billId: string) => markBillPaid(billId),
    onSuccess: (data) => {
      void queryClient.invalidateQueries({ queryKey: paymentQueryKeys.all })
      options.onSuccess?.(data)
    },
    onError: (err) => {
      options.onError?.(err)
    },
  })
}

/**
 * usePreferredDay — fetches the project's preferred payment day of month.
 *
 * Usage:
 *   const { data } = usePreferredDay()
 */
export function usePreferredDay() {
  return useQuery<CyclePreference, Error>({
    queryKey: paymentQueryKeys.preferredDay(),
    queryFn: getPreferredDay,
  })
}

/**
 * useSetPreferredDay — mutation that creates or updates the preferred payment day.
 * Invalidates the preferred-day query on success.
 *
 * Usage:
 *   const { mutate } = useSetPreferredDay()
 *   mutate({ preferredDayOfMonth: 10 })
 */
export function useSetPreferredDay(
  options: {
    onSuccess?: (data: CyclePreference) => void
    onError?: (err: Error) => void
  } = {},
) {
  const queryClient = useQueryClient()

  return useMutation<CyclePreference, Error, SetPreferredDayRequest>({
    mutationFn: (data: SetPreferredDayRequest) => setPreferredDay(data),
    onSuccess: (data) => {
      void queryClient.invalidateQueries({ queryKey: paymentQueryKeys.preferredDay() })
      options.onSuccess?.(data)
    },
    onError: (err) => {
      options.onError?.(err)
    },
  })
}
