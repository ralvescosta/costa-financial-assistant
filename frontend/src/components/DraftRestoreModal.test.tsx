import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DraftRestoreModal } from '@/components/DraftRestoreModal'

describe('DraftRestoreModal', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.documentElement.classList.remove('dark')
  })

  it('uses the shared primary-action token contract for the restore button', () => {
    document.documentElement.classList.add('dark')

    render(
      <DraftRestoreModal
        draftKey="draft-1"
        onRestore={vi.fn()}
        onDiscard={vi.fn()}
        remainingSeconds={30}
      />,
    )

    const button = screen.getByRole('button', { name: /restore/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
    expect(button.className).toContain('hover:bg-[color:var(--color-primary-action-hover)]')
  })
})
