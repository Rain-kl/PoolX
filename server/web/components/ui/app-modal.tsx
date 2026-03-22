'use client';

import { useEffect, type ReactNode } from 'react';

import { cn } from '@/lib/utils/cn';

interface AppModalProps {
  isOpen: boolean;
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
  size?: 'md' | 'lg' | 'xl';
}

const sizeClassNameMap = {
  md: 'max-w-2xl',
  lg: 'max-w-4xl',
  xl: 'max-w-5xl',
} satisfies Record<NonNullable<AppModalProps['size']>, string>;

export function AppModal({
  isOpen,
  title,
  description,
  children,
  footer,
  onClose,
  size = 'lg',
}: AppModalProps) {
  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    window.addEventListener('keydown', handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) {
    return null;
  }

  return (
    <div className='fixed inset-0 z-50 flex items-center justify-center bg-slate-950/55 px-4 py-6' role='dialog' aria-modal='true' aria-label={title}>
      <button type='button' className='absolute inset-0 cursor-default' onClick={onClose} aria-label='关闭弹窗' />
      <div
        className={cn(
          'relative flex max-h-[calc(100vh-3rem)] w-full flex-col overflow-hidden rounded-[28px] border border-[var(--border-default)] bg-[var(--surface-panel)] shadow-2xl',
          sizeClassNameMap[size],
        )}
      >
        <div className='flex items-start justify-between gap-4 border-b border-[var(--border-default)] px-6 py-5'>
          <div className='space-y-2'>
            <h2 className='text-xl font-semibold text-[var(--foreground-primary)]'>{title}</h2>
            {description ? (
              <p className='max-w-2xl text-sm leading-6 text-[var(--foreground-secondary)]'>{description}</p>
            ) : null}
          </div>
          <button
            type='button'
            onClick={onClose}
            className='inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-[var(--border-default)] bg-[var(--control-background)] text-lg text-[var(--foreground-primary)] transition hover:bg-[var(--control-background-hover)]'
            aria-label='关闭弹窗'
          >
            ×
          </button>
        </div>
        <div className='min-h-0 flex-1 overflow-y-auto px-6 py-6'>{children}</div>
        {footer ? <div className='border-t border-[var(--border-default)] px-6 py-5'>{footer}</div> : null}
      </div>
    </div>
  );
}
