/**
 * TypeScript interfaces for navigation, sidebar, and breadcrumb state.
 */

export interface MenuItem {
  id: string
  label: string
  path: string
  /** Lucide/icon identifier or component name */
  icon?: string
  /** Whether this item requires authentication */
  requiresAuth?: boolean
}

export interface NavigationConfig {
  items: MenuItem[]
}

export interface NavigationState {
  currentRoute: string
  activeMenuItem?: string
  sidebarOpen: boolean
  breadcrumbs: BreadcrumbItem[]
}

export interface BreadcrumbItem {
  label: string
  route: string
  icon?: string
}
