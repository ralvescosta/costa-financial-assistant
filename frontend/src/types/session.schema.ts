/**
 * Zod schemas for client-side session storage and draft restoration metadata.
 *
 * Auth tokens are transported exclusively via HTTP-only cookies (never
 * stored in localStorage). Only non-sensitive session metadata and
 * short-lived, TTL-bounded draft state are persisted in browser storage.
 */

import { z } from 'zod'

/** Non-sensitive auth session metadata persisted for page-reload recovery. */
export const SessionMetadataSchema = z.object({
  userId: z.string(),
  username: z.string(),
  expiryTimestamp: z.number().int(),
  refreshAtTimestamp: z.number().int(),
  activeProjectId: z.string().optional(),
})

/** Draft state entry stored with a TTL for one-time restoration after re-login. */
export const DraftRestoreDataSchema = z.object({
  key: z.string(),
  /** Serialised draft payload — consumers must parse their own schema */
  data: z.unknown(),
  savedAt: z.number().int(),
  ttl: z.number().int().min(1),
  /** True once the draft has been restored — prevents second-use */
  used: z.boolean(),
})

export type SessionMetadata = z.infer<typeof SessionMetadataSchema>
export type DraftRestoreData = z.infer<typeof DraftRestoreDataSchema>
