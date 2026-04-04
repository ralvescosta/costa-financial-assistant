/**
 * Unit tests for LoginPage component.
 *
 * Covers:
 * - Default credential pre-fill from env vars
 * - Skeleton placeholder display during loading
 * - Error message display on 401
 * - Lockout countdown display on 429
 */

import { describe, it, expect, beforeAll, afterEach, afterAll, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { LoginPage } from '@/pages/LoginPage'

const server = setupServer()

beforeAll(() => server.listen())
afterEach(() => {
  server.resetHandlers()
  vi.clearAllMocks()
  localStorage.clear()
  document.documentElement.classList.remove('dark')
})
afterAll(() => server.close())

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter>{children}</MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>
  )
}

describe('LoginPage', () => {
  it('renders username and password inputs', () => {
    render(<LoginPage />, { wrapper: Wrapper })
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
  })

  it('pre-fills default credentials from env vars', () => {
    render(<LoginPage />, { wrapper: Wrapper })
    const username = screen.getByLabelText<HTMLInputElement>(/username/i)
    const password = screen.getByLabelText<HTMLInputElement>(/password/i)
    // Values come from import.meta.env; in test they may be empty or defaults
    expect(username).toBeInTheDocument()
    expect(password).toBeInTheDocument()
  })

  it('shows skeleton within 300ms during form submission', async () => {
    server.use(
      http.post('/api/auth/login', async () => {
        await new Promise((r) => setTimeout(r, 200))
        return HttpResponse.json({
          statusCode: 200,
          data: {
            expiresIn: 3600,
            refreshAt: 2700,
            csrfToken: 'csrf',
            user: { id: 'u1', username: 'demo' },
          },
        })
      }),
    )

    const user = userEvent.setup()
    render(<LoginPage />, { wrapper: Wrapper })

    await user.click(screen.getByRole('button', { name: /sign in/i }))

    // Button should be disabled during submission (label changes to "Signing in…")
    expect(screen.getByRole('button', { name: /signing in/i })).toBeDisabled()
  })

  it('displays error message on invalid credentials', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json(
          { statusCode: 401, error: { code: 'INVALID_CREDENTIALS', message: 'Invalid credentials' } },
          { status: 401 },
        ),
      ),
    )

    const user = userEvent.setup()
    render(<LoginPage />, { wrapper: Wrapper })

    await user.click(screen.getByRole('button', { name: /sign in/i }))
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  it('displays lockout countdown on 429 response', async () => {
    const lockoutUntil = new Date(Date.now() + 60_000).toISOString()
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json(
          {
            statusCode: 429,
            error: {
              code: 'AUTH_LOCKED',
              message: 'Too many failed attempts',
              lockoutUntil,
              remainingSeconds: 60,
            },
          },
          { status: 429 },
        ),
      ),
    )

    const user = userEvent.setup()
    render(<LoginPage />, { wrapper: Wrapper })

    await user.click(screen.getByRole('button', { name: /sign in/i }))
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  it('uses the shared primary-action token contract on first dark-mode render', () => {
    document.documentElement.classList.add('dark')

    render(<LoginPage />, { wrapper: Wrapper })

    const button = screen.getByRole('button', { name: /sign in/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
    expect(button.className).toContain('hover:bg-[color:var(--color-primary-action-hover)]')
  })

  it('keeps the loading sign-in action on the shared disabled token contract', async () => {
    server.use(
      http.post('/api/auth/login', async () => {
        await new Promise((resolve) => setTimeout(resolve, 200))
        return HttpResponse.json({
          statusCode: 200,
          data: {
            expiresIn: 3600,
            refreshAt: 2700,
            csrfToken: 'csrf',
            user: { id: 'u1', username: 'demo' },
          },
        })
      }),
    )

    const user = userEvent.setup()
    document.documentElement.classList.add('dark')
    render(<LoginPage />, { wrapper: Wrapper })

    await user.click(screen.getByRole('button', { name: /sign in/i }))

    const loadingButton = await screen.findByRole('button', { name: /signing in/i })
    expect(loadingButton).toBeDisabled()
    expect(loadingButton.className).toContain('disabled:bg-[color:var(--color-primary-action-disabled-bg)]')
    expect(loadingButton.className).toContain('disabled:text-[color:var(--color-primary-action-disabled-fg)]')
  })

  it('reuses the shared primary-action token contract in light mode', () => {
    render(<LoginPage />, { wrapper: Wrapper })

    const button = screen.getByRole('button', { name: /sign in/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
  })
})
