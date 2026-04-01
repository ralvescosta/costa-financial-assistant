/**
 * AppLayout — root layout component wiring Sidebar + HamburgerMenu together
 * with the main content area.
 *
 * Layout behaviour:
 *   desktop  (≥1024px) : fixed 200px sidebar + fluid content
 *   tablet/mobile      : full-width content, sidebar toggled via hamburger
 */

import type { ReactNode } from 'react'
import { Sidebar } from '@/components/Sidebar'
import { HamburgerMenu } from '@/components/HamburgerMenu'
import { useResponsiveNavigation } from '@/hooks/useResponsiveNavigation'
import { useAuthRefresh } from '@/hooks/useAuthRefresh'
import { useTokenRefreshInterceptor } from '@/hooks/useTokenRefreshInterceptor'

interface AppLayoutProps {
  children: ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  useAuthRefresh()
  useTokenRefreshInterceptor()

  const { sidebarOpen, toggleSidebar, closeSidebar, isDesktop } =
    useResponsiveNavigation()

  return (
    <div className="flex min-h-screen bg-[color:var(--color-surface)]">
      <Sidebar isOpen={sidebarOpen || isDesktop} onClose={closeSidebar} />

      <div className="flex flex-1 flex-col min-w-0">
        {/* Top bar (mobile/tablet only) */}
        <header className="flex h-16 items-center border-b border-[color:var(--color-border)] bg-[color:var(--color-surface)] px-4 lg:hidden">
          <HamburgerMenu isOpen={sidebarOpen} onToggle={toggleSidebar} />
          <span className="ml-3 text-sm font-semibold text-[color:var(--color-text-primary)]">
            Financial App
          </span>
        </header>

        <main id="main-content" className="flex-1 overflow-auto p-6">
          {children}
        </main>
      </div>
    </div>
  )
}
