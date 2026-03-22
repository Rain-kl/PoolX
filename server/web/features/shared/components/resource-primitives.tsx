import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from 'react';

import { cn } from '@/lib/utils/cn';

interface ResourceFieldProps {
  label: string;
  hint?: string;
  error?: string;
  className?: string;
  tooltip?: string;
  children: ReactNode;
}

function FieldTooltip({ content }: { content: string }) {
  return (
    <span className="group/tooltip relative inline-flex">
      <span className="inline-flex h-5 w-5 items-center justify-center rounded-full border border-[var(--border-default)] bg-[var(--surface-elevated)] text-[11px] font-semibold text-[var(--foreground-secondary)] transition group-hover/tooltip:border-[var(--border-strong)] group-hover/tooltip:text-[var(--foreground-primary)]">
        ?
      </span>
      <span className="pointer-events-none absolute top-full left-1/2 z-20 mt-2 w-64 -translate-x-1/2 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-panel)] px-3 py-2 text-xs leading-5 text-[var(--foreground-secondary)] opacity-0 shadow-[var(--shadow-soft)] transition-opacity delay-700 duration-200 group-hover/tooltip:opacity-100">
        {content}
      </span>
    </span>
  );
}

export function ResourceField({
  label,
  hint,
  error,
  className,
  tooltip,
  children,
}: ResourceFieldProps) {
  return (
    <label className={cn('block space-y-2', className)}>
      <span className="flex items-center gap-2 text-sm font-medium text-[var(--foreground-primary)]">
        <span>{label}</span>
        {tooltip ? <FieldTooltip content={tooltip} /> : null}
      </span>
      {children}
      {error ? (
        <span className="block text-xs text-[var(--status-danger-foreground)]">
          {error}
        </span>
      ) : hint ? (
        <span className="block text-xs text-[var(--foreground-secondary)]">
          {hint}
        </span>
      ) : null}
    </label>
  );
}

export function ResourceInput(props: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={cn(
        'w-full rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3 text-sm text-[var(--foreground-primary)] transition outline-none placeholder:text-[var(--foreground-muted)] focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--accent-soft)] disabled:cursor-not-allowed disabled:opacity-60',
        props.className,
      )}
    />
  );
}

export function ResourceTextarea(
  props: TextareaHTMLAttributes<HTMLTextAreaElement>,
) {
  return (
    <textarea
      {...props}
      className={cn(
        'min-h-28 w-full rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3 text-sm text-[var(--foreground-primary)] transition outline-none placeholder:text-[var(--foreground-muted)] focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--accent-soft)] disabled:cursor-not-allowed disabled:opacity-60',
        props.className,
      )}
    />
  );
}

export function ResourceSelect(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      {...props}
      className={cn(
        'w-full rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3 text-sm text-[var(--foreground-primary)] transition outline-none focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--accent-soft)] disabled:cursor-not-allowed disabled:opacity-60',
        props.className,
      )}
    />
  );
}

function baseButtonClassName(className?: string) {
  return cn(
    'inline-flex items-center justify-center rounded-2xl px-4 py-3 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-60',
    className,
  );
}

export function PrimaryButton({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      className={baseButtonClassName(
        cn(
          'bg-[var(--brand-primary)] text-[var(--foreground-inverse)] hover:opacity-90',
          className,
        ),
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
      className={baseButtonClassName(
        cn(
          'border border-[var(--border-default)] bg-[var(--control-background)] text-[var(--foreground-primary)] hover:bg-[var(--control-background-hover)]',
          className,
        ),
      )}
    />
  );
}

export function DangerButton({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      className={baseButtonClassName(
        cn(
          'border border-[var(--status-danger-border)] bg-[var(--status-danger-soft)] text-[var(--status-danger-foreground)] hover:opacity-90',
          className,
        ),
      )}
    />
  );
}

interface ToggleFieldProps {
  label: string;
  description?: string;
  checked: boolean;
  disabled?: boolean;
  tooltip?: string;
  onChange: (checked: boolean) => void;
}

export function ToggleField({
  label,
  description,
  checked,
  disabled,
  tooltip,
  onChange,
}: ToggleFieldProps) {
  return (
    <label className="flex self-start cursor-pointer items-start gap-3 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3">
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        className="mt-1 h-4 w-4 rounded border-[var(--border-default)] accent-[var(--brand-primary)]"
      />
      <span className="flex flex-col">
        <span className="flex items-center gap-2 text-sm font-medium text-[var(--foreground-primary)]">
          <span>{label}</span>
          {tooltip ? <FieldTooltip content={tooltip} /> : null}
        </span>
        {description ? (
          <span className="mt-1 block text-xs leading-5 text-[var(--foreground-secondary)]">
            {description}
          </span>
        ) : null}
      </span>
    </label>
  );
}

export function CodeBlock({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <pre
      className={cn(
        'overflow-x-auto rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4 text-xs leading-6 text-[var(--foreground-primary)]',
        className,
      )}
    >
      {children}
    </pre>
  );
}
