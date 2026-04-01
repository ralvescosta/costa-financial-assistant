import { lazy, Suspense } from 'react'
import { BrowserRouter, Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { AppLayout } from '@/app/AppLayout'
import { useAuthContext } from '@/hooks/useAuthContext'
import { SkeletonPlaceholder } from '@/components/SkeletonPlaceholder'

// Auth page (no layout)
const LoginPage = lazy(() => import('@/pages/LoginPage').then((m) => ({ default: m.LoginPage })))

// Protected pages (rendered inside AppLayout)
const DashboardPage = lazy(() =>
  import('@/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })),
)
const UploadPage = lazy(() => import('@/pages/UploadPage'))
const DocumentDetailPage = lazy(() => import('@/pages/DocumentDetailPage'))
const PaymentDashboardPage = lazy(() => import('@/pages/PaymentDashboardPage'))
const ReconciliationPage = lazy(() => import('@/pages/ReconciliationPage'))
const HistoryDashboardPage = lazy(() => import('@/pages/HistoryDashboardPage'))
const SettingsPage = lazy(() => import('@/pages/SettingsPage'))
const BillsPage = lazy(() =>
  import('@/pages/BillsPage').then((m) => ({ default: m.BillsPage })),
)

function PageLoader() {
  return (
    <div className="flex h-full w-full flex-col gap-3 p-6">
      <SkeletonPlaceholder height="h-8" width="w-48" />
      <SkeletonPlaceholder height="h-4" width="w-full" />
      <SkeletonPlaceholder height="h-4" width="w-3/4" />
    </div>
  )
}

/**
 * Layout route that renders AppLayout + Outlet for all protected pages.
 * Unauthenticated users are redirected to /login.
 */
function ProtectedLayout() {
  const { isAuthenticated } = useAuthContext()
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return (
    <AppLayout>
      <Suspense fallback={<PageLoader />}>
        <Outlet />
      </Suspense>
    </AppLayout>
  )
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />

          {/* Protected routes — rendered inside AppLayout via ProtectedLayout */}
          <Route element={<ProtectedLayout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/upload" element={<UploadPage />} />
            <Route path="/documents/:id" element={<DocumentDetailPage />} />
            <Route path="/bills" element={<BillsPage />} />
            <Route path="/payments" element={<PaymentDashboardPage />} />
            <Route path="/reconcile" element={<ReconciliationPage />} />
            <Route path="/history" element={<HistoryDashboardPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}
