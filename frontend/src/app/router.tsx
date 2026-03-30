import { lazy, Suspense } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

// Lazy-loaded page components (added as stories are implemented)
const UploadPage = lazy(() => import('@/pages/UploadPage'))
const DocumentDetailPage = lazy(() => import('@/pages/DocumentDetailPage'))
const PaymentDashboardPage = lazy(() => import('@/pages/PaymentDashboardPage'))
const ReconciliationPage = lazy(() => import('@/pages/ReconciliationPage'))
const HistoryDashboardPage = lazy(() => import('@/pages/HistoryDashboardPage'))
const SettingsPage = lazy(() => import('@/pages/SettingsPage'))

function PageLoader() {
  return (
    <div className="flex h-full w-full items-center justify-center">
      <span className="text-text-secondary text-sm">Loading…</span>
    </div>
  )
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/" element={<Navigate to="/upload" replace />} />
          <Route path="/upload" element={<UploadPage />} />
          <Route path="/documents/:id" element={<DocumentDetailPage />} />
          <Route path="/payments" element={<PaymentDashboardPage />} />
          <Route path="/reconcile" element={<ReconciliationPage />} />
          <Route path="/history" element={<HistoryDashboardPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}
