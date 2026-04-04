import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ProjectSwitcher } from '@/components/ProjectSwitcher'

const inviteAsync = vi.fn().mockResolvedValue(undefined)
const updateRole = vi.fn()

vi.mock('@/hooks/useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: 'project-1', name: 'Demo Project' },
    isLoading: false,
  }),
  useProjectMembers: () => ({
    members: [{ id: 'member-1', userId: 'owner@example.com', role: 'write' }],
    isLoading: false,
  }),
  useInviteProjectMember: () => ({
    isPending: false,
    mutateAsync: inviteAsync,
  }),
  useUpdateProjectMemberRole: () => ({
    mutate: updateRole,
  }),
}))

describe('ProjectSwitcher', () => {
  afterEach(() => {
    inviteAsync.mockClear()
    updateRole.mockClear()
    document.documentElement.classList.remove('dark')
  })

  it('uses the shared primary-action token contract for the invite button', () => {
    document.documentElement.classList.add('dark')

    render(<ProjectSwitcher />)

    const button = screen.getByRole('button', { name: /invite/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
    expect(button.className).toContain('disabled:bg-[color:var(--color-primary-action-disabled-bg)]')
  })

  it('keeps the same shared primary-action contract in light mode', () => {
    render(<ProjectSwitcher />)

    const button = screen.getByRole('button', { name: /invite/i })
    expect(button.className).toContain('bg-[color:var(--color-primary-action-bg)]')
    expect(button.className).toContain('text-[color:var(--color-primary-action-fg)]')
  })
})
