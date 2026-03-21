import { AppCard } from '@/components/ui/app-card';
import { StatusBadge } from '@/components/ui/status-badge';

interface FeaturePlaceholderProps {
  title: string;
  description: string;
  milestones: string[];
}

export function FeaturePlaceholder({
  title,
  description,
  milestones,
}: FeaturePlaceholderProps) {
  return (
    <AppCard
      title={title}
      description={description}
      action={<StatusBadge label='阶段 1 骨架完成' variant='success' />}
    >
      <ul className='space-y-3 text-sm leading-6 text-[var(--foreground-secondary)]'>
        {milestones.map((item) => (
          <li key={item} className='flex gap-3'>
            <span className='mt-2 h-2 w-2 shrink-0 rounded-full bg-[var(--brand-primary)]' />
            <span>{item}</span>
          </li>
        ))}
      </ul>
    </AppCard>
  );
}
