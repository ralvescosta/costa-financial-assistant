/**
 * Integration test: Login → dashboard navigation flow.
 *
 * User lands on /login → sees auto-filled credentials → submits → lands on /dashboard.
 */

import { describe, it, expect, beforeAll, afterEach, afterAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { LoginPage } from '@/pages/LoginPage'

const server = setupServer()

beforeAll(() => server.listen())
afterEach(() => {
  server.resetHandlers()
  localStorage.clear()
  document.documentElement.classList.remove('dark')
})
afterAll(() => server.close())

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function DashboardStub() {
  return <div>Dashboard</div>
}

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  )
}

describe('LoginPage integration', () => {
  it('navigates to /dashboard after successful login', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json({
          statusCode: 200,
          data: {
            expiresIn: 3600,
            refreshAt: 2700,
            csrfToken: 'csrf',
            user: { id: 'u1', username: 'demo' },
          },
        }),
      ),
    )

    const user = userEvent.setup()
    render(
      <Wrapper>
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/dashboard" element={<DashboardStub />} />
          </Routes>
        </MemoryRouter>
      </Wrapper>,
    )

    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('preserves the dark-mode action contract on first paint after a refresh', () => {
    document.documentElement.classList.add('dark')

    const { unmount } = render(
      <Wrapper>
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
          </Routes>
        </MemoryRouter>
      </Wrapper>,
    )

    let button = screen.getByRole('button', { name: /sign in/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')

    unmount()

    render(
      <Wrapper>
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
          </Routes>
        </MemoryRouter>
      </Wrapper>,
    )

    button = screen.getByRole('button', { name: /sign in/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
  })
})
