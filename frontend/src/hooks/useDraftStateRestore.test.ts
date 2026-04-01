/**
 * Unit tests for useDraftStateRestore hook.
 *
 * Covers:
 * - Draft saved to localStorage with TTL
 * - Draft restored and marked as used (one-time restore)
 * - Expired draft returns null
 * - Used draft returns null
 * - clearDraftState removes entry
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useDraftStateRestore } from '@/hooks/useDraftStateRestore'

describe('useDraftStateRestore', () => {
  beforeEach(() => localStorage.clear())
  afterEach(() => {
    localStorage.clear()
    vi.useRealTimers()
  })

  it('saveDraftState writes an entry to localStorage', () => {
    const { result } = renderHook(() => useDraftStateRestore())
    act(() => result.current.saveDraftState('form-1', { title: 'Test' }, 1800))
    expect(localStorage.getItem('cfa:draft:form-1')).toBeTruthy()
  })

  it('getDraftState returns the saved data', () => {
    const { result } = renderHook(() => useDraftStateRestore())
    act(() => result.current.saveDraftState('form-2', { amount: 42 }, 1800))
    const data = result.current.getDraftState('form-2')
    expect(data).toEqual({ amount: 42 })
  })

  it('getDraftState returns null on second call (one-time restore)', () => {
    const { result } = renderHook(() => useDraftStateRestore())
    act(() => result.current.saveDraftState('form-3', { x: 1 }, 1800))
    result.current.getDraftState('form-3')
    const secondRead = result.current.getDraftState('form-3')
    expect(secondRead).toBeNull()
  })

  it('getDraftState returns null for expired entry', () => {
    vi.useFakeTimers()
    const { result } = renderHook(() => useDraftStateRestore())
    act(() => result.current.saveDraftState('form-4', { y: 2 }, 10)) // 10s TTL
    vi.advanceTimersByTime(11_000) // advance past TTL
    const data = result.current.getDraftState('form-4')
    expect(data).toBeNull()
  })

  it('clearDraftState removes the entry', () => {
    const { result } = renderHook(() => useDraftStateRestore())
    act(() => result.current.saveDraftState('form-5', { z: 3 }, 1800))
    act(() => result.current.clearDraftState('form-5'))
    expect(localStorage.getItem('cfa:draft:form-5')).toBeNull()
  })

  it('getDraftState returns null for unknown key', () => {
    const { result } = renderHook(() => useDraftStateRestore())
    expect(result.current.getDraftState('unknown-key')).toBeNull()
  })
})
