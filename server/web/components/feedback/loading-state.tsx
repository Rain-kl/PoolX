import { cn } from '@/lib/utils/cn';

interface LoadingStateProps {
  className?: string;
}

export function LoadingState({ className }: LoadingStateProps) {
  return (
    <div className={cn('space-y-3', className)}>
      <div className='h-4 w-32 animate-pulse rounded bg-[var(--control-background-hover)]' />
      <div className='h-24 animate-pulse rounded-2xl bg-[var(--surface-muted)]' />
      <div className='h-24 animate-pulse rounded-2xl bg-[var(--surface-muted)]' />
    </div>
  );
}
