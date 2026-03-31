import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getCurrentProject,
  listProjectMembers,
  inviteProjectMember,
  updateProjectMemberRole,
} from '@/services/projectsApi'
import type { ProjectResponse, ProjectMemberResponse, ProjectMemberRole } from '@/services/projectsApi'

// ─── Query keys ───────────────────────────────────────────────────────────────

export const projectQueryKeys = {
  currentProject: () => ['project', 'current'] as const,
  members: (pageSize?: number, pageToken?: string) =>
    ['project', 'members', pageSize, pageToken ?? ''] as const,
}

// ─── useCurrentProject ───────────────────────────────────────────────────────

export interface UseCurrentProjectResult {
  project: ProjectResponse | undefined
  isLoading: boolean
  isError: boolean
  error: Error | null
  refetch: () => void
}

/**
 * useCurrentProject — query hook that returns the project bound to the
 * caller's JWT claims. The project context is central to all project-scoped
 * operations in the application.
 *
 * Usage:
 *   const { project, isLoading } = useCurrentProject()
 */
export function useCurrentProject(): UseCurrentProjectResult {
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: projectQueryKeys.currentProject(),
    queryFn: getCurrentProject,
  })

  return {
    project: data,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  }
}

// ─── useProjectMembers ───────────────────────────────────────────────────────

export interface UseProjectMembersResult {
  members: ProjectMemberResponse[]
  nextPageToken: string | undefined
  isLoading: boolean
  isError: boolean
  error: Error | null
}

/**
 * useProjectMembers — paginated query hook that returns all members of the
 * current project.
 */
export function useProjectMembers(pageSize = 25, pageToken?: string): UseProjectMembersResult {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: projectQueryKeys.members(pageSize, pageToken),
    queryFn: () => listProjectMembers(pageSize, pageToken),
  })

  return {
    members: data?.items ?? [],
    nextPageToken: data?.nextPageToken,
    isLoading,
    isError,
    error: error as Error | null,
  }
}

// ─── useInviteProjectMember ───────────────────────────────────────────────────

export interface InviteMemberVars {
  email: string
  role: ProjectMemberRole
}

/**
 * useInviteProjectMember — mutation hook for inviting a user to the project.
 * Invalidates the members list on success.
 */
export function useInviteProjectMember() {
  const queryClient = useQueryClient()

  return useMutation<ProjectMemberResponse, Error, InviteMemberVars>({
    mutationFn: ({ email, role }) => inviteProjectMember(email, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', 'members'] })
    },
  })
}

// ─── useUpdateProjectMemberRole ───────────────────────────────────────────────

export interface UpdateMemberRoleVars {
  memberId: string
  role: ProjectMemberRole
}

/**
 * useUpdateProjectMemberRole — mutation hook that changes the role of an
 * existing project member. Invalidates the members list on success.
 */
export function useUpdateProjectMemberRole() {
  const queryClient = useQueryClient()

  return useMutation<ProjectMemberResponse, Error, UpdateMemberRoleVars>({
    mutationFn: ({ memberId, role }) => updateProjectMemberRole(memberId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', 'members'] })
    },
  })
}
