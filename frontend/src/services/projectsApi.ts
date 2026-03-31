const BASE = '/api/v1/projects'

/** Shape returned by GET /api/v1/projects/current */
export interface ProjectResponse {
  id: string
  ownerId: string
  name: string
  createdAt: string
  updatedAt: string
}

/** Shape of a project member */
export interface ProjectMemberResponse {
  id: string
  projectId: string
  userId: string
  role: 'read_only' | 'update' | 'write'
  invitedBy?: string
  createdAt: string
  updatedAt: string
}

/** Shape returned by GET /api/v1/projects/members */
export interface ListMembersResponse {
  items: ProjectMemberResponse[]
  nextPageToken?: string
}

/** Role values used throughout the collaboration UI */
export type ProjectMemberRole = 'read_only' | 'update' | 'write'

/**
 * Fetches the project associated with the caller's JWT claims.
 */
export async function getCurrentProject(): Promise<ProjectResponse> {
  const res = await fetch(`${BASE}/current`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Get project failed: ${res.status}`)
  }
  return res.json() as Promise<ProjectResponse>
}

/**
 * Lists all members of the caller's project.
 */
export async function listProjectMembers(
  pageSize = 25,
  pageToken?: string,
): Promise<ListMembersResponse> {
  const params = new URLSearchParams({ pageSize: String(pageSize) })
  if (pageToken) params.set('pageToken', pageToken)
  const res = await fetch(`${BASE}/members?${params}`)
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `List members failed: ${res.status}`)
  }
  return res.json() as Promise<ListMembersResponse>
}

/**
 * Invites a user by email to the current project with the given role.
 */
export async function inviteProjectMember(
  email: string,
  role: ProjectMemberRole,
): Promise<ProjectMemberResponse> {
  const res = await fetch(`${BASE}/members/invite`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, role }),
  })
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Invite member failed: ${res.status}`)
  }
  return res.json() as Promise<ProjectMemberResponse>
}

/**
 * Updates the role of an existing project member.
 */
export async function updateProjectMemberRole(
  memberId: string,
  role: ProjectMemberRole,
): Promise<ProjectMemberResponse> {
  const res = await fetch(`${BASE}/members/${encodeURIComponent(memberId)}/role`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ role }),
  })
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { title?: string } | null
    throw new Error(body?.title ?? `Update member role failed: ${res.status}`)
  }
  return res.json() as Promise<ProjectMemberResponse>
}
