'use client';

import { cn } from '@/lib/utils/cn';
import { themeModes, type ThemeMode } from '@/lib/theme/theme';
import { useTheme } from '@/components/providers/theme-provider';

interface ThemeToggleProps {
  className?: string;
}

export function ThemeToggle({ className }: ThemeToggleProps) {
  const { themeMode, resolvedTheme, setThemeMode } = useTheme();
  const currentIndex = themeModes.indexOf(themeMode);
  const nextTheme = themeModes[(currentIndex + 1) % themeModes.length] as ThemeMode;

  const icon =
    themeMode === 'light' ? (
      <svg
        className='h-5 w-5'
        viewBox='0 0 24 24'
        fill='none'
        stroke='currentColor'
        strokeWidth='1.8'
        strokeLinecap='round'
        strokeLinejoin='round'
      >
        <circle cx='12' cy='12' r='4' />
        <path d='M12 2.5v2.2M12 19.3v2.2M21.5 12h-2.2M4.7 12H2.5M18.7 5.3l-1.6 1.6M6.9 17.1l-1.6 1.6M18.7 18.7l-1.6-1.6M6.9 6.9 5.3 5.3' />
      </svg>
    ) : themeMode === 'dark' ? (
      <svg
        className='h-5 w-5'
        viewBox='0 0 24 24'
        fill='none'
        stroke='currentColor'
        strokeWidth='1.8'
        strokeLinecap='round'
        strokeLinejoin='round'
      >
        <path d='M20 15.2A7.8 7.8 0 1 1 8.8 4 6.4 6.4 0 0 0 20 15.2Z' />
      </svg>
    ) : (
      <svg
        className='h-5 w-5'
        viewBox='0 0 24 24'
        fill='none'
        stroke='currentColor'
        strokeWidth='1.8'
        strokeLinecap='round'
        strokeLinejoin='round'
      >
        <rect x='3.5' y='4' width='17' height='12' rx='2.5' />
        <path d='M8 20h8M12 16v4' />
      </svg>
    );

  const label =
    themeMode === 'light' ? '浅色模式' : themeMode === 'dark' ? '深色模式' : '跟随系统';

  return (
    <button
      type='button'
      onClick={() => setThemeMode(nextTheme)}
      className={cn(
        'inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-[var(--border-default)] bg-[var(--control-background)] text-[var(--foreground-primary)] transition hover:bg-[var(--control-background-hover)]',
        className,
      )}
      aria-label={`当前${label}，点击切换到${nextTheme === 'light' ? '浅色模式' : nextTheme === 'dark' ? '深色模式' : '跟随系统'}`}
      title={`${label}（当前生效：${resolvedTheme === 'dark' ? '深色' : '浅色'}），点击切换到${nextTheme === 'light' ? '浅色模式' : nextTheme === 'dark' ? '深色模式' : '跟随系统'}`}
    >
      {icon}
    </button>
  );
}
