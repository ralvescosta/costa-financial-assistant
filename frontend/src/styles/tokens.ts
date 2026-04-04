/**
 * Design token system for Costa Financial Assistant.
 *
 * Primitive tokens define the raw palette.
 * Semantic tokens map palette values to intent-driven names.
 * Components reference ONLY semantic tokens — never primitives directly.
 *
 * CSS variable binding is in src/styles/index.css.
 * Tailwind mapping is in tailwind.config.js.
 */

// ─── Primitive palette ──────────────────────────────────────────────────────
export const palette = {
  // Brand
  indigo50: '#eef2ff',
  indigo600: '#4f46e5',
  indigo700: '#4338ca',

  // Neutrals
  white: '#ffffff',
  gray50: '#f9fafb',
  gray100: '#f3f4f6',
  gray200: '#e5e7eb',
  gray400: '#9ca3af',
  gray600: '#4b5563',
  gray900: '#111827',

  // Semantic feedback
  red50: '#fef2f2',
  red600: '#dc2626',
  green50: '#f0fdf4',
  green600: '#16a34a',
  amber50: '#fffbeb',
  amber500: '#f59e0b',
} as const

// ─── Semantic token definitions ─────────────────────────────────────────────
type SemanticTokens = {
  colorPrimary: string
  colorPrimaryHover: string
  colorPrimaryActionBg: string
  colorPrimaryActionFg: string
  colorPrimaryActionHover: string
  colorPrimaryActionFocus: string
  colorPrimaryActionDisabledBg: string
  colorPrimaryActionDisabledFg: string
  colorSurface: string
  colorSurfaceRaised: string
  colorBorder: string
  colorTextPrimary: string
  colorTextSecondary: string
  colorTextDisabled: string
  colorDanger: string
  colorDangerBg: string
  colorSuccess: string
  colorSuccessBg: string
  colorWarning: string
  colorWarningBg: string
  colorOverlay: string
  // Auth / login-specific tokens
  colorInputError: string
  colorLockoutWarning: string
  colorLoadingSkeleton: string
  // Navigation tokens
  colorSidebarBg: string
  colorMenuItemText: string
  colorMenuItemActiveBg: string
  colorHamburgerIcon: string
  // Session / draft restore tokens
  colorSessionWarning: string
  colorDraftRestoreModal: string
}

export const lightTokens: SemanticTokens = {
  colorPrimary: palette.indigo600,
  colorPrimaryHover: palette.indigo700,
  colorPrimaryActionBg: palette.indigo600,
  colorPrimaryActionFg: palette.white,
  colorPrimaryActionHover: palette.indigo700,
  colorPrimaryActionFocus: palette.indigo600,
  colorPrimaryActionDisabledBg: palette.gray200,
  colorPrimaryActionDisabledFg: palette.gray600,
  colorSurface: palette.white,
  colorSurfaceRaised: palette.gray50,
  colorBorder: palette.gray200,
  colorTextPrimary: palette.gray900,
  colorTextSecondary: palette.gray600,
  colorTextDisabled: palette.gray400,
  colorDanger: palette.red600,
  colorDangerBg: palette.red50,
  colorSuccess: palette.green600,
  colorSuccessBg: palette.green50,
  colorWarning: palette.amber500,
  colorWarningBg: palette.amber50,
  colorOverlay: 'rgba(0, 0, 0, 0.4)',
  // Auth
  colorInputError: palette.red600,
  colorLockoutWarning: palette.amber500,
  colorLoadingSkeleton: palette.gray200,
  // Navigation
  colorSidebarBg: palette.white,
  colorMenuItemText: palette.gray600,
  colorMenuItemActiveBg: palette.indigo50,
  colorHamburgerIcon: palette.gray600,
  // Session
  colorSessionWarning: palette.amber500,
  colorDraftRestoreModal: palette.white,
}

export const darkTokens: SemanticTokens = {
  colorPrimary: palette.indigo50,
  colorPrimaryHover: palette.indigo600,
  colorPrimaryActionBg: palette.indigo600,
  colorPrimaryActionFg: palette.white,
  colorPrimaryActionHover: palette.indigo700,
  colorPrimaryActionFocus: palette.indigo50,
  colorPrimaryActionDisabledBg: '#3a3a5c',
  colorPrimaryActionDisabledFg: '#a0aec0',
  colorSurface: '#1e1e2e',
  colorSurfaceRaised: '#2a2a3d',
  colorBorder: '#3a3a5c',
  colorTextPrimary: '#e2e8f0',
  colorTextSecondary: '#a0aec0',
  colorTextDisabled: '#718096',
  colorDanger: '#fc8181',
  colorDangerBg: '#3b1a1a',
  colorSuccess: '#68d391',
  colorSuccessBg: '#1a3b1a',
  colorWarning: '#f6e05e',
  colorWarningBg: '#3b2f1a',
  colorOverlay: 'rgba(0, 0, 0, 0.7)',
  // Auth
  colorInputError: '#fc8181',
  colorLockoutWarning: '#f6e05e',
  colorLoadingSkeleton: '#3a3a5c',
  // Navigation
  colorSidebarBg: '#1e1e2e',
  colorMenuItemText: '#a0aec0',
  colorMenuItemActiveBg: '#2a2a3d',
  colorHamburgerIcon: '#a0aec0',
  // Session
  colorSessionWarning: '#f6e05e',
  colorDraftRestoreModal: '#2a2a3d',
}

export type Theme = 'light' | 'dark'
