import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ErrorBoundary } from '@/app/ErrorBoundary'

function ThrowOnRender(): never {
  throw new Error('boom')
}

describe('ErrorBoundary', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.documentElement.classList.remove('dark')
  })

  it('uses the shared primary-action token contract for the recovery button', () => {
    document.documentElement.classList.add('dark')
    vi.spyOn(console, 'error').mockImplementation(() => { })

    render(
      <ErrorBoundary>
        <ThrowOnRender />
      </ErrorBoundary>,
    )

    const button = screen.getByRole('button', { name: /reload page/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
    expect(button.className).toContain('hover:bg-[color:var(--color-primary-action-hover)]')
  })

  it('keeps the same shared primary-action contract in light mode', () => {
    vi.spyOn(console, 'error').mockImplementation(() => { })

    render(
      <ErrorBoundary>
        <ThrowOnRender />
      </ErrorBoundary>,
    )

    const button = screen.getByRole('button', { name: /reload page/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
  })
})
