/**
 * useDraftStateRestore — short-lived draft state persistence with TTL enforcement.
 *
 * Saves serialised draft data to localStorage under a namespaced key and
 * enforces a TTL (seconds). Once restored, the entry is marked `used: true`
 * so it cannot be consumed a second time.
 *
 * Auth tokens are NEVER stored — only the user's UI work-in-progress state.
 */

import { useCallback } from 'react'
import { DraftRestoreDataSchema, type DraftRestoreData } from '@/types/session.schema'

const DRAFT_PREFIX = 'cfa:draft:'

function storageKey(key: string): string {
  return `${DRAFT_PREFIX}${key}`
}

export interface UseDraftStateRestore {
  /** Persist draft data under `key` with a TTL (seconds, default 30 min). */
  saveDraftState: (key: string, data: unknown, ttlSeconds?: number) => void
  /**
   * Retrieve draft data for `key` if it exists, has not expired, and has not
   * been used before. Returns null otherwise and clears the stale entry.
   */
  getDraftState: (key: string) => unknown | null
  /** Remove the draft entry for `key`. */
  clearDraftState: (key: string) => void
}

const DEFAULT_TTL_SECONDS = 30 * 60 // 30 minutes

export function useDraftStateRestore(): UseDraftStateRestore {
  const saveDraftState = useCallback(
    (key: string, data: unknown, ttlSeconds = DEFAULT_TTL_SECONDS) => {
      const entry: DraftRestoreData = {
        key,
        data,
        savedAt: Math.floor(Date.now() / 1000),
        ttl: ttlSeconds,
        used: false,
      }
      localStorage.setItem(storageKey(key), JSON.stringify(entry))
    },
    [],
  )

  const getDraftState = useCallback((key: string): unknown | null => {
    try {
      const raw = localStorage.getItem(storageKey(key))
      if (!raw) return null

      const parsed = DraftRestoreDataSchema.safeParse(JSON.parse(raw))
      if (!parsed.success) {
        localStorage.removeItem(storageKey(key))
        return null
      }

      const entry = parsed.data
      const nowSec = Math.floor(Date.now() / 1000)

      // Expired
      if (entry.savedAt + entry.ttl < nowSec) {
        localStorage.removeItem(storageKey(key))
        return null
      }

      // Already restored
      if (entry.used) {
        localStorage.removeItem(storageKey(key))
        return null
      }

      // Mark as used — one-time restore
      const updated: DraftRestoreData = { ...entry, used: true }
      localStorage.setItem(storageKey(key), JSON.stringify(updated))

      return entry.data
    } catch {
      return null
    }
  }, [])

  const clearDraftState = useCallback((key: string) => {
    localStorage.removeItem(storageKey(key))
  }, [])

  return { saveDraftState, getDraftState, clearDraftState }
}
