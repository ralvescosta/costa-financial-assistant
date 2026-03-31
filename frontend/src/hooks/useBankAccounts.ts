import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createBankAccount,
  deleteBankAccount,
  listBankAccounts,
} from '@/services/bankAccountsApi'
import type { BankAccount, CreateBankAccountRequest } from '@/types/bankAccounts'

const QUERY_KEY = ['bank-accounts'] as const

export interface UseBankAccountsResult {
  bankAccounts: BankAccount[]
  isLoading: boolean
  isError: boolean
  error: Error | null
  refetch: () => void
  create: (req: CreateBankAccountRequest) => Promise<BankAccount>
  remove: (bankAccountId: string) => Promise<void>
  isCreating: boolean
  isDeleting: boolean
}

/**
 * useBankAccounts — combined query + mutation hook for bank account CRUD.
 *
 * Usage:
 *   const { bankAccounts, create, remove, isLoading } = useBankAccounts()
 */
export function useBankAccounts(): UseBankAccountsResult {
  const queryClient = useQueryClient()

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: QUERY_KEY,
    queryFn: listBankAccounts,
  })

  const createMutation = useMutation({
    mutationFn: createBankAccount,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteBankAccount,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })

  return {
    bankAccounts: data?.items ?? [],
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
    create: (req) => createMutation.mutateAsync(req),
    remove: (id) => deleteMutation.mutateAsync(id),
    isCreating: createMutation.isPending,
    isDeleting: deleteMutation.isPending,
  }
}
