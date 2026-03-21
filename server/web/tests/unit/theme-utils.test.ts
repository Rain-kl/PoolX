import { describe, expect, it } from 'vitest';

import { isThemeMode, resolveTheme } from '@/lib/theme/theme';

describe('theme utils', () => {
  it('validates theme mode values', () => {
    expect(isThemeMode('light')).toBe(true);
    expect(isThemeMode('dark')).toBe(true);
    expect(isThemeMode('system')).toBe(true);
    expect(isThemeMode('custom')).toBe(false);
  });

  it('resolves system mode using system preference', () => {
    expect(resolveTheme('system', true)).toBe('dark');
    expect(resolveTheme('system', false)).toBe('light');
    expect(resolveTheme('dark', false)).toBe('dark');
    expect(resolveTheme('light', true)).toBe('light');
  });
});
