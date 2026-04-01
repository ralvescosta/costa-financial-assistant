import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles/index.css'
import { AppRouter } from './app/router'
import { AppProviders } from './app/providers'
import { ThemeProvider } from './app/theme/ThemeProvider'
import { AuthProvider } from './hooks/useAuthContext'
import { ErrorBoundary } from './app/ErrorBoundary'

const container = document.getElementById('root')
if (!container) {
  throw new Error('Root element #root not found')
}

createRoot(container).render(
  <StrictMode>
    <ErrorBoundary>
      <ThemeProvider>
        <AuthProvider>
          <AppProviders>
            <AppRouter />
          </AppProviders>
        </AuthProvider>
      </ThemeProvider>
    </ErrorBoundary>
  </StrictMode>
)
