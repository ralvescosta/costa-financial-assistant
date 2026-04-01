/**
 * Zod schemas for BFF lockout error responses and countdown calculations.
 */

import { z } from 'zod'

export const LockoutErrorSchema = z.object({
  code: z.literal('AUTH_LOCKED'),
  message: z.string(),
  lockoutUntil: z.string().datetime(),
  remainingSeconds: z.number().int().min(1),
})

export const LockoutErrorResponseSchema = z.object({
  statusCode: z.literal(429),
  error: LockoutErrorSchema,
})

export const AuthErrorSchema = z.object({
  code: z.enum(['INVALID_CREDENTIALS', 'SESSION_EXPIRED']),
  message: z.string(),
})

export const AuthErrorResponseSchema = z.object({
  statusCode: z.literal(401),
  error: AuthErrorSchema,
})

/**
 * Given a lockout ISO timestamp, calculates the remaining seconds until
 * the lockout expires. Returns 0 if the lockout has already expired.
 */
export function calcLockoutRemainingSeconds(lockoutUntil: string): number {
  const expiryMs = new Date(lockoutUntil).getTime()
  const nowMs = Date.now()
  return Math.max(0, Math.ceil((expiryMs - nowMs) / 1000))
}

export type LockoutError = z.infer<typeof LockoutErrorSchema>
export type LockoutErrorResponse = z.infer<typeof LockoutErrorResponseSchema>
export type AuthError = z.infer<typeof AuthErrorSchema>
export type AuthErrorResponse = z.infer<typeof AuthErrorResponseSchema>
