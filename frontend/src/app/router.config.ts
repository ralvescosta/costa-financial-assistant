/**
 * Route and navigation configuration.
 *
 * Centralises path constants, route elements, and sidebar menu metadata.
 * AppRouter (router.tsx) and Sidebar both consume this file.
 */

import type { MenuItem } from '@/types/navigation'

export const NAVIGATION_ITEMS: MenuItem[] = [
  {
    id: 'dashboard',
    label: 'Dashboard',
    path: '/dashboard',
    icon: '🏠',
    requiresAuth: true,
  },
  {
    id: 'documents',
    label: 'Documents',
    path: '/upload',
    icon: '📄',
    requiresAuth: true,
  },
  {
    id: 'bills',
    label: 'Bills',
    path: '/bills',
    icon: '💳',
    requiresAuth: true,
  },
  {
    id: 'payments',
    label: 'Payments',
    path: '/payments',
    icon: '💰',
    requiresAuth: true,
  },
  {
    id: 'analytics',
    label: 'Analytics',
    path: '/history',
    icon: '📊',
    requiresAuth: true,
  },
  {
    id: 'settings',
    label: 'Settings',
    path: '/settings',
    icon: '⚙️',
    requiresAuth: true,
  },
]

export const ROUTES = {
  LOGIN: '/login',
  DASHBOARD: '/dashboard',
  UPLOAD: '/upload',
  DOCUMENT_DETAIL: '/documents/:id',
  BILLS: '/bills',
  PAYMENTS: '/payments',
  HISTORY: '/history',
  RECONCILE: '/reconcile',
  SETTINGS: '/settings',
} as const
