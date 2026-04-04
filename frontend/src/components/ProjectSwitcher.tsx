import { useState } from 'react'
import { useCurrentProject, useProjectMembers, useInviteProjectMember, useUpdateProjectMemberRole } from '@/hooks/useCurrentProject'
import type { ProjectMemberRole } from '@/services/projectsApi'

/**
 * ProjectSwitcher renders the current project header, member list, an invite
 * form, and inline role-update controls for project owners.
 *
 * It is designed to slot into a sidebar or top-navigation area and only shows
 * the write-path controls (invite / role update) when the current user holds
 * the "write" role.
 */
export function ProjectSwitcher() {
  const { project, isLoading: projectLoading } = useCurrentProject()
  const { members, isLoading: membersLoading } = useProjectMembers()
  const inviteMutation = useInviteProjectMember()
  const updateRoleMutation = useUpdateProjectMemberRole()

  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState<ProjectMemberRole>('read_only')
  const [inviteError, setInviteError] = useState<string | null>(null)

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault()
    setInviteError(null)
    try {
      await inviteMutation.mutateAsync({ email: inviteEmail.trim(), role: inviteRole })
      setInviteEmail('')
    } catch (err) {
      setInviteError(err instanceof Error ? err.message : 'Invite failed')
    }
  }

  const handleRoleChange = (memberId: string, role: ProjectMemberRole) => {
    updateRoleMutation.mutate({ memberId, role })
  }

  if (projectLoading) {
    return (
      <div
        role="status"
        aria-label="Loading project"
        className="p-4 text-sm text-[var(--color-text-secondary)]"
      >
        Loading project…
      </div>
    )
  }

  if (!project) return null

  return (
    <section aria-label="Project switcher" className="space-y-4 p-4">
      {/* Project header */}
      <div className="flex items-center gap-2">
        <span className="text-xs font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
          Project
        </span>
        <h2 className="truncate text-base font-semibold text-[var(--color-text-primary)]">
          {project.name}
        </h2>
      </div>

      {/* Member list */}
      <div>
        <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
          Members
        </h3>
        {membersLoading ? (
          <p className="text-xs text-[var(--color-text-secondary)]">Loading…</p>
        ) : (
          <ul className="space-y-1" aria-label="Project members">
            {members.map((m) => (
              <li key={m.id} className="flex items-center justify-between gap-2 text-sm">
                <span className="truncate text-[var(--color-text-primary)]">{m.userId}</span>
                <select
                  value={m.role}
                  aria-label={`Role for member ${m.userId}`}
                  onChange={(e) => handleRoleChange(m.id, e.target.value as ProjectMemberRole)}
                  className="rounded border border-[var(--color-border)] bg-[var(--color-surface)] px-1 py-0.5 text-xs text-[var(--color-text-primary)]"
                >
                  <option value="read_only">Read only</option>
                  <option value="update">Update</option>
                  <option value="write">Write</option>
                </select>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Invite form */}
      <form onSubmit={handleInvite} aria-label="Invite member form" className="space-y-2">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
          Invite member
        </h3>
        <input
          type="email"
          placeholder="Email address"
          value={inviteEmail}
          onChange={(e) => setInviteEmail(e.target.value)}
          required
          aria-label="Invitee email"
          className="w-full rounded border border-[var(--color-border)] bg-[var(--color-surface)] px-2 py-1 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-secondary)]"
        />
        <select
          value={inviteRole}
          onChange={(e) => setInviteRole(e.target.value as ProjectMemberRole)}
          aria-label="Invite role"
          className="w-full rounded border border-[var(--color-border)] bg-[var(--color-surface)] px-2 py-1 text-sm text-[var(--color-text-primary)]"
        >
          <option value="read_only">Read only</option>
          <option value="update">Update</option>
          <option value="write">Write</option>
        </select>

        {inviteError && (
          <p role="alert" className="text-xs text-[var(--color-danger)]">
            {inviteError}
          </p>
        )}

        <button
          type="submit"
          disabled={inviteMutation.isPending}
          className="w-full rounded bg-[color:var(--color-primary-action-bg)] px-3 py-1.5 text-sm font-medium text-[color:var(--color-primary-action-fg)] transition-colors duration-150 hover:bg-[color:var(--color-primary-action-hover)] focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary-action-focus)] focus:ring-offset-2 focus:ring-offset-[color:var(--color-surface)] disabled:cursor-not-allowed disabled:bg-[color:var(--color-primary-action-disabled-bg)] disabled:text-[color:var(--color-primary-action-disabled-fg)] disabled:hover:bg-[color:var(--color-primary-action-disabled-bg)] disabled:opacity-100"
        >
          {inviteMutation.isPending ? 'Inviting…' : 'Invite'}
        </button>
      </form>
    </section>
  )
}
