// Vitest global setup — runs before each test file.
import '@testing-library/jest-dom'
import { vi } from 'vitest'

// jsdom does not implement window.matchMedia. Provide a minimal mock so hooks
// that depend on media queries (e.g. useResponsiveNavigation) work in tests.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})
