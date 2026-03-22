'use client';

import type { ButtonHTMLAttributes } from 'react';

import { cn } from '@/lib/utils/cn';

type SwitchProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onChange'> & {
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
};

export function Switch({
  checked = false,
  className,
  disabled,
  onCheckedChange,
  type,
  ...props
}: SwitchProps) {
  return (
    <button
      type={type ?? 'button'}
      role="switch"
      aria-checked={checked}
      data-state={checked ? 'checked' : 'unchecked'}
      disabled={disabled}
      onClick={() => {
        if (disabled) {
          return;
        }
        onCheckedChange?.(!checked);
      }}
      className={cn(
        'peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border border-transparent transition-colors outline-none',
        'focus-visible:ring-2 focus-visible:ring-[var(--brand-primary)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--surface-base)]',
        'disabled:cursor-not-allowed disabled:opacity-50',
        checked
          ? 'bg-[var(--brand-primary)]'
          : 'bg-[var(--surface-elevated)] border-[var(--border-default)]',
        className,
      )}
      {...props}
    >
      <span
        className={cn(
          'pointer-events-none block h-5 w-5 rounded-full bg-white shadow-sm transition-transform',
          checked ? 'translate-x-5' : 'translate-x-0.5',
        )}
      />
    </button>
  );
}
