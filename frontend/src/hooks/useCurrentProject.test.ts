import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement } from 'react'
import {
  useCurrentProject,
  useInviteProjectMember,
  useUpdateProjectMemberRole,
} from './useCurrentProject'
import type { InviteMemberVars, UpdateMemberRoleVars } from './useCurrentProject'
import * as projectsApi from '@/services/projectsApi'

// ─── Mock service module ──────────────────────────────────────────────────────

vi.mock('@/services/projectsApi', () => ({
  getCurrentProject: vi.fn(),
  listProjectMembers: vi.fn(),
  inviteProjectMember: vi.fn(),
  updateProjectMemberRole: vi.fn(),
}))

const mockGetCurrentProject = vi.mocked(projectsApi.getCurrentProject)
const mockInvite = vi.mocked(projectsApi.inviteProjectMember)
const mockUpdateRole = vi.mocked(projectsApi.updateProjectMemberRole)

// ─── helpers ──────────────────────────────────────────────────────────────────

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const fakeProject: projectsApi.ProjectResponse = {
  id: 'proj-uuid-1',
  ownerId: 'user-uuid-1',
  name: 'My Finances',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

const fakeMember: projectsApi.ProjectMemberResponse = {
  id: 'member-uuid-1',
  projectId: 'proj-uuid-1',
  userId: 'user-uuid-2',
  role: 'read_only',
  invitedBy: 'user-uuid-1',
  createdAt: '2024-01-02T00:00:00Z',
  updatedAt: '2024-01-02T00:00:00Z',
}

// ─── tests ────────────────────────────────────────────────────────────────────

describe('useCurrentProject', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('given the API returns a project, when the hook mounts, then project data is returned', () => {
    it('exposes the project name and owner', async () => {
      // Arrange
      mockGetCurrentProject.mockResolvedValueOnce(fakeProject)

      // Act
      const { result } = renderHook(() => useCurrentProject(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.project?.name).toBe('My Finances')
      expect(result.current.project?.ownerId).toBe('user-uuid-1')
      expect(result.current.isError).toBe(false)
      expect(result.current.error).toBeNull()
      expect(mockGetCurrentProject).toHaveBeenCalledOnce()
    })
  })

  describe('given the API fails, when the hook mounts, then error state is exposed', () => {
    it('sets isError and returns the error object', async () => {
      // Arrange
      mockGetCurrentProject.mockRejectedValueOnce(new Error('Get project failed: 403'))

      // Act
      const { result } = renderHook(() => useCurrentProject(), {
        wrapper: makeWrapper(),
      })

      // Assert
      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.isError).toBe(true)
      expect(result.current.project).toBeUndefined()
      expect(result.current.error?.message).toContain('403')
    })
  })

  describe('given a write-role caller, when invite is called with email and role, then member is returned', () => {
    it('resolves with the new member on success', async () => {
      // Arrange
      mockInvite.mockResolvedValueOnce(fakeMember)
      const { result } = renderHook(() => useInviteProjectMember(), {
        wrapper: makeWrapper(),
      })

      // Act
      let member: projectsApi.ProjectMemberResponse | undefined
      await act(async () => {
        const vars: InviteMemberVars = { email: 'alice@example.com', role: 'read_only' }
        member = await result.current.mutateAsync(vars)
      })

      // Assert
      expect(member?.userId).toBe('user-uuid-2')
      expect(member?.role).toBe('read_only')
      // Verify the API was called with the right args
      expect(mockInvite.mock.calls[0][0]).toBe('alice@example.com')
      expect(mockInvite.mock.calls[0][1]).toBe('read_only')
    })
  })

  describe('given an existing member, when role update is called, then updated member is returned', () => {
    it('resolves with the updated member on role change', async () => {
      // Arrange
      const updatedMember = { ...fakeMember, role: 'write' as const }
      mockUpdateRole.mockResolvedValueOnce(updatedMember)
      const { result } = renderHook(() => useUpdateProjectMemberRole(), {
        wrapper: makeWrapper(),
      })

      // Act
      let member: projectsApi.ProjectMemberResponse | undefined
      await act(async () => {
        const vars: UpdateMemberRoleVars = { memberId: 'member-uuid-1', role: 'write' }
        member = await result.current.mutateAsync(vars)
      })

      // Assert
      expect(member?.role).toBe('write')
      expect(mockUpdateRole.mock.calls[0][0]).toBe('member-uuid-1')
      expect(mockUpdateRole.mock.calls[0][1]).toBe('write')
    })
  })

  describe('given invite fails, when mutate is called, then error is propagated', () => {
    it('throws an error on invite failure', async () => {
      // Arrange
      mockInvite.mockRejectedValueOnce(new Error('Invite member failed: 409'))
      const { result } = renderHook(() => useInviteProjectMember(), {
        wrapper: makeWrapper(),
      })

      // Act + Assert
      await expect(
        act(async () => {
          await result.current.mutateAsync({ email: 'dup@example.com', role: 'read_only' })
        }),
      ).rejects.toThrow('Invite member failed: 409')
    })
  })
})
