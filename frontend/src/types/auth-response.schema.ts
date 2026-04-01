/**
 * Zod schemas for BFF login and refresh API response validation.
 *
 * Validates the full response envelope before updating auth state
 * so malformed server responses fail loudly rather than silently.
 */

import { z } from 'zod'

export const UserSummarySchema = z.object({
  id: z.string(),
  username: z.string(),
  email: z.string().email().optional(),
})

export const ProjectSummarySchema = z.object({
  id: z.string(),
  name: z.string(),
  role: z.enum(['read_only', 'update', 'write']),
})

export const LoginSuccessDataSchema = z.object({
  expiresIn: z.number().int().min(60),
  refreshAt: z.number().int().min(1),
  csrfToken: z.string(),
  user: UserSummarySchema,
  activeProject: ProjectSummarySchema.optional(),
})

export const LoginSuccessResponseSchema = z.object({
  statusCode: z.literal(200),
  data: LoginSuccessDataSchema,
})

export const RefreshSuccessDataSchema = z.object({
  expiresIn: z.number().int().min(60),
  refreshAt: z.number().int().min(1),
  csrfToken: z.string(),
})

export const RefreshSuccessResponseSchema = z.object({
  statusCode: z.literal(200),
  data: RefreshSuccessDataSchema,
})

export type LoginSuccessData = z.infer<typeof LoginSuccessDataSchema>
export type LoginSuccessResponse = z.infer<typeof LoginSuccessResponseSchema>
export type RefreshSuccessData = z.infer<typeof RefreshSuccessDataSchema>
export type RefreshSuccessResponse = z.infer<typeof RefreshSuccessResponseSchema>
