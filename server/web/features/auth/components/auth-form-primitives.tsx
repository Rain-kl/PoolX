import type { ButtonHTMLAttributes, InputHTMLAttributes, ReactNode } from 'react';

import { cn } from '@/lib/utils/cn';

export function AuthFormField({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: ReactNode;
}) {
  return (
    <label className='block space-y-2'>
      <span className='text-sm font-medium text-[var(--foreground-primary)]'>{label}</span>
      {children}
      {hint ? <span className='block text-xs text-[var(--foreground-secondary)]'>{hint}</span> : null}
    </label>
  );
}

export function AuthInput(props: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={cn(
        'w-full rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3 text-sm text-[var(--foreground-primary)] outline-none transition placeholder:text-[var(--foreground-muted)] focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--accent-soft)]',
        props.className,
      )}
    />
  );
}

export function AuthButton({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      className={cn(
        'inline-flex w-full items-center justify-center rounded-2xl bg-[var(--brand-primary)] px-4 py-3 text-sm font-medium text-[var(--foreground-inverse)] transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60',
        className,
      )}
    />
  );
}

export function SecondaryButton({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      className={cn(
        'inline-flex items-center justify-center rounded-2xl border border-[var(--border-default)] bg-[var(--control-background)] px-4 py-3 text-sm font-medium text-[var(--foreground-primary)] transition hover:bg-[var(--control-background-hover)] disabled:cursor-not-allowed disabled:opacity-60',
        className,
      )}
    />
  );
}
