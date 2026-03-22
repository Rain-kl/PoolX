'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { getKernelCapability } from '@/features/capability/api/capability';
import {
  createPortProfile,
  deletePortProfileTemplate,
  deletePortProfile,
  getPortProfile,
  getPortProfiles,
  getPortProfileTemplates,
  getProxyNodeOptions,
  previewPortProfile,
  savePortProfileTemplate,
  saveRuntimeConfig,
  updatePortProfile,
} from '@/features/workspace/api/workspace';
import type {
  LoadBalanceStrategy,
  PortProfilePayload,
  PortProfileProxySettings,
  PortProfileWithNodes,
  PortProfilePreview,
  PortProfileStrategy,
  PortProfileTemplateItem,
} from '@/features/workspace/types';
import {
  CodeBlock,
  DangerButton,
  PrimaryButton,
  ResourceField,
  ResourceInput,
  ResourceSelect,
  SecondaryButton,
} from '@/features/shared/components/resource-primitives';
import { formatDateTime } from '@/lib/utils/date';

const workspaceListQueryKey = ['workspace', 'profiles'] as const;
const defaultMixedPort = 7890;
const defaultSocksPort = 7891;
const defaultHTTPPort = 7892;

const defaultProxySettings: PortProfileProxySettings = {
  strategy_type: 'select',
  test_url: 'https://cp.cloudflare.com/generate_204',
  test_interval_seconds: 300,
  load_balance_strategy: 'consistent-hashing',
  load_balance_lazy: false,
  load_balance_disable_udp: false,
  udp_enabled: true,
  auth_enabled: false,
  auth_username: '',
  auth_password: '',
};

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

const defaultPayload: PortProfilePayload = {
  name: '',
  listen_host: '127.0.0.1',
  mixed_port: defaultMixedPort,
  socks_port: 0,
  http_port: 0,
  proxy_settings: defaultProxySettings,
  include_in_runtime: true,
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
    proxy_settings: {
      ...state.proxy_settings,
      test_url: state.proxy_settings.test_url.trim(),
      auth_username: state.proxy_settings.auth_username.trim(),
      auth_password: state.proxy_settings.auth_password.trim(),
    },
  };
}

function normalizeProxySettings(
  value?: Partial<PortProfileProxySettings> | null,
): PortProfileProxySettings {
  return {
    strategy_type: value?.strategy_type ?? 'select',
    test_url: value?.test_url ?? 'https://cp.cloudflare.com/generate_204',
    test_interval_seconds: value?.test_interval_seconds ?? 300,
    load_balance_strategy: value?.load_balance_strategy ?? 'consistent-hashing',
    load_balance_lazy: value?.load_balance_lazy ?? false,
    load_balance_disable_udp: value?.load_balance_disable_udp ?? false,
    udp_enabled: value?.udp_enabled ?? true,
    auth_enabled: value?.auth_enabled ?? false,
    auth_username: value?.auth_username ?? '',
    auth_password: value?.auth_password ?? '',
  };
}

function toProfilePayload(item: PortProfileWithNodes): PortProfilePayload {
  return {
    name: item.profile.name,
    listen_host: item.profile.listen_host,
    mixed_port: item.profile.mixed_port,
    socks_port: item.profile.socks_port,
    http_port: item.profile.http_port,
    proxy_settings: normalizeProxySettings(item.profile.proxy_settings),
    include_in_runtime: item.profile.include_in_runtime,
    node_ids: item.node_ids,
  };
}

export function WorkspacePage() {
  const queryClient = useQueryClient();
  const [selectedProfileId, setSelectedProfileId] = useState<number | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [payload, setPayload] = useState<PortProfilePayload>(defaultPayload);
  const [nodeSearch, setNodeSearch] = useState('');
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [preview, setPreview] = useState<PortProfilePreview | null>(null);
  const [templateName, setTemplateName] = useState('');
  const usesMixedPort = payload.mixed_port > 0;

  const capabilityQuery = useQuery({
    queryKey: ['capability'],
    queryFn: getKernelCapability,
  });
  const profilesQuery = useQuery({
    queryKey: workspaceListQueryKey,
    queryFn: getPortProfiles,
  });
  const templatesQuery = useQuery({
    queryKey: ['workspace', 'templates'],
    queryFn: getPortProfileTemplates,
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
    if (
      isCreating ||
      !profilesQuery.data ||
      profilesQuery.data.length === 0 ||
      selectedProfileId !== null
    ) {
      return
    }
    setSelectedProfileId(profilesQuery.data[0].profile.id)
  }, [isCreating, profilesQuery.data, selectedProfileId])

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
      proxy_settings: normalizeProxySettings(profile.proxy_settings),
      include_in_runtime: profile.include_in_runtime,
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
      setIsCreating(false)
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

  const toggleRuntimeInclusionMutation = useMutation({
    mutationFn: async ({
      item,
      checked,
    }: {
      item: PortProfileWithNodes;
      checked: boolean;
    }) =>
      updatePortProfile(item.profile.id, {
        ...toProfilePayload(item),
        include_in_runtime: checked,
      }),
    onSuccess: async (_, variables) => {
      const message = variables.checked
        ? '端口配置已加入最终配置。'
        : '端口配置已移出最终配置。';
      setFeedback({ tone: 'success', message });
      await queryClient.invalidateQueries({ queryKey: workspaceListQueryKey });
      await queryClient.invalidateQueries({
        queryKey: ['workspace', 'profile', variables.item.profile.id],
      });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  })

  const previewMutation = useMutation({
    mutationFn: async () => previewPortProfile(toPayload(payload)),
    onSuccess: (result) => {
      setPreview(result)
      setFeedback({ tone: 'success', message: '配置片段预览已生成。' })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const saveRuntimeMutation = useMutation({
    mutationFn: async (id: number) => saveRuntimeConfig(id),
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '当前片段已保存为最新快照。' })
      await queryClient.invalidateQueries({ queryKey: workspaceListQueryKey })
      await queryClient.invalidateQueries({ queryKey: ['workspace', 'profile', selectedProfileId] })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const saveTemplateMutation = useMutation({
    mutationFn: async () => savePortProfileTemplate(templateName, toPayload(payload)),
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '工作台模板已保存。' })
      setTemplateName('')
      await queryClient.invalidateQueries({ queryKey: ['workspace', 'templates'] })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const deleteTemplateMutation = useMutation({
    mutationFn: async (id: number) => deletePortProfileTemplate(id),
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '工作台模板已删除。' })
      await queryClient.invalidateQueries({ queryKey: ['workspace', 'templates'] })
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) })
    },
  })

  const profileList = profilesQuery.data ?? []
  const templateList = templatesQuery.data ?? []
  const nodeOptions = nodeOptionsQuery.data ?? []
  const selectedNodeSet = useMemo(() => new Set(payload.node_ids), [payload.node_ids])
  const capability = capabilityQuery.data

  const handleCreateNew = () => {
    setIsCreating(true)
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
    'load-balance': '按负载均衡策略分散请求，可切换一致性哈希或轮询',
  } satisfies Record<PortProfileStrategy, string>

  const applyTemplate = (item: PortProfileTemplateItem) => {
    setIsCreating(true)
    setSelectedProfileId(null)
    setPreview(null)
    setPayload({
      name: item.template.name,
      listen_host: item.template.listen_host,
      mixed_port: item.template.mixed_port,
      socks_port: item.template.socks_port,
      http_port: item.template.http_port,
      proxy_settings: normalizeProxySettings(item.template.proxy_settings),
      include_in_runtime: item.template.include_in_runtime,
      node_ids: item.node_ids,
    })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="工作台"
        description="创建端口配置、选择节点与策略，并生成后续可合并的工作台配置片段。"
        action={
          <SecondaryButton type="button" onClick={handleCreateNew}>
            新建端口配置
          </SecondaryButton>
        }
      />

      {feedback ? <InlineMessage tone={feedback.tone} message={feedback.message} /> : null}
      {capability ? <InlineMessage tone='info' message={capability.message} /> : null}

      <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <div className="space-y-6">
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
                    setIsCreating(false)
                    setSelectedProfileId(item.profile.id)
                    setFeedback(null)
                  }}
                  className={`w-full rounded-2xl border p-4 text-left transition ${
                    selectedProfileId === item.profile.id
                      ? 'border-[var(--border-strong)] bg-[var(--accent-soft)]'
                      : 'border-[var(--border-default)] bg-[var(--surface-muted)] hover:bg-[var(--surface-base)]'
                  }`}
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                      {item.profile.name}
                    </p>
                    <div
                      className="flex items-center gap-2"
                      onClick={(event) => event.stopPropagation()}
                    >
                      <Switch
                        id={`profile-runtime-${item.profile.id}`}
                        checked={item.profile.include_in_runtime}
                        disabled={toggleRuntimeInclusionMutation.isPending}
                        onCheckedChange={(checked) => {
                          setFeedback(null)
                          toggleRuntimeInclusionMutation.mutate({
                            item,
                            checked,
                          })
                        }}
                      />
                    </div>
                  </div>
                  <div className="mt-2 text-xs leading-6 text-[var(--foreground-secondary)]">
                    <p>Mixed：{item.profile.mixed_port}</p>
                    <p>策略：{item.profile.proxy_settings.strategy_type}</p>
                    <p>节点数：{item.node_ids.length}</p>
                    <p>更新：{formatDateTime(item.profile.updated_at)}</p>
                  </div>
                </button>
              ))}
            </div>
          </AppCard>

          <AppCard
            title="模板"
            description="保存常用工作台配置，后续可直接套用。"
          >
            <div className="space-y-4">
              <div className="flex gap-2">
                <ResourceInput
                  value={templateName}
                  onChange={(event) => setTemplateName(event.target.value)}
                  placeholder="模板名称，留空时使用当前策略组名称"
                />
                <PrimaryButton
                  type="button"
                  onClick={() => {
                    setFeedback(null)
                    saveTemplateMutation.mutate()
                  }}
                  disabled={saveTemplateMutation.isPending || !capability?.supports_templates}
                >
                  {saveTemplateMutation.isPending ? '保存中...' : '保存模板'}
                </PrimaryButton>
              </div>
              {templateList.length === 0 ? (
                <EmptyState title="暂无模板" description="保存一个常用工作台配置作为模板。" />
              ) : (
                <div className="space-y-3">
                  {templateList.map((item) => (
                    <div
                      key={item.template.id}
                      className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div className="space-y-1">
                          <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                            {item.template.name}
                          </p>
                          <p className="text-xs text-[var(--foreground-secondary)]">
                            {item.template.proxy_settings.strategy_type} · 节点 {item.node_ids.length} 个
                          </p>
                        </div>
                        <div className="flex gap-2">
                          <SecondaryButton type="button" onClick={() => applyTemplate(item)}>
                            套用
                          </SecondaryButton>
                          <DangerButton
                            type="button"
                            onClick={() => deleteTemplateMutation.mutate(item.template.id)}
                            disabled={deleteTemplateMutation.isPending}
                          >
                            删除
                          </DangerButton>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </AppCard>
        </div>

        <div className="space-y-6">
          <AppCard
            title="端口配置表单"
            description="保存前可先生成片段预览，确认监听入口、策略和节点绑定是否符合预期。"
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
                  {previewMutation.isPending ? '生成中...' : '生成片段预览'}
                </PrimaryButton>
                {!isCreating && selectedProfileId ? (
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
              <ResourceField label="策略组名称" hint={`${strategyDescription[payload.proxy_settings.strategy_type]}；该名称会作为端口配置名称，且必须唯一。`}>
                <ResourceInput
                  value={payload.name}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      name: event.target.value,
                    }))
                  }
                  placeholder="例如：POOLX"
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
              <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4 xl:col-span-2">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="space-y-1">
                    <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                      端口模式
                    </p>
                    <p className="text-xs text-[var(--foreground-secondary)]">
                      Mixed 与 SOCKS/HTTP 为二选一。开启 Mixed 时，将只保留一个统一入口。
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Switch
                      id="use-mixed-port"
                      checked={usesMixedPort}
                      onCheckedChange={(checked) =>
                        setPayload((current) => ({
                          ...current,
                          mixed_port: checked ? current.mixed_port || defaultMixedPort : 0,
                          socks_port: checked ? 0 : current.socks_port || defaultSocksPort,
                          http_port: checked ? 0 : current.http_port || defaultHTTPPort,
                        }))
                      }
                    />
                    <Label htmlFor="use-mixed-port">启用 Mixed 端口</Label>
                  </div>
                </div>
                <div className="mt-4 grid gap-4 md:grid-cols-2">
                  {usesMixedPort ? (
                    <ResourceField label="Mixed 端口" hint="该端口在工作台配置中必须唯一。">
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
                  ) : (
                    <>
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
                    </>
                  )}
                </div>
              </div>
              <ResourceField label="策略类型">
                <ResourceSelect
                  value={payload.proxy_settings.strategy_type}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      proxy_settings: {
                        ...current.proxy_settings,
                        strategy_type: event.target.value as PortProfileStrategy,
                      },
                    }))
                  }
                >
                  {(capability?.supported_strategies ?? ['select', 'url-test', 'fallback', 'load-balance']).map((strategy) => (
                    <option key={strategy} value={strategy}>{strategy}</option>
                  ))}
                </ResourceSelect>
              </ResourceField>
              <ResourceField label="测试 URL">
                <ResourceInput
                  value={payload.proxy_settings.test_url}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      proxy_settings: {
                        ...current.proxy_settings,
                        test_url: event.target.value,
                      },
                    }))
                  }
                />
              </ResourceField>
              <ResourceField label="测试间隔（秒）">
                <ResourceInput
                  value={String(payload.proxy_settings.test_interval_seconds)}
                  onChange={(event) =>
                    setPayload((current) => ({
                      ...current,
                      proxy_settings: {
                        ...current.proxy_settings,
                        test_interval_seconds: Number.parseInt(event.target.value, 10) || 0,
                      },
                    }))
                  }
                  inputMode="numeric"
                />
              </ResourceField>
              {payload.proxy_settings.strategy_type === 'load-balance' ? (
                <>
                  <ResourceField label="负载均衡策略" hint="一致性哈希会尽量让相同顶级域名走同一节点，轮询会平均分配请求。">
                    <ResourceSelect
                      value={payload.proxy_settings.load_balance_strategy}
                      onChange={(event) =>
                        setPayload((current) => ({
                          ...current,
                          proxy_settings: {
                            ...current.proxy_settings,
                            load_balance_strategy: event.target.value as LoadBalanceStrategy,
                          },
                        }))
                      }
                    >
                      <option value="consistent-hashing">consistent-hashing</option>
                      <option value="round-robin">round-robin</option>
                    </ResourceSelect>
                  </ResourceField>
                  <div className="grid gap-3 lg:grid-cols-2">
                    <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4">
                      <div className="flex items-center justify-between gap-3">
                        <div className="space-y-1">
                          <Label htmlFor="load-balance-lazy">延迟懒加载探测</Label>
                          <p className="text-xs text-[var(--foreground-secondary)]">
                            只在需要时触发健康检查，减少空载时的探测请求。
                          </p>
                        </div>
                        <Switch
                          id="load-balance-lazy"
                          checked={payload.proxy_settings.load_balance_lazy}
                          onCheckedChange={(checked) =>
                            setPayload((current) => ({
                              ...current,
                              proxy_settings: {
                                ...current.proxy_settings,
                                load_balance_lazy: checked,
                              },
                            }))
                          }
                        />
                      </div>
                    </div>
                    <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4">
                      <div className="flex items-center justify-between gap-3">
                        <div className="space-y-1">
                          <Label htmlFor="load-balance-disable-udp">禁用 UDP</Label>
                          <p className="text-xs text-[var(--foreground-secondary)]">
                            为当前负载均衡组关闭 UDP 转发，适合只处理 TCP 的场景。
                          </p>
                        </div>
                        <Switch
                          id="load-balance-disable-udp"
                          checked={payload.proxy_settings.load_balance_disable_udp}
                          onCheckedChange={(checked) =>
                            setPayload((current) => ({
                              ...current,
                              proxy_settings: {
                                ...current.proxy_settings,
                                load_balance_disable_udp: checked,
                              },
                            }))
                          }
                        />
                      </div>
                    </div>
                  </div>
                </>
              ) : null}
              <details className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4 xl:col-span-2">
                <summary className="cursor-pointer list-none text-sm font-semibold text-[var(--foreground-primary)]">
                  高级设置
                </summary>
                <p className="mt-2 text-xs text-[var(--foreground-secondary)]">
                  UDP 默认开启，监听鉴权默认关闭；仅在需要时展开配置。
                </p>
                <div className="mt-4 grid gap-3">
                  <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-base)] p-4">
                    <div className="flex items-center justify-between gap-3">
                      <div className="space-y-1">
                        <Label htmlFor="advanced-udp-enabled">UDP</Label>
                        <p className="text-xs text-[var(--foreground-secondary)]">
                          控制当前监听入口是否允许 UDP 转发。
                        </p>
                      </div>
                      <Switch
                        id="advanced-udp-enabled"
                        checked={payload.proxy_settings.udp_enabled}
                        onCheckedChange={(checked) =>
                          setPayload((current) => ({
                            ...current,
                            proxy_settings: {
                              ...current.proxy_settings,
                              udp_enabled: checked,
                            },
                          }))
                        }
                      />
                    </div>
                  </div>
                  <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-base)] p-4">
                    <div className="flex items-center justify-between gap-3">
                      <div className="space-y-1">
                        <Label htmlFor="advanced-auth-enabled">开启鉴权</Label>
                        <p className="text-xs text-[var(--foreground-secondary)]">
                          开启后会为当前监听入口生成 `users` 凭据列表。
                        </p>
                      </div>
                      <Switch
                        id="advanced-auth-enabled"
                        checked={payload.proxy_settings.auth_enabled}
                        onCheckedChange={(checked) =>
                          setPayload((current) => ({
                            ...current,
                            proxy_settings: {
                              ...current.proxy_settings,
                              auth_enabled: checked,
                              auth_username: checked ? current.proxy_settings.auth_username : '',
                              auth_password: checked ? current.proxy_settings.auth_password : '',
                            },
                          }))
                        }
                      />
                    </div>
                    {payload.proxy_settings.auth_enabled ? (
                      <div className="mt-4 grid gap-4 md:grid-cols-2">
                        <ResourceField label="鉴权用户名">
                          <ResourceInput
                            value={payload.proxy_settings.auth_username}
                            onChange={(event) =>
                              setPayload((current) => ({
                                ...current,
                                proxy_settings: {
                                  ...current.proxy_settings,
                                  auth_username: event.target.value,
                                },
                              }))
                            }
                            placeholder="username1"
                          />
                        </ResourceField>
                        <ResourceField label="鉴权密码">
                          <ResourceInput
                            value={payload.proxy_settings.auth_password}
                            onChange={(event) =>
                              setPayload((current) => ({
                                ...current,
                                proxy_settings: {
                                  ...current.proxy_settings,
                                  auth_password: event.target.value,
                                },
                              }))
                            }
                            placeholder="password1"
                          />
                        </ResourceField>
                      </div>
                    ) : null}
                  </div>
                </div>
              </details>
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
                          来源：{node.source_config_name} · 标签：{node.tags || '未设置'} · 最近状态：{node.last_test_status}
                        </p>
                      </div>
                    </label>
                  ))}
                </div>
              ) : null}
            </div>
          </AppCard>

          <AppCard
            title="配置片段预览"
            description="这里展示的是单个端口配置生成的可合并片段；最终启动文件会在运行阶段由多个片段聚合生成。"
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
                  {saveRuntimeMutation.isPending ? '保存中...' : '保存片段快照'}
                </PrimaryButton>
              ) : null
            }
          >
            {!preview ? (
              <EmptyState
                title="尚未生成片段预览"
                description="填写表单并点击“生成片段预览”后，会在这里看到当前端口配置的合并片段。"
              />
            ) : (
              <div className="space-y-4">
                <div className="grid gap-4 lg:grid-cols-3">
                  <AppCard title="内核" description="当前片段使用的渲染器。">
                    <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                      {preview.kernel_type}
                    </p>
                  </AppCard>
                  <AppCard title="校验和" description="用于后续比较片段内容是否变化。">
                    <p className="break-all text-xs text-[var(--foreground-primary)]">
                      {preview.checksum}
                    </p>
                  </AppCard>
                  <AppCard title="选中节点" description="本次参与片段渲染的节点数量。">
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
