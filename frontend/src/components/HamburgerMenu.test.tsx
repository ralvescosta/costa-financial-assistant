/**
 * Unit tests for HamburgerMenu component.
 *
 * Covers:
 * - Renders a button with accessible label
 * - aria-pressed reflects isOpen state
 * - Calls onToggle on click
 * - Has lg:hidden class (desktop hidden behaviour)
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HamburgerMenu } from '@/components/HamburgerMenu'

describe('HamburgerMenu', () => {
  it('renders a button with accessible label', () => {
    render(<HamburgerMenu isOpen={false} onToggle={vi.fn()} />)
    expect(screen.getByRole('button', { name: /open navigation menu/i })).toBeInTheDocument()
  })

  it('changes aria-label when open', () => {
    render(<HamburgerMenu isOpen={true} onToggle={vi.fn()} />)
    expect(screen.getByRole('button', { name: /close navigation menu/i })).toBeInTheDocument()
  })

  it('sets aria-pressed=true when open', () => {
    render(<HamburgerMenu isOpen={true} onToggle={vi.fn()} />)
    expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'true')
  })

  it('sets aria-pressed=false when closed', () => {
    render(<HamburgerMenu isOpen={false} onToggle={vi.fn()} />)
    expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'false')
  })

  it('calls onToggle when clicked', async () => {
    const onToggle = vi.fn()
    const user = userEvent.setup()
    render(<HamburgerMenu isOpen={false} onToggle={onToggle} />)
    await user.click(screen.getByRole('button'))
    expect(onToggle).toHaveBeenCalledOnce()
  })

  it('has lg:hidden class to hide on desktop', () => {
    render(<HamburgerMenu isOpen={false} onToggle={vi.fn()} />)
    const btn = screen.getByRole('button')
    expect(btn.className).toMatch(/lg:hidden/)
  })
})
