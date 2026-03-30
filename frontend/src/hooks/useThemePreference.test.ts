import { describe, it, expect, beforeEach, vi } from 'vitest';
import { resolveInitialTheme, getPersistedTheme, persistTheme } from '../app/theme/themeStorage';

// ─── Unit tests for theme storage and bootstrap logic ─────────────────────────
// Pattern: BDD (Given / When / Then) + Triple-A (Arrange / Act / Assert)

describe('resolveInitialTheme', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('returns the persisted theme when one is stored', () => {
    // Arrange
    localStorage.setItem('costa-theme', 'dark');

    // Act
    const result = resolveInitialTheme();

    // Assert
    expect(result).toBe('dark');
  });

  it('falls back to light when nothing is persisted and OS prefers light', () => {
    // Arrange – no stored preference, OS prefers light
    localStorage.clear();
    vi.spyOn(window, 'matchMedia').mockReturnValue({
      matches: false,
      media: '(prefers-color-scheme: dark)',
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as MediaQueryList);

    // Act
    const result = resolveInitialTheme();

    // Assert
    expect(result).toBe('light');
  });

  it('falls back to dark when nothing is persisted and OS prefers dark', () => {
    // Arrange
    localStorage.clear();
    vi.spyOn(window, 'matchMedia').mockReturnValue({
      matches: true,
      media: '(prefers-color-scheme: dark)',
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as MediaQueryList);

    // Act
    const result = resolveInitialTheme();

    // Assert
    expect(result).toBe('dark');
  });
});

describe('getPersistedTheme / persistTheme', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('returns null when nothing is stored', () => {
    expect(getPersistedTheme()).toBeNull();
  });

  it('returns the stored theme after persist', () => {
    // Arrange + Act
    persistTheme('dark');

    // Assert
    expect(getPersistedTheme()).toBe('dark');
  });

  it('overwrites previous value on re-persist', () => {
    // Arrange
    persistTheme('dark');

    // Act
    persistTheme('light');

    // Assert
    expect(getPersistedTheme()).toBe('light');
  });
});
