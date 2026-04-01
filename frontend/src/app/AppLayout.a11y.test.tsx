/**
 * Accessibility test: AppLayout sidebar keyboard traversal.
 */

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AuthProvider } from '@/hooks/useAuthContext'
import { AppLayout } from '@/app/AppLayout'

const queryClient = new QueryClient()

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter>{children}</MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>
  )
}

describe('AppLayout accessibility', () => {
  it('navigation is keyboard-traversable', () => {
    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>,
      { wrapper: Wrapper },
    )

    const nav = screen.getByRole('navigation', { name: /main navigation/i })
    const links = nav.querySelectorAll('a')
    expect(links.length).toBeGreaterThan(0)

    links.forEach((link) => {
      expect(link.getAttribute('tabindex')).not.toBe('-1')
    })
  })

  it('hamburger button has aria-label', () => {
    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>,
      { wrapper: Wrapper },
    )

    // The hamburger is only visible at mobile viewports but rendered in DOM
    const hamburgers = document.querySelectorAll('[aria-label*="navigation menu"]')
    expect(hamburgers.length).toBeGreaterThan(0)
  })
})
