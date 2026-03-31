import { lightTokens, darkTokens, type Theme } from '../../styles/tokens';

/**
 * STORAGE_KEY is the localStorage key that persists the user's theme preference.
 */
const STORAGE_KEY = 'costa-theme';

/** Reads the persisted theme preference from localStorage, or null if not set. */
export function getPersistedTheme(): Theme | null {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    if (value === 'light' || value === 'dark') {
      return value as Theme;
    }
  } catch {
    // localStorage not available (e.g. private browsing restrictions)
  }
  return null;
}

/** Persists the user's theme choice to localStorage. */
export function persistTheme(theme: Theme): void {
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    // ignore write failures
  }
}

/**
 * Resolves the initial theme: persisted preference → OS preference → 'light'.
 * Never triggers a page reload.
 */
export function resolveInitialTheme(): Theme {
  const persisted = getPersistedTheme();
  if (persisted) return persisted;

  if (typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }
  return 'light';
}

/**
 * Applies a theme to the document root by toggling the `dark` class and
 * injecting the corresponding CSS custom properties inline so the switch is
 * instantaneous (no page reload, no flash).
 */
export function applyTheme(theme: Theme): void {
  const tokens = theme === 'dark' ? darkTokens : lightTokens;
  const root = document.documentElement;

  root.classList.toggle('dark', theme === 'dark');

  for (const [key, value] of Object.entries(tokens)) {
    root.style.setProperty(`--${key}`, value);
  }
}
