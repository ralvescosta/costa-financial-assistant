import { describe, it, expect, vi, beforeEach } from 'vitest'
import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement } from 'react'
import { useBankAccounts } from './useBankAccounts'
import type { BankAccount } from '@/types/bankAccounts'
import * as bankAccountsApi from '@/services/bankAccountsApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('@/services/bankAccountsApi', () => ({
  listBankAccounts: vi.fn(),
  createBankAccount: vi.fn(),
  deleteBankAccount: vi.fn(),
}))

const mockList = vi.mocked(bankAccountsApi.listBankAccounts)
const mockCreate = vi.mocked(bankAccountsApi.createBankAccount)
const mockDelete = vi.mocked(bankAccountsApi.deleteBankAccount)

// ─── helpers ──────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const checkingAccount: BankAccount = {
  id: 'ba-uuid-1',
  projectId: 'proj-1',
  label: 'Checking Account',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

const savingsAccount: BankAccount = {
  id: 'ba-uuid-2',
  projectId: 'proj-1',
  label: 'Savings Account',
  createdAt: '2024-01-02T00:00:00Z',
  updatedAt: '2024-01-02T00:00:00Z',
}

// ─── tests ────────────────────────────────────────────────────────────────────

describe('useBankAccounts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // ── List ─────────────────────────────────────────────────────────────────

  describe('given the API returns bank accounts, when the hook mounts, then accounts are loaded', () => {
    it('populates bankAccounts with the fetched items', async () => {
      // Arrange
      mockList.mockResolvedValueOnce({ items: [checkingAccount, savingsAccount] })

      // Act
      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.bankAccounts).toHaveLength(2)
      expect(result.current.bankAccounts[0].label).toBe('Checking Account')
      expect(result.current.bankAccounts[1].label).toBe('Savings Account')
      expect(mockList).toHaveBeenCalledOnce()
    })
  })

  describe('given the API returns an empty list, when the hook mounts, then an empty array is produced', () => {
    it('returns empty bankAccounts when API returns empty items', async () => {
      // Arrange
      mockList.mockResolvedValueOnce({ items: [] })

      // Act
      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.bankAccounts).toHaveLength(0)
    })
  })

  describe('given a list API error, when the hook mounts, then isError is true', () => {
    it('transitions to error state when list API rejects', async () => {
      // Arrange
      mockList.mockRejectedValueOnce(new Error('List bank accounts failed: 500'))

      // Act
      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isError).toBe(true))
      expect(result.current.error?.message).toContain('List bank accounts failed')
    })
  })

  // ── Create ────────────────────────────────────────────────────────────────

  describe('given a valid label, when create is called, then the new account is returned and list is refreshed', () => {
    it('resolves with the created bank account on success', async () => {
      // Arrange
      mockList.mockResolvedValue({ items: [checkingAccount] })
      mockCreate.mockResolvedValueOnce(checkingAccount)

      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })
      await waitFor(() => expect(result.current.isLoading).toBe(false))

      // Act
      let created: BankAccount | undefined
      await act(async () => {
        created = await result.current.create({ label: 'Checking Account' })
      })

      // Assert
      expect(created?.id).toBe('ba-uuid-1')
      expect(created?.label).toBe('Checking Account')
      expect(mockCreate.mock.calls[0][0]).toEqual({ label: 'Checking Account' })
    })
  })

  describe('given a duplicate label, when create is called, then the error is surfaced', () => {
    it('rejects when the API returns a conflict error', async () => {
      // Arrange
      mockList.mockResolvedValue({ items: [] })
      mockCreate.mockRejectedValueOnce(
        new Error('bank account label already exists in this project'),
      )

      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })
      await waitFor(() => expect(result.current.isLoading).toBe(false))

      // Act + Assert
      await expect(
        act(async () => {
          await result.current.create({ label: 'Checking Account' })
        }),
      ).rejects.toThrow('bank account label already exists')
    })
  })

  // ── Delete ────────────────────────────────────────────────────────────────

  describe('given an existing account ID, when remove is called, then the account is deleted and list is refreshed', () => {
    it('resolves without error on successful delete', async () => {
      // Arrange
      mockList.mockResolvedValue({ items: [checkingAccount] })
      mockDelete.mockResolvedValueOnce(undefined)

      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })
      await waitFor(() => expect(result.current.isLoading).toBe(false))

      // Act
      await act(async () => {
        await result.current.remove('ba-uuid-1')
      })

      // Assert
      expect(mockDelete.mock.calls[0][0]).toBe('ba-uuid-1')
    })
  })

  describe('given an account in use, when remove is called, then attribution guard error is surfaced', () => {
    it('rejects when the API returns a 409 error', async () => {
      // Arrange
      mockList.mockResolvedValue({ items: [checkingAccount] })
      mockDelete.mockRejectedValueOnce(
        new Error('bank account is referenced by statement records'),
      )

      const { result } = renderHook(() => useBankAccounts(), {
        wrapper: makeWrapper(),
      })
      await waitFor(() => expect(result.current.isLoading).toBe(false))

      // Act + Assert
      await expect(
        act(async () => {
          await result.current.remove('ba-uuid-1')
        }),
      ).rejects.toThrow('bank account is referenced by statement records')
    })
  })
})
