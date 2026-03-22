'use client';

import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { marked } from 'marked';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { LoadingState } from '@/components/feedback/loading-state';
import { AppCard } from '@/components/ui/app-card';
import { StatusBadge } from '@/components/ui/status-badge';
import { getPublicStatus } from '@/features/auth/api/public';
import { getAboutContent } from '@/features/settings/api/settings';
import { formatDateTime } from '@/lib/utils/date';

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

export function AboutPage() {
  const aboutQuery = useQuery({
    queryKey: ['about'],
    queryFn: getAboutContent,
  });

  const statusQuery = useQuery({
    queryKey: ['public-status'],
    queryFn: getPublicStatus,
  });

  if (aboutQuery.isLoading || statusQuery.isLoading) {
    return <LoadingState />;
  }

  if (aboutQuery.isError) {
    return <ErrorState title='关于页加载失败' description={getErrorMessage(aboutQuery.error)} />;
  }

  if (statusQuery.isError) {
    return <ErrorState title='系统状态加载失败' description={getErrorMessage(statusQuery.error)} />;
  }

  const aboutContent = aboutQuery.data?.trim() ?? '';
  const status = statusQuery.data;

  if (!status) {
    return <EmptyState title='暂无系统信息' description='未能获取当前系统的公开状态信息。' />;
  }

  return (
    <div className='space-y-6'>
      <AppCard
        title='关于 PoolX'
        description='公开展示当前系统的简介、版本信息与项目入口。'
        action={<StatusBadge label={status.version || 'dev'} variant='info' />}
      >
        <div className='grid gap-4 md:grid-cols-3'>
          <div className='rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] px-4 py-4'>
            <p className='text-xs uppercase tracking-[0.2em] text-[var(--foreground-muted)]'>系统名称</p>
            <p className='mt-2 text-sm font-medium text-[var(--foreground-primary)]'>{status.system_name || 'PoolX'}</p>
          </div>
          <div className='rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] px-4 py-4'>
            <p className='text-xs uppercase tracking-[0.2em] text-[var(--foreground-muted)]'>Server 启动时间</p>
            <p className='mt-2 text-sm font-medium text-[var(--foreground-primary)]'>
              {formatDateTime(new Date(status.start_time * 1000))}
            </p>
          </div>
          <div className='rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] px-4 py-4'>
            <p className='text-xs uppercase tracking-[0.2em] text-[var(--foreground-muted)]'>项目仓库</p>
            <a
              href='https://github.com/Rain-kl/PoolX'
              target='_blank'
              rel='noreferrer'
              className='mt-2 block text-sm font-medium text-[var(--brand-primary)] transition hover:opacity-80'
            >
              项目仓库
            </a>
          </div>
        </div>
      </AppCard>

      {aboutContent ? (
        <AppCard title='项目介绍' description='以下内容由系统设置中的“关于内容”维护。'>
          <div
            className='prose prose-sm max-w-none text-[var(--foreground-primary)] [&_a]:text-[var(--brand-primary)]'
            dangerouslySetInnerHTML={{ __html: marked.parse(aboutContent) as string }}
          />
        </AppCard>
      ) : (
        <EmptyState
          title='尚未配置关于内容'
          description='可在设置页中维护关于内容，这里会自动同步展示。'
        />
      )}

      <div className='flex flex-wrap gap-3 text-sm'>
        <Link href='/login' className='text-[var(--brand-primary)] transition hover:opacity-80'>
          前往登录
        </Link>
      </div>
    </div>
  );
}
