/**
 * auth.config — centralized auth configuration sourced from environment variables.
 *
 * Default credentials are applicable ONLY in development mode. A production
 * safety check ensures these values are never present in production builds.
 */

function assertDev(varName: string, value: string | undefined): string | undefined {
  if (import.meta.env.PROD && value) {
    console.warn(
      `[auth.config] ${varName} is set in a production build. This is a misconfiguration. ` +
        'Default credentials MUST NOT be present in production.',
    )
    return undefined
  }
  return value
}

export const authConfig = {
  defaultUsername: assertDev(
    'VITE_DEFAULT_USERNAME',
    import.meta.env.VITE_DEFAULT_USERNAME as string | undefined,
  ),
  defaultPassword: assertDev(
    'VITE_DEFAULT_PASSWORD',
    import.meta.env.VITE_DEFAULT_PASSWORD as string | undefined,
  ),
} as const
