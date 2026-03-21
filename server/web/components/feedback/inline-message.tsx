import { cn } from '@/lib/utils/cn';

type InlineMessageTone = 'info' | 'success' | 'danger';

const toneClasses: Record<InlineMessageTone, string> = {
  info: 'border-[var(--status-info-border)] bg-[var(--status-info-soft)] text-[var(--status-info-foreground)]',
  success:
    'border-[var(--status-success-border)] bg-[var(--status-success-soft)] text-[var(--status-success-foreground)]',
  danger:
    'border-[var(--status-danger-border)] bg-[var(--status-danger-soft)] text-[var(--status-danger-foreground)]',
};

interface InlineMessageProps {
  tone?: InlineMessageTone;
  message: string;
  className?: string;
}

export function InlineMessage({
  tone = 'info',
  message,
  className,
}: InlineMessageProps) {
  return (
    <div
      className={cn(
        'rounded-2xl border px-4 py-3 text-sm leading-6',
        toneClasses[tone],
        className,
      )}
    >
      {message}
    </div>
  );
}
