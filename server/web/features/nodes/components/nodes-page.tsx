'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import {
  getNodeTestResults,
  getProxyNodes,
  testProxyNodes,
  updateProxyNodeStatus,
} from '@/features/nodes/api/nodes';
import type { ProxyNodeItem } from '@/features/nodes/types';
import {
  PrimaryButton,
  ResourceField,
  ResourceInput,
  ResourceSelect,
  SecondaryButton,
} from '@/features/shared/components/resource-primitives';
import { formatDateTime } from '@/lib/utils/date';

const proxyNodesQueryKey = ['proxy-nodes', 'list'] as const;

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

function getStatusLabel(node: ProxyNodeItem) {
  switch (node.last_test_status) {
    case 'success':
      return '最近测试成功';
    case 'failed':
      return '最近测试失败';
    default:
      return '尚未测试';
  }
}

export function NodesPage() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [keywordInput, setKeywordInput] = useState('');
  const [keyword, setKeyword] = useState('');
  const [enabledFilter, setEnabledFilter] = useState<'all' | 'true' | 'false'>(
    'all',
  );
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [activeNodeId, setActiveNodeId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [testUrl, setTestUrl] = useState('https://cp.cloudflare.com/generate_204');
  const [timeoutMs, setTimeoutMs] = useState('8000');

  const nodesQuery = useQuery({
    queryKey: [...proxyNodesQueryKey, page, keyword, enabledFilter],
    queryFn: () =>
      getProxyNodes({
        page,
        keyword,
        enabled: enabledFilter,
      }),
  });

  const recentResultsQuery = useQuery({
    queryKey: ['proxy-nodes', 'results', activeNodeId],
    queryFn: () => getNodeTestResults(activeNodeId as number),
    enabled: activeNodeId !== null,
  });

  const toggleMutation = useMutation({
    mutationFn: async ({ id, enabled }: { id: number; enabled: boolean }) =>
      updateProxyNodeStatus(id, enabled),
    onSuccess: async (_, variables) => {
      setFeedback({
        tone: 'success',
        message: variables.enabled ? '节点已启用。' : '节点已禁用。',
      });
      await queryClient.invalidateQueries({ queryKey: proxyNodesQueryKey });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const testMutation = useMutation({
    mutationFn: async (nodeIds: number[]) =>
      testProxyNodes({
        nodeIds,
        timeoutMs: Number.parseInt(timeoutMs, 10) || 8000,
        testUrl: testUrl.trim(),
      }),
    onSuccess: async (result) => {
      setFeedback({
        tone: 'success',
        message: `测试已完成，共返回 ${result.length} 条结果。`,
      });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: proxyNodesQueryKey }),
        queryClient.invalidateQueries({ queryKey: ['proxy-nodes', 'results'] }),
      ]);
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const nodes = useMemo(() => nodesQuery.data ?? [], [nodesQuery.data]);

  const selectedIdSet = useMemo(() => new Set(selectedIds), [selectedIds]);

  const handleSearch = () => {
    setPage(0);
    setSelectedIds([]);
    setFeedback(null);
    setKeyword(keywordInput.trim());
  };

  const handleReset = () => {
    setPage(0);
    setKeywordInput('');
    setKeyword('');
    setEnabledFilter('all');
    setSelectedIds([]);
    setFeedback(null);
  };

  const handleToggleSelection = (nodeId: number, checked: boolean) => {
    setSelectedIds((previous) =>
      checked
        ? Array.from(new Set([...previous, nodeId]))
        : previous.filter((item) => item !== nodeId),
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="节点池"
        description="查看已导入节点、按条件筛选、启用或禁用节点，并通过内核发起真实代理请求测试。"
        action={
          <PrimaryButton
            type="button"
            onClick={() => {
              if (selectedIds.length === 0) {
                setFeedback({ tone: 'danger', message: '请先选择至少一个节点。' });
                return;
              }
              setFeedback(null);
              testMutation.mutate(selectedIds);
            }}
            disabled={testMutation.isPending}
          >
            {testMutation.isPending ? '批量测试中...' : '批量测试选中节点'}
          </PrimaryButton>
        }
      />

      {feedback ? (
        <InlineMessage tone={feedback.tone} message={feedback.message} />
      ) : null}

      <AppCard title="筛选条件" description="支持按节点名称、地址和启用状态筛选。">
        <div className="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)_auto]">
          <ResourceField label="关键字">
            <ResourceInput
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              placeholder="输入节点名称、类型或地址"
            />
          </ResourceField>
          <ResourceField label="启用状态">
            <ResourceSelect
              value={enabledFilter}
              onChange={(event) =>
                setEnabledFilter(event.target.value as 'all' | 'true' | 'false')
              }
            >
              <option value="all">全部</option>
              <option value="true">仅启用</option>
              <option value="false">仅禁用</option>
            </ResourceSelect>
          </ResourceField>
          <div className="flex items-end gap-2">
            <PrimaryButton type="button" onClick={handleSearch}>
              查询
            </PrimaryButton>
            <SecondaryButton type="button" onClick={handleReset}>
              重置
            </SecondaryButton>
          </div>
        </div>
      </AppCard>

      <AppCard
        title="测试参数"
        description="服务端会临时拉起 Mihomo 进程，通过本地 mixed-port 代理请求测试 URL，返回真实链路结果。"
      >
        <div className="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(0,180px)]">
          <ResourceField label="测试 URL">
            <ResourceInput
              value={testUrl}
              onChange={(event) => setTestUrl(event.target.value)}
              placeholder="https://cp.cloudflare.com/generate_204"
            />
          </ResourceField>
          <ResourceField label="超时（毫秒）">
            <ResourceInput
              value={timeoutMs}
              onChange={(event) => setTimeoutMs(event.target.value)}
              inputMode="numeric"
              placeholder="8000"
            />
          </ResourceField>
        </div>
      </AppCard>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
        <AppCard
          title="节点列表"
          description="列表字段保留了后续工作台与运行控制会复用的标准化节点信息。"
        >
          <div className="space-y-4">
            {nodesQuery.isLoading ? <LoadingState /> : null}
            {nodesQuery.isError ? (
              <ErrorState
                title="加载节点失败"
                description={getErrorMessage(nodesQuery.error)}
              />
            ) : null}
            {!nodesQuery.isLoading && !nodesQuery.isError && nodes.length === 0 ? (
              <EmptyState
                title="暂无节点"
                description="先去配置导入页上传 YAML 并完成导入。"
              />
            ) : null}

            {!nodesQuery.isLoading && !nodesQuery.isError && nodes.length > 0 ? (
              <div className="space-y-3">
                {nodes.map((node) => (
                  <div
                    key={node.id}
                    className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                  >
                    <div className="flex flex-col gap-4">
                      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                        <div className="flex gap-3">
                          <input
                            type="checkbox"
                            checked={selectedIdSet.has(node.id)}
                            onChange={(event) =>
                              handleToggleSelection(node.id, event.target.checked)
                            }
                            className="mt-1 h-4 w-4 rounded border-[var(--border-default)] accent-[var(--brand-primary)]"
                          />
                          <div className="space-y-1">
                            <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                              {node.name}
                            </p>
                            <p className="text-sm text-[var(--foreground-secondary)]">
                              {node.type.toUpperCase()} · {node.server}:{node.port}
                            </p>
                            <div className="text-xs text-[var(--foreground-secondary)]">
                              <p>来源：{node.source_config_name}</p>
                              <p>{getStatusLabel(node)}</p>
                              <p>
                                最近耗时：
                                {node.last_latency_ms !== undefined
                                  ? ` ${node.last_latency_ms} ms`
                                  : ' 未记录'}
                              </p>
                              <p>
                                最近测试：
                                {node.last_tested_at
                                  ? ` ${formatDateTime(node.last_tested_at)}`
                                  : ' 未执行'}
                              </p>
                              {node.last_test_error ? (
                                <p>最近错误：{node.last_test_error}</p>
                              ) : null}
                            </div>
                          </div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <SecondaryButton
                            type="button"
                            onClick={() => setActiveNodeId(node.id)}
                          >
                            最近记录
                          </SecondaryButton>
                          <SecondaryButton
                            type="button"
                            onClick={() => {
                              setFeedback(null);
                              testMutation.mutate([node.id]);
                            }}
                            disabled={testMutation.isPending}
                          >
                            测试
                          </SecondaryButton>
                          <PrimaryButton
                            type="button"
                            onClick={() =>
                              toggleMutation.mutate({
                                id: node.id,
                                enabled: !node.enabled,
                              })
                            }
                            disabled={toggleMutation.isPending}
                          >
                            {node.enabled ? '禁用' : '启用'}
                          </PrimaryButton>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : null}

            <div className="flex justify-end gap-2">
              <SecondaryButton
                type="button"
                onClick={() => setPage((previous) => Math.max(previous - 1, 0))}
                disabled={page === 0 || nodesQuery.isLoading}
              >
                上一页
              </SecondaryButton>
              <PrimaryButton
                type="button"
                onClick={() => setPage((previous) => previous + 1)}
                disabled={nodesQuery.isLoading || nodes.length === 0}
              >
                下一页
              </PrimaryButton>
            </div>
          </div>
        </AppCard>

        <AppCard
          title="最近测试记录"
          description="用于验证测试结果已落库，可支撑后续的流式进度和历史回看。"
        >
          <div className="space-y-3">
            {activeNodeId === null ? (
              <p className="text-sm text-[var(--foreground-secondary)]">
                从左侧选择一个节点查看最近记录。
              </p>
            ) : recentResultsQuery.isLoading ? (
              <LoadingState />
            ) : recentResultsQuery.isError ? (
              <ErrorState
                title="加载测试记录失败"
                description={getErrorMessage(recentResultsQuery.error)}
              />
            ) : !recentResultsQuery.data || recentResultsQuery.data.length === 0 ? (
              <EmptyState
                title="暂无记录"
                description="对节点执行测试后，结果会展示在这里。"
              />
            ) : (
              recentResultsQuery.data.map((item) => (
                <div
                  key={item.id}
                  className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                >
                  <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                    {item.status === 'success' ? '成功' : '失败'}
                  </p>
                  <div className="mt-2 text-xs leading-6 text-[var(--foreground-secondary)]">
                    <p>目标：{item.dial_address}</p>
                    <p>
                      耗时：
                      {item.latency_ms !== undefined
                        ? ` ${item.latency_ms} ms`
                        : ' 未记录'}
                    </p>
                    <p>开始：{formatDateTime(item.started_at)}</p>
                    <p>结束：{formatDateTime(item.finished_at)}</p>
                    {item.error_message ? <p>错误：{item.error_message}</p> : null}
                  </div>
                </div>
              ))
            )}
          </div>
        </AppCard>
      </div>
    </div>
  );
}
