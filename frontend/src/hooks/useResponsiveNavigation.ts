/**
 * useResponsiveNavigation — tracks viewport size and exposes breakpoint booleans
 * for the sidebar/hamburger responsive behaviour.
 *
 * Breakpoints:
 *   desktop  : ≥ 1024 px  — sidebar always visible, hamburger hidden
 *   tablet   : 768–1023 px — sidebar toggleable via hamburger
 *   mobile   : < 768 px   — sidebar hidden by default, overlay when open
 */

import { useCallback, useEffect, useState } from 'react'

export type Breakpoint = 'desktop' | 'tablet' | 'mobile'

export interface UseResponsiveNavigationResult {
  breakpoint: Breakpoint
  isDesktop: boolean
  isTablet: boolean
  isMobile: boolean
  sidebarOpen: boolean
  openSidebar: () => void
  closeSidebar: () => void
  toggleSidebar: () => void
}

function getBreakpoint(width: number): Breakpoint {
  if (width >= 1024) return 'desktop'
  if (width >= 768) return 'tablet'
  return 'mobile'
}

export function useResponsiveNavigation(): UseResponsiveNavigationResult {
  const [breakpoint, setBreakpoint] = useState<Breakpoint>(() =>
    getBreakpoint(window.innerWidth),
  )
  const [sidebarOpen, setSidebarOpen] = useState(() => window.innerWidth >= 1024)

  useEffect(() => {
    const mql = window.matchMedia('(min-width: 1024px)')
    const mqlTablet = window.matchMedia('(min-width: 768px)')

    function handleResize() {
      const bp = getBreakpoint(window.innerWidth)
      setBreakpoint(bp)
      // Auto-open sidebar when reaching desktop width
      if (bp === 'desktop') setSidebarOpen(true)
    }

    window.addEventListener('resize', handleResize)
    mql.addEventListener('change', handleResize)
    mqlTablet.addEventListener('change', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      mql.removeEventListener('change', handleResize)
      mqlTablet.removeEventListener('change', handleResize)
    }
  }, [])

  const openSidebar = useCallback(() => setSidebarOpen(true), [])
  const closeSidebar = useCallback(() => setSidebarOpen(false), [])
  const toggleSidebar = useCallback(() => setSidebarOpen((prev) => !prev), [])

  return {
    breakpoint,
    isDesktop: breakpoint === 'desktop',
    isTablet: breakpoint === 'tablet',
    isMobile: breakpoint === 'mobile',
    sidebarOpen,
    openSidebar,
    closeSidebar,
    toggleSidebar,
  }
}
