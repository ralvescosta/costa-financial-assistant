/**
 * Sidebar — persistent navigation sidebar.
 *
 * Renders all main navigation menu items, highlights the active route, and
 * supports keyboard navigation. On mobile it acts as an overlay controlled
 * by the parent layout via `isOpen` / `onClose` props.
 */

import { NavLink, useMatch } from 'react-router-dom'
import { NAVIGATION_ITEMS } from '@/app/router.config'
import type { MenuItem } from '@/types/navigation'

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

function NavItem({ item, onClose }: { item: MenuItem; onClose: () => void }) {
  const match = useMatch(item.path)
  return (
    <NavLink
      to={item.path}
      onClick={onClose}
      aria-current={match ? 'page' : undefined}
      className={({ isActive }) =>
        [
          'flex items-center gap-3 px-4 py-2 rounded-md text-sm font-medium',
          'transition-colors duration-150',
          'focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary)] focus:ring-offset-2 focus:ring-offset-[color:var(--color-sidebar-bg)]',
          isActive
            ? 'bg-[color:var(--color-menu-item-active-bg)] text-[color:var(--color-primary)] font-semibold'
            : 'text-[color:var(--color-menu-item-text)] hover:bg-surface-raised',
        ].join(' ')
      }
    >
      {item.icon && (
        <span className="text-lg" aria-hidden="true">
          {item.icon}
        </span>
      )}
      {item.label}
    </NavLink>
  )
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  return (
    <>
      {/* Mobile overlay backdrop */}
      {isOpen && (
        <div
          role="presentation"
          className="fixed inset-0 z-20 bg-[color:var(--color-overlay)] lg:hidden"
          onClick={onClose}
        />
      )}

      <nav
        aria-label="Main navigation"
        className={[
          'fixed inset-y-0 left-0 z-30 flex flex-col',
          'w-[var(--sidebar-width,200px)] bg-[color:var(--color-sidebar-bg)]',
          'border-r border-[color:var(--color-border)]',
          'transition-transform duration-200 ease-in-out',
          'lg:translate-x-0 lg:static lg:z-auto',
          isOpen ? 'translate-x-0' : '-translate-x-full',
        ].join(' ')}
      >
        {/* Logo / brand area */}
        <div className="flex h-16 items-center px-4 border-b border-[color:var(--color-border)]">
          <span className="text-base font-semibold text-[color:var(--color-text-primary)]">
            Financial App
          </span>
        </div>

        {/* Navigation items */}
        <ul role="list" className="flex flex-col gap-1 p-3 flex-1 overflow-y-auto">
          {NAVIGATION_ITEMS.map((item) => (
            <li key={item.id}>
              <NavItem item={item} onClose={onClose} />
            </li>
          ))}
        </ul>
      </nav>
    </>
  )
}
