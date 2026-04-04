/**
 * ErrorBoundary — catches render errors in the application tree and
 * displays a fallback with a session-expiry message.
 */

import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // Log to observability — do NOT log credentials or sensitive state
    console.error('[ErrorBoundary] Render error:', error.message, info.componentStack)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8 text-center">
          <h1 className="text-xl font-semibold text-[color:var(--color-text-primary)]">
            Something went wrong
          </h1>
          <p className="text-sm text-[color:var(--color-text-secondary)]">
            Your session may have expired. Please refresh the page to continue.
          </p>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="
              rounded-md bg-[color:var(--color-primary-action-bg)] px-4 py-2
              text-sm font-semibold text-[color:var(--color-primary-action-fg)]
              hover:bg-[color:var(--color-primary-action-hover)]
              active:bg-[color:var(--color-primary-action-hover)]
              focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary-action-focus)]
              focus:ring-offset-2 focus:ring-offset-[color:var(--color-surface)]
              transition-colors duration-150
            "
          >
            Reload page
          </button>
        </div>
      )
    }

    return this.props.children
  }
}
