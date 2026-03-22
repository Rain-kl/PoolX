import type { ComponentPropsWithoutRef } from 'react';

import { cn } from '@/lib/utils/cn';

export function Label({
  className,
  ...props
}: ComponentPropsWithoutRef<'label'>) {
  return (
    <label
      className={cn(
        'text-sm font-medium leading-none text-[var(--foreground-primary)] peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
        className,
      )}
      {...props}
    />
  );
}
