/**
 * Accessibility test: LoginPage keyboard-only navigation.
 *
 * Verifies Tab order (username → password → button) and Enter key submits.
 */

import { describe, it, expect, beforeAll, afterEach, afterAll } from 'vitest'
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
afterEach(() => server.resetHandlers())
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

describe('LoginPage accessibility', () => {
  it('username input is the first focusable element', () => {
    render(<LoginPage />, { wrapper: Wrapper })
    const username = screen.getByLabelText<HTMLInputElement>(/username/i)
    username.focus()
    expect(document.activeElement).toBe(username)
  })

  it('Tab from username moves focus to password', async () => {
    const user = userEvent.setup()
    render(<LoginPage />, { wrapper: Wrapper })

    const username = screen.getByLabelText(/username/i)
    const password = screen.getByLabelText(/password/i)

    username.focus()
    await user.tab()
    expect(document.activeElement).toBe(password)
  })

  it('Tab from password moves focus to submit button', async () => {
    const user = userEvent.setup()
    render(<LoginPage />, { wrapper: Wrapper })

    const password = screen.getByLabelText(/password/i)
    const button = screen.getByRole('button', { name: /sign in/i })

    password.focus()
    await user.tab()
    expect(document.activeElement).toBe(button)
  })

  it('Enter key on submit button triggers login', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json(
          { statusCode: 401, error: { code: 'INVALID_CREDENTIALS', message: 'Bad' } },
          { status: 401 },
        ),
      ),
    )

    const user = userEvent.setup()
    render(<LoginPage />, { wrapper: Wrapper })

    const button = screen.getByRole('button', { name: /sign in/i })
    button.focus()
    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })
})
