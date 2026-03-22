interface ErrorStateProps {
  title: string;
  description: string;
}

export function ErrorState({ title, description }: ErrorStateProps) {
  return (
    <div className='rounded-2xl border border-[var(--status-danger-border)] bg-[var(--status-danger-soft)] px-5 py-6 text-sm'>
      <p className='text-base font-semibold text-[var(--status-danger-foreground)]'>{title}</p>
      <p className='mt-2 leading-6 text-[var(--status-danger-foreground)]/80'>{description}</p>
    </div>
  );
}
