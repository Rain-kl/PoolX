'use client';

import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import { getLogs } from '@/features/logs/api/logs';
import type { AppLogItem, LogClassification } from '@/features/logs/types';
import {
  formatLogLine,
  getClassificationMeta,
  matchesLogKeyword,
} from '@/features/logs/lib/log-display';
import {
  PrimaryButton,
  ResourceInput,
  ResourceSelect,
  SecondaryButton,
} from '@/features/shared/components/resource-primitives';

const AUTO_REFRESH_INTERVAL_MS = 5000;

type ClassificationFilter = LogClassification | 'all';

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '日志请求失败，请稍后重试。';
}

function mergeLogs(existing: AppLogItem[], incoming: AppLogItem[]) {
  if (incoming.length === 0) {
    return existing;
  }

  const items = [...existing];
  const seen = new Set(existing.map((item) => item.id));
  for (const item of incoming) {
    if (!seen.has(item.id)) {
      items.push(item);
      seen.add(item.id);
    }
  }
  return items.sort((left, right) => left.id - right.id);
}

export function LogsPage() {
  const [classification, setClassification] = useState<ClassificationFilter>('all');
  const [search, setSearch] = useState('');
  const [logs, setLogs] = useState<AppLogItem[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [feedback, setFeedback] = useState<string>('');

  const logsQuery = useQuery({
    queryKey: ['app-logs', classification],
    queryFn: () => getLogs({ classification, limit: 100 }),
  });

  useEffect(() => {
    if (logsQuery.data?.items) {
      setLogs(logsQuery.data.items);
    }
  }, [logsQuery.data]);

  const lastLogId = logs.length > 0 ? logs[logs.length - 1].id : 0;

  useEffect(() => {
    if (isPaused) {
      return;
    }

    const timer = window.setInterval(async () => {
      try {
        const response = await getLogs({
          classification,
          afterId: lastLogId,
          limit: 100,
        });
        if (response.items.length > 0) {
          setLogs((current) => mergeLogs(current, response.items));
        }
      } catch {
        // keep the page usable even if periodic sync fails once
      }
    }, AUTO_REFRESH_INTERVAL_MS);

    return () => window.clearInterval(timer);
  }, [classification, isPaused, lastLogId]);

  const filteredLogs = useMemo(() => {
    const keyword = search.trim().toLowerCase();
    if (!keyword) {
      return logs;
    }

    return logs.filter((item) => matchesLogKeyword(item, keyword));
  }, [logs, search]);

  const handleManualRefresh = async () => {
    setFeedback('');
    try {
      const response = await getLogs({
        classification,
        afterId: lastLogId,
        limit: 100,
      });
      if (response.items.length > 0) {
        setLogs((current) => mergeLogs(current, response.items));
        setFeedback(`已同步 ${response.items.length} 条新日志。`);
        return;
      }
      setFeedback('当前没有新的日志。');
    } catch (error) {
      setFeedback(getErrorMessage(error));
    }
  };

  const handleCopy = async () => {
    const content = filteredLogs
      .map(formatLogLine)
      .join('\n');

    if (!content) {
      setFeedback('当前没有可复制的日志。');
      return;
    }

    await navigator.clipboard.writeText(content);
    setFeedback('日志内容已复制到剪贴板。');
  };

  const handleDownload = () => {
    if (filteredLogs.length === 0) {
      setFeedback('当前没有可下载的日志。');
      return;
    }

    const content = filteredLogs
      .map(formatLogLine)
      .join('\n');
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `poolx-logs-${Date.now()}.txt`;
    link.click();
    URL.revokeObjectURL(url);
    setFeedback('日志文件已生成。');
  };

  if (logsQuery.isLoading) {
    return <LoadingState />;
  }

  if (logsQuery.isError) {
    return <ErrorState title='日志加载失败' description={getErrorMessage(logsQuery.error)} />;
  }

  return (
    <div className='space-y-6'>
      <PageHeader
        title='系统日志'
        description='默认展示最近 100 条。'
        action={
          <div className='flex flex-wrap gap-2'>
            <SecondaryButton type='button' onClick={() => setIsPaused((value) => !value)}>
              {isPaused ? '继续同步' : '暂停同步'}
            </SecondaryButton>
            <SecondaryButton type='button' onClick={() => void handleCopy()}>
              复制
            </SecondaryButton>
            <SecondaryButton type='button' onClick={handleDownload}>
              下载日志
            </SecondaryButton>
            <PrimaryButton type='button' onClick={() => void handleManualRefresh()}>
              获取新日志
            </PrimaryButton>
          </div>
        }
      />

      {feedback ? <InlineMessage tone='info' message={feedback} /> : null}

      <AppCard
        title='日志筛选'
        description='支持按分类切换、日志级别关键字搜索和增量刷新。'
      >
        <div className='grid gap-4 lg:grid-cols-2'>
          <ResourceSelect
            value={classification}
            onChange={(event) => setClassification(event.target.value as ClassificationFilter)}
          >
            <option value='all'>全部日志</option>
            <option value='system'>系统日志</option>
            <option value='business'>业务日志</option>
            <option value='security'>安全日志</option>
          </ResourceSelect>
          <ResourceInput
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder='搜索日志时间、分类、级别或消息内容'
          />
        </div>
      </AppCard>

      <AppCard
        title='日志流'
        description={`当前展示 ${filteredLogs.length} 条日志${lastLogId > 0 ? `，最新日志 ID 为 ${lastLogId}` : ''}`}
      >
        {filteredLogs.length === 0 ? (
          <EmptyState title='暂无日志' description='当前筛选条件下还没有日志数据。' />
        ) : (
          <div className='overflow-hidden rounded-[var(--radius-lg)] border border-[var(--border-default)] bg-[var(--surface-elevated)]'>
            <div className='max-h-[620px] overflow-y-auto px-4 py-3'>
              <div className='space-y-2'>
                {filteredLogs.map((item) => {
                  const meta = getClassificationMeta(item.classification);
                  const logLine = formatLogLine(item);

                  return (
                    <div
                      key={item.id}
                      className='flex gap-3 rounded-2xl border border-transparent px-3 py-2 transition hover:border-[var(--border-default)] hover:bg-[var(--surface-card)]'
                    >
                      <span className={`mt-1 h-5 w-1.5 shrink-0 rounded-full ${meta.dot}`} />
                      <div className='min-w-0 flex-1'>
                        <p
                          className='overflow-hidden text-ellipsis whitespace-nowrap font-mono text-xs text-[var(--foreground-primary)]'
                          title={logLine}
                        >
                          {logLine}
                        </p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </AppCard>
    </div>
  );
}
