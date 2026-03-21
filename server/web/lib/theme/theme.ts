export const THEME_STORAGE_KEY = 'poolx-theme-mode';
export const THEME_MEDIA_QUERY = '(prefers-color-scheme: dark)';

export const themeModes = ['light', 'dark', 'system'] as const;

export type ThemeMode = (typeof themeModes)[number];
export type ResolvedTheme = Exclude<ThemeMode, 'system'>;

export function isThemeMode(value: string | null | undefined): value is ThemeMode {
  return themeModes.includes(value as ThemeMode);
}

export function resolveTheme(
  mode: ThemeMode,
  systemPrefersDark: boolean,
): ResolvedTheme {
  if (mode === 'system') {
    return systemPrefersDark ? 'dark' : 'light';
  }

  return mode;
}

export function getThemeInitScript() {
  return `(() => {
    const storageKey = '${THEME_STORAGE_KEY}';
    const mediaQuery = '${THEME_MEDIA_QUERY}';
    const isThemeMode = (value) => value === 'light' || value === 'dark' || value === 'system';
    const storedValue = window.localStorage.getItem(storageKey);
    const themeMode = isThemeMode(storedValue) ? storedValue : 'system';
    const resolvedTheme = themeMode === 'system'
      ? (window.matchMedia(mediaQuery).matches ? 'dark' : 'light')
      : themeMode;
    const root = document.documentElement;

    root.dataset.themeMode = themeMode;
    root.dataset.theme = resolvedTheme;
    root.style.colorScheme = resolvedTheme;
  })();`;
}
