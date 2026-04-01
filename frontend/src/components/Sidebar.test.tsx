/**
 * Unit tests for Sidebar component.
 *
 * Covers:
 * - All menu items render
 * - Current route is highlighted (aria-current="page")
 * - Click navigates and calls onClose
 * - Keyboard accessibility
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { Sidebar } from '@/components/Sidebar'

function renderSidebar(path = '/dashboard', isOpen = true) {
  const onClose = vi.fn()
  render(
    <MemoryRouter initialEntries={[path]}>
      <Sidebar isOpen={isOpen} onClose={onClose} />
    </MemoryRouter>,
  )
  return { onClose }
}

describe('Sidebar', () => {
  it('renders all navigation menu items', () => {
    renderSidebar()
    expect(screen.getByRole('link', { name: /dashboard/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /documents/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /bills/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /payments/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /analytics/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /settings/i })).toBeInTheDocument()
  })

  it('marks the current route link with aria-current="page"', () => {
    renderSidebar('/dashboard')
    const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
    expect(dashboardLink).toHaveAttribute('aria-current', 'page')
  })

  it('calls onClose when a nav item is clicked', async () => {
    const user = userEvent.setup()
    const { onClose } = renderSidebar('/settings')
    await user.click(screen.getByRole('link', { name: /dashboard/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('has an accessible nav landmark label', () => {
    renderSidebar()
    expect(screen.getByRole('navigation', { name: /main navigation/i })).toBeInTheDocument()
  })

  it('nav items are keyboard-focusable', () => {
    renderSidebar()
    const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
    dashboardLink.focus()
    expect(document.activeElement).toBe(dashboardLink)
  })
})
