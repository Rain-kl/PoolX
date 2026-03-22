'use client';

import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import { getKernelCapability } from '@/features/capability/api/capability';
import {
  getRuntimeLogs,
  getRuntimeStatus,
  reloadRuntime,
  startRuntime,
  stopRuntime,
} from '@/features/runtime/api/runtime';
import type { RuntimeLogItem } from '@/features/runtime/types';
import {
  CodeBlock,
  DangerButton,
  PrimaryButton,
  SecondaryButton,
} from '@/features/shared/components/resource-primitives';
import { formatDateTime } from '@/lib/utils/date';

const runtimeStatusQueryKey = ['runtime', 'status'] as const;

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '运行控制请求失败，请稍后重试。';
}

function mergeLogs(existing: RuntimeLogItem[], incoming: RuntimeLogItem[]) {
  if (incoming.length === 0) {
    return existing;
  }
  const result = [...existing];
  const seen = new Set(existing.map((item) => item.seq));
  for (const item of incoming) {
    if (!seen.has(item.seq)) {
      result.push(item);
    }
  }
  return result.slice(-300);
}

export function RuntimePage() {
  const queryClient = useQueryClient();
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [logs, setLogs] = useState<RuntimeLogItem[]>([]);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const capabilityQuery = useQuery({
    queryKey: ['capability'],
    queryFn: getKernelCapability,
  });
  const statusQuery = useQuery({
    queryKey: runtimeStatusQueryKey,
    queryFn: getRuntimeStatus,
    refetchInterval: autoRefresh ? 5000 : false,
  });

  const logsQuery = useQuery({
    queryKey: ['runtime', 'logs', 0],
    queryFn: () => getRuntimeLogs(0, 100),
  });

  useEffect(() => {
    if (logsQuery.data?.items) {
      setLogs(logsQuery.data.items);
    }
  }, [logsQuery.data]);

  const lastSeq = logs.length > 0 ? logs[logs.length - 1].seq : 0;

  useEffect(() => {
    if (!autoRefresh) {
      return;
    }
    const timer = window.setInterval(async () => {
      try {
        const response = await getRuntimeLogs(lastSeq, 100);
        if (response.items.length > 0) {
          setLogs((current) => mergeLogs(current, response.items));
        }
      } catch {
        // keep runtime page usable during transient polling failures
      }
    }, 3000);
    return () => window.clearInterval(timer);
  }, [autoRefresh, lastSeq]);

  const startMutation = useMutation({
    mutationFn: startRuntime,
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: 'Mihomo 已启动。' });
      await queryClient.invalidateQueries({ queryKey: runtimeStatusQueryKey });
      const response = await getRuntimeLogs(0, 100);
      setLogs(response.items);
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const stopMutation = useMutation({
    mutationFn: stopRuntime,
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: 'Mihomo 已停止。' });
      await queryClient.invalidateQueries({ queryKey: runtimeStatusQueryKey });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const reloadMutation = useMutation({
    mutationFn: reloadRuntime,
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: 'Mihomo 已执行热重载。' });
      await queryClient.invalidateQueries({ queryKey: runtimeStatusQueryKey });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const status = statusQuery.data;
  const capability = capabilityQuery.data;
  const listenerSummary = useMemo(() => status?.listeners ?? [], [status?.listeners]);

  if (statusQuery.isLoading) {
    return <LoadingState />;
  }

  if (statusQuery.isError || !status) {
    return (
      <ErrorState
        title="运行状态加载失败"
        description={getErrorMessage(statusQuery.error)}
      />
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="内核"
        description="聚合工作台端口配置后生成最终 Mihomo 配置，并在这里执行启动、停止、热重载和日志查看。"
        action={
          <div className="flex flex-wrap gap-2">
            <PrimaryButton
              type="button"
              onClick={() => {
                setFeedback(null);
                startMutation.mutate();
              }}
              disabled={startMutation.isPending || status.running || !capability?.supports_start}
            >
              {startMutation.isPending ? '启动中...' : '启动'}
            </PrimaryButton>
            <SecondaryButton
              type="button"
              onClick={() => {
                setFeedback(null);
                reloadMutation.mutate();
              }}
              disabled={reloadMutation.isPending || !status.running || !capability?.supports_reload}
            >
              {reloadMutation.isPending ? '重载中...' : '热重载'}
            </SecondaryButton>
            <DangerButton
              type="button"
              onClick={() => {
                setFeedback(null);
                stopMutation.mutate();
              }}
              disabled={stopMutation.isPending || !status.running}
            >
              {stopMutation.isPending ? '停止中...' : '停止'}
            </DangerButton>
            <SecondaryButton type="button" onClick={() => setAutoRefresh((value) => !value)}>
              {autoRefresh ? '暂停自动刷新' : '开启自动刷新'}
            </SecondaryButton>
            <SecondaryButton
              type="button"
              onClick={() => window.open('/zashboard/', '_blank', 'noopener,noreferrer')}
              disabled={!status.running}
            >
              打开 Clash 控制台
            </SecondaryButton>
          </div>
        }
      />

      {feedback ? <InlineMessage tone={feedback.tone} message={feedback.message} /> : null}
      {capability ? <InlineMessage tone={capability.binary_exists ? 'info' : 'danger'} message={capability.message} /> : null}

      <div className="grid gap-4 lg:grid-cols-4">
        <AppCard title="进程状态" description="当前 Mihomo 进程是否正在运行。">
          <p className="text-sm font-semibold text-[var(--foreground-primary)]">
            {status.running ? '运行中' : '已停止'}
          </p>
          <p className="mt-2 text-xs text-[var(--foreground-secondary)]">
            {status.instance.status}
          </p>
        </AppCard>
        <AppCard title="API 健康" description="本地 external-controller 检查结果。">
          <p className="text-sm font-semibold text-[var(--foreground-primary)]">
            {status.api_healthy ? '正常' : '未就绪'}
          </p>
          <p className="mt-2 text-xs text-[var(--foreground-secondary)]">
            {status.api_version || status.instance.controller_address}
          </p>
        </AppCard>
        <AppCard title="端口配置" description="当前参与最终配置聚合的工作台配置数量。">
          <p className="text-sm font-semibold text-[var(--foreground-primary)]">
            {status.profile_count}
          </p>
        </AppCard>
        <AppCard title="监听入口" description="当前最终配置会暴露的监听数量。">
          <p className="text-sm font-semibold text-[var(--foreground-primary)]">
            {status.listener_count}
          </p>
        </AppCard>
      </div>

      <AppCard title="实例详情" description="展示最近动作、PID、配置文件位置和错误摘要。">
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="space-y-2 text-sm text-[var(--foreground-secondary)]">
            <p>PID：{status.instance.pid ?? '未运行'}</p>
            <p>最近动作：{status.instance.last_action || '无'}</p>
            <p>启动时间：{status.instance.last_started_at ? formatDateTime(status.instance.last_started_at) : '未启动'}</p>
            <p>重载时间：{status.instance.last_reloaded_at ? formatDateTime(status.instance.last_reloaded_at) : '未重载'}</p>
          </div>
          <div className="space-y-2 text-sm text-[var(--foreground-secondary)]">
            <p>运行目录：{status.instance.work_dir || '未生成'}</p>
            <p>配置路径：{status.instance.config_path || '未生成'}</p>
            <p>控制地址：{status.instance.controller_address || '未生成'}</p>
            <p>最近错误：{status.instance.last_error || '无'}</p>
          </div>
        </div>
      </AppCard>

      <AppCard title="监听列表" description="每个监听入口对应一个工作台端口配置和策略组。">
        {listenerSummary.length === 0 ? (
          <EmptyState title="暂无监听入口" description="请先在工作台启用至少一个端口配置。" />
        ) : (
          <div className="grid gap-3 lg:grid-cols-2">
            {listenerSummary.map((listener) => (
              <div
                key={listener.name}
                className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
              >
                <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                  {listener.name}
                </p>
                <div className="mt-2 space-y-1 text-sm text-[var(--foreground-secondary)]">
                  <p>协议：{listener.type}</p>
                  <p>监听：{listener.listen}:{listener.port}</p>
                  <p>工作台：{listener.profile_name}</p>
                  <p>策略组：{listener.proxy_group_name}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </AppCard>

      <AppCard title="最终配置预览" description="这里展示的是聚合多个工作台片段后生成的最终 Mihomo 配置。">
        {!status.rendered_config_preview ? (
          <EmptyState title="暂无最终配置" description="启用工作台端口配置后，这里会显示当前聚合结果。" />
        ) : (
          <CodeBlock className="max-h-[480px] overflow-auto">
            {status.rendered_config_preview}
          </CodeBlock>
        )}
      </AppCard>

      <AppCard title="运行日志" description={`当前缓存 ${logs.length} 条运行日志${lastSeq > 0 ? `，最新序号 ${lastSeq}` : ''}`}>
        {logsQuery.isLoading ? <LoadingState /> : null}
        {logs.length === 0 ? (
          <EmptyState title="暂无运行日志" description="启动 Mihomo 后，这里会持续展示 stdout 和 stderr 输出。" />
        ) : (
          <div className="overflow-hidden rounded-[var(--radius-lg)] border border-[var(--border-default)] bg-[var(--surface-elevated)]">
            <div className="max-h-[520px] overflow-y-auto px-4 py-3">
              <div className="space-y-2">
                {logs.map((item) => (
                  <div
                    key={item.seq}
                    className="rounded-2xl border border-transparent px-3 py-2 font-mono text-xs leading-6 text-[var(--foreground-primary)] hover:border-[var(--border-default)] hover:bg-[var(--surface-card)]"
                  >
                    [{formatDateTime(item.created_at)}] [{item.stream}] [{item.level}] {item.message}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </AppCard>
    </div>
  );
}
