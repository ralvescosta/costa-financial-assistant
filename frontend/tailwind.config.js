/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: 'var(--color-primary)',
        'primary-hover': 'var(--color-primary-hover)',
        surface: 'var(--color-surface)',
        'surface-raised': 'var(--color-surface-raised)',
        border: 'var(--color-border)',
        'text-primary': 'var(--color-text-primary)',
        'text-secondary': 'var(--color-text-secondary)',
        'text-disabled': 'var(--color-text-disabled)',
        danger: 'var(--color-danger)',
        'danger-bg': 'var(--color-danger-bg)',
        success: 'var(--color-success)',
        'success-bg': 'var(--color-success-bg)',
        warning: 'var(--color-warning)',
        'warning-bg': 'var(--color-warning-bg)',
        overlay: 'var(--color-overlay)',
        // Auth tokens
        'input-error': 'var(--color-input-error)',
        'lockout-warning': 'var(--color-lockout-warning)',
        'loading-skeleton': 'var(--color-loading-skeleton)',
        // Navigation tokens
        'sidebar-bg': 'var(--color-sidebar-bg)',
        'menu-item-text': 'var(--color-menu-item-text)',
        'menu-item-active-bg': 'var(--color-menu-item-active-bg)',
        'hamburger-icon': 'var(--color-hamburger-icon)',
        // Session tokens
        'session-warning': 'var(--color-session-warning)',
        'draft-restore-modal': 'var(--color-draft-restore-modal)',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
