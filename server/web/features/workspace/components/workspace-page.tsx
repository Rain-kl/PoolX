'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import {
  createPortProfile,
  deletePortProfile,
  getPortProfile,
  getPortProfiles,
  getProxyNodeOptions,
  previewPortProfile,
  saveRuntimeConfig,
  updatePortProfile,
} from '@/features/workspace/api/workspace';
import type {
  PortProfilePayload,
  PortProfilePreview,
  PortProfileStrategy,
} from '@/features/workspace/types';
import {
  CodeBlock,
  DangerButton,
  PrimaryButton,
  ResourceField,
  ResourceInput,
  ResourceSelect,
  SecondaryButton,
  ToggleField,
} from '@/features/shared/components/resource-primitives';
import { formatDateTime } from '@/lib/utils/date';

const workspaceListQueryKey = ['workspace', 'profiles'] as const;

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

const defaultPayload: PortProfilePayload = {
  name: '',
  listen_host: '127.0.0.1',
  mixed_port: 7890,
  socks_port: 0,
  http_port: 0,
  strategy_type: 'select',
  strategy_group_name: 'POOLX',
  test_url: 'https://cp.cloudflare.com/generate_204',
  test_interval_seconds: 300,
  enabled: true,
  node_ids: [],
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

function toPayload(state: PortProfilePayload) {
  return {
    ...state,
    name: state.name.trim(),
    listen_host: state.listen_host.trim(),
    strategy_group_name: state.strategy_group_name.trim(),
    test_url: state.test_url.trim(),
  };
}

export function WorkspacePage() {
  const queryClient = useQueryClient();
  const [selectedProfileId, setSelectedProfileId] = useState<number | null>(null);
  const [payload, setPayload] = useState<PortProfilePayload>(defaultPayload);
  const [nodeSearch, setNodeSearch] = useState('');
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [preview, setPreview] = useState<PortProfilePreview | null>(null);

  const profilesQuery = useQuery({
    queryKey: workspaceListQueryKey,
    queryFn: getPortProfiles,
  });

  const detailQuery = useQuery({
    queryKey: ['workspace', 'profile', selectedProfileId],
    queryFn: () => getPortProfile(selectedProfileId as number),
    enabled: selectedProfileId !== null,
  });

  const nodeOptionsQuery = useQuery({
    queryKey: ['workspace', 'node-options', nodeSearch],
    queryFn: () => getProxyNodeOptions(nodeSearch),
  });

  useEffect(() => {
    if (!profilesQuery.data || profilesQuery.data.length === 0 || selectedProfileId !== null) {
      return
    }
    setSelectedProfileId(profilesQuery.data[0].profile.id)
  }, [profilesQuery.data, selectedProfileId])

  useEffect(() => {
    if (!detailQuery.data) {
      return
    }
    const { profile, node_ids } = detailQuery.data
    setPayload({
      name: profile.name,
      listen_host: profile.listen_host,
      mixed_port: profile.mixed_port,
      socks_port: profile.socks_port,
      http_port: profile.http_port,
      strategy_type: profile.strategy_type,
      strategy_group_name: profile.strategy_group_name,
      test_url: profile.test_url,
      test_interval_seconds: profile.test_interval_seconds,
      enabled: profile.enabled,
      node_ids,
    })
    if (detailQuery.data.runtime) {
      setPreview({
        profile,
        node_ids,
        nodes: detailQuery.data.nodes,
        kernel_type: detailQuery.data.runtime.kernel_type,
        checksum: detailQuery.data.runtime.checksum,
        content: detailQuery.data.runtime.rendered_config,
      })
    } else {
      setPreview(null)
    }
  }, [detailQuery.data])

  const createMutation = useMutation({
    mutationFn: async () => createPortProfile(toPayload(payload)),
    onSuccess: async (result) => {
      setFeedback({ tone: 'success', message: '端口配置已创建。' })
      setSelectedProfileId(result.profile.id)
      await queryClient.invalidateQueries({ queryKey: workspaceListQueryKey })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const updateMutation = useMutation({
    mutationFn: async (id: number) => updatePortProfile(id, toPayload(payload)),
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '端口配置已保存。' })
      await queryClient.invalidateQueries({ queryKey: workspaceListQueryKey })
      await queryClient.invalidateQueries({ queryKey: ['workspace', 'profile', selectedProfileId] })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => deletePortProfile(id),
    onSuccess: async (_, id) => {
      setFeedback({ tone: 'success', message: '端口配置已删除。' })
      setPreview(null)
      setSelectedProfileId((current) => (current === id ? null : current))
      setPayload(defaultPayload)
      await queryClient.invalidateQueries({ queryKey: workspaceListQueryKey })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const previewMutation = useMutation({
    mutationFn: async () => previewPortProfile(toPayload(payload)),
    onSuccess: (result) => {
      setPreview(result)
      setFeedback({ tone: 'success', message: '配置预览已生成。' })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const saveRuntimeMutation = useMutation({
    mutationFn: async (id: number) => saveRuntimeConfig(id),
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '最新预览已保存为运行快照。' })
      await queryClient.invalidateQueries({ queryKey: workspaceListQueryKey })
      await queryClient.invalidateQueries({ queryKey: ['workspace', 'profile', selectedProfileId] })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const profileList = profilesQuery.data ?? []
  const nodeOptions = nodeOptionsQuery.data ?? []
  const selectedNodeSet = useMemo(() => new Set(payload.node_ids), [payload.node_ids])

  const handleCreateNew = () => {
    setSelectedProfileId(null)
    setPayload(defaultPayload)
    setPreview(null)
    setFeedback(null)
  }

  const handleToggleNode = (nodeId: number, checked: boolean) => {
    setPayload((current) => ({
      ...current,
      node_ids: checked
        ? Array.from(new Set([...current.node_ids, nodeId]))
        : current.node_ids.filter((item) => item !== nodeId),
    }))
  }

  const strategyDescription = {
    select: '手动选择固定出口',
    'url-test': '按 URL 延迟自动选择',
    fallback: '按优先级与可用性自动切换',
    'load-balance': '按一致性哈希分散请求',
  } satisfies Record<PortProfileStrategy, string>

  return (
    <div className="space-y-6">
      <PageHeader
        title="工作台"
        description="创建端口配置、选择节点与策略，并生成 Mihomo 配置预览。"
        action={
          <SecondaryButton type="button" onClick={handleCreateNew}>
            新建端口配置
          </SecondaryButton>
        }
      />

      {feedback ? <InlineMessage tone={feedback.tone} message={feedback.message} /> : null}

      <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <AppCard
          title="端口配置"
          description="左侧维护当前工作台的端口配置集合。"
        >
          <div className="space-y-3">
            {profilesQuery.isLoading ? <LoadingState /> : null}
            {profilesQuery.isError ? (
              <ErrorState
                title="加载端口配置失败"
                description={getErrorMessage(profilesQuery.error)}
              />
            ) : null}
            {!profilesQuery.isLoading && !profilesQuery.isError && profileList.length === 0 ? (
              <EmptyState title="暂无端口配置" description="点击右上角“新建端口配置”开始。" />
            ) : null}
            {profileList.map((item) => (
              <button
                key={item.profile.id}
                type="button"
                onClick={() => {
                  setSelectedProfileId(item.profile.id)
                  setFeedback(null)
                }}
                className={`w-full rounded-2xl border p-4 text-left transition ${
                  selectedProfileId === item.profile.id
                    ? 'border-[var(--border-strong)] bg-[var(--accent-soft)]'
                    : 'border-[var(--border-default)] bg-[var(--surface-muted)] hover:bg-[var(--surface-base)]'
                }`}
              >
                <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                  {item.profile.name}
                </p>
                <div className="mt-2 text-xs leading-6 text-[var(--foreground-secondary)]">
                  <p>Mixed：{item.profile.mixed_port}</p>
                  <p>策略：{item.profile.strategy_type}</p>
                  <p>节点数：{item.node_ids.length}</p>
                  <p>更新：{formatDateTime(item.profile.updated_at)}</p>
                </div>
              </button>
            ))}
          </div>
        </AppCard>

        <div className="space-y-6">
          <AppCard
            title="端口配置表单"
            description="保存前可先生成预览，确认策略和端口是否符合预期。"
            action={
              <div className="flex flex-wrap gap-2">
                <PrimaryButton
                  type="button"
                  onClick={() => {
                    setFeedback(null)
                    previewMutation.mutate()
                  }}
                  disabled={previewMutation.isPending}
                >
                  {previewMutation.isPending ? '生成中...' : '生成预览'}
                </PrimaryButton>
                {selectedProfileId ? (
                  <>
                    <PrimaryButton
                      type="button"
                      onClick={() => {
                        setFeedback(null)
                        updateMutation.mutate(selectedProfileId)
                      }}
                      disabled={updateMutation.isPending}
                    >
                      {updateMutation.isPending ? '保存中...' : '保存配置'}
                    </PrimaryButton>
                    <DangerButton
                      type="button"
                      onClick={() => {
                        if (!window.confirm('确认删除当前端口配置吗？')) {
                          return
                        }
                        setFeedback(null)
                        deleteMutation.mutate(selectedProfileId)
                      }}
                      disabled={deleteMutation.isPending}
                    >
                      删除
                    </DangerButton>
                  </>
                ) : (
                  <PrimaryButton
                    type="button"
                    onClick={() => {
                      setFeedback(null)
                      createMutation.mutate()
                    }}
                    disabled={createMutation.isPending}
                  >
                    {createMutation.isPending ? '创建中...' : '创建配置'}
                  </PrimaryButton>
                )}
              </div>
            }
          >
            <div className="grid gap-4 xl:grid-cols-2">
              <ResourceField label="名称">
                <ResourceInput
                  value={payload.name}
                  onChange={(event) =>
                    setPayload((current) => ({ ...current, name: event.target.value }))
                  }
                  placeholder="例如：默认工作台"
                />
              </ResourceField>
              <ResourceField label="监听地址">
                <ResourceInput
                  value={payload.listen_host}
                  onChange={(event) =>
                    setPayload((current) => ({ ...current, listen_host: event.target.value }))
                  }
                  placeholder="127.0.0.1"
                />
              </ResourceField>
              <ResourceField label="Mixed 端口">
                <ResourceInput
                  value={String(payload.mixed_port)}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      mixed_port: Number.parseInt(event.target.value, 10) || 0,
                    }))
                  }
                  inputMode="numeric"
                />
              </ResourceField>
              <ResourceField label="SOCKS 端口">
                <ResourceInput
                  value={String(payload.socks_port)}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      socks_port: Number.parseInt(event.target.value, 10) || 0,
                    }))
                  }
                  inputMode="numeric"
                />
              </ResourceField>
              <ResourceField label="HTTP 端口">
                <ResourceInput
                  value={String(payload.http_port)}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      http_port: Number.parseInt(event.target.value, 10) || 0,
                    }))
                  }
                  inputMode="numeric"
                />
              </ResourceField>
              <ResourceField label="策略类型">
                <ResourceSelect
                  value={payload.strategy_type}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      strategy_type: event.target.value as PortProfileStrategy,
                    }))
                  }
                >
                  <option value="select">select</option>
                  <option value="url-test">url-test</option>
                  <option value="fallback">fallback</option>
                  <option value="load-balance">load-balance</option>
                </ResourceSelect>
              </ResourceField>
              <ResourceField label="策略组名称" hint={strategyDescription[payload.strategy_type]}>
                <ResourceInput
                  value={payload.strategy_group_name}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      strategy_group_name: event.target.value,
                    }))
                  }
                />
              </ResourceField>
              <ResourceField label="测试 URL">
                <ResourceInput
                  value={payload.test_url}
                  onChange={(event) =>
                    setPayload((current) => ({ ...current, test_url: event.target.value }))
                  }
                />
              </ResourceField>
              <ResourceField label="测试间隔（秒）">
                <ResourceInput
                  value={String(payload.test_interval_seconds)}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      test_interval_seconds: Number.parseInt(event.target.value, 10) || 0,
                    }))
                  }
                  inputMode="numeric"
                />
              </ResourceField>
              <div className="xl:col-span-2">
                <ToggleField
                  label="启用此配置"
                  checked={payload.enabled}
                  onChange={(checked) =>
                    setPayload((current) => ({ ...current, enabled: checked }))
                  }
                />
              </div>
            </div>
          </AppCard>

          <AppCard
            title="节点选择器"
            description="从已导入节点中选择工作台要使用的节点，支持关键字过滤。"
          >
            <div className="space-y-4">
              <ResourceField label="搜索节点">
                <ResourceInput
                  value={nodeSearch}
                  onChange={(event) => setNodeSearch(event.target.value)}
                  placeholder="输入节点名称、类型或地址"
                />
              </ResourceField>
              {nodeOptionsQuery.isLoading ? <LoadingState /> : null}
              {nodeOptionsQuery.isError ? (
                <ErrorState
                  title="加载节点选项失败"
                  description={getErrorMessage(nodeOptionsQuery.error)}
                />
              ) : null}
              {!nodeOptionsQuery.isLoading && !nodeOptionsQuery.isError ? (
                <div className="grid gap-3 lg:grid-cols-2">
                  {nodeOptions.map((node) => (
                    <label
                      key={node.id}
                      className="flex gap-3 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                    >
                      <input
                        type="checkbox"
                        checked={selectedNodeSet.has(node.id)}
                        onChange={(event) => handleToggleNode(node.id, event.target.checked)}
                        className="mt-1 h-4 w-4 rounded border-[var(--border-default)] accent-[var(--brand-primary)]"
                      />
                      <div className="min-w-0 space-y-1">
                        <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                          {node.name}
                        </p>
                        <p className="text-sm text-[var(--foreground-secondary)]">
                          {node.type.toUpperCase()} · {node.server}:{node.port}
                        </p>
                        <p className="text-xs text-[var(--foreground-secondary)]">
                          来源：{node.source_config_name} · 最近状态：{node.last_test_status}
                        </p>
                      </div>
                    </label>
                  ))}
                </div>
              ) : null}
            </div>
          </AppCard>

          <AppCard
            title="配置预览"
            description="生成后会返回 YAML 预览和校验和，可选择保存为最新运行快照。"
            action={
              selectedProfileId && preview ? (
                <PrimaryButton
                  type="button"
                  onClick={() => {
                    setFeedback(null)
                    saveRuntimeMutation.mutate(selectedProfileId)
                  }}
                  disabled={saveRuntimeMutation.isPending}
                >
                  {saveRuntimeMutation.isPending ? '保存中...' : '保存为运行快照'}
                </PrimaryButton>
              ) : null
            }
          >
            {!preview ? (
              <EmptyState
                title="尚未生成预览"
                description="填写表单并点击“生成预览”后，会在这里看到渲染结果。"
              />
            ) : (
              <div className="space-y-4">
                <div className="grid gap-4 lg:grid-cols-3">
                  <AppCard title="内核" description="当前预览使用的渲染器。">
                    <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                      {preview.kernel_type}
                    </p>
                  </AppCard>
                  <AppCard title="校验和" description="用于后续比较配置是否变化。">
                    <p className="break-all text-xs text-[var(--foreground-primary)]">
                      {preview.checksum}
                    </p>
                  </AppCard>
                  <AppCard title="选中节点" description="本次参与渲染的节点数量。">
                    <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                      {preview.node_ids.length}
                    </p>
                  </AppCard>
                </div>
                <CodeBlock className="max-h-[480px] overflow-auto">
                  {preview.content}
                </CodeBlock>
              </div>
            )}
          </AppCard>
        </div>
      </div>
    </div>
  )
}

