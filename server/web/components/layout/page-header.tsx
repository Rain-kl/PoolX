import type {ReactNode} from 'react';

interface PageHeaderProps {
    title: string;
    description?: string;
    action?: ReactNode;
}

export function PageHeader({title, description, action}: PageHeaderProps) {
    return (
        <div className='flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between'>
            <div className='space-y-3'>
                <p className='text-sm font-medium uppercase tracking-[0.24em] text-[var(--brand-primary)]'>PoolX</p>
                <div className='space-y-2'>
                    <h1 className='text-3xl font-semibold tracking-tight text-[var(--foreground-primary)]'>{title}</h1>
                    {description ? (
                        <p className='max-w-3xl text-sm leading-7 text-[var(--foreground-secondary)]'>
                            {description}
                        </p>
                    ) : null}
                </div>
            </div>
            {action ? <div className='flex flex-wrap gap-3'>{action}</div> : null}
        </div>
    );
}
