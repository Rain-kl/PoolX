'use client';

import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';

import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import {
  importSourceConfig,
  parseSourceConfig,
} from '@/features/import/api/source-import';
import type { SourceParseResult } from '@/features/import/types';
import {
  PrimaryButton,
  ResourceField,
  SecondaryButton,
} from '@/features/shared/components/resource-primitives';

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

function duplicateScopeLabel(scope: string) {
  switch (scope) {
    case 'batch':
      return '批内重复';
    case 'database':
      return '库内已存在';
    default:
      return '可导入';
  }
}

export function SourceImportPage() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [parseResult, setParseResult] = useState<SourceParseResult | null>(null);
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);

  const parseMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFile) {
        throw new Error('请先选择 YAML 文件。');
      }
      return parseSourceConfig(selectedFile);
    },
    onSuccess: (result) => {
      setParseResult(result);
      setFeedback({
        tone: 'success',
        message: `解析完成，识别到 ${result.summary.valid_nodes} 个有效节点。`,
      });
    },
    onError: (error) => {
      setParseResult(null);
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const importMutation = useMutation({
    mutationFn: async () => {
      if (!parseResult) {
        throw new Error('请先完成解析。');
      }
      return importSourceConfig(parseResult.source_config.id);
    },
    onSuccess: (result) => {
      setFeedback({
        tone: 'success',
        message: `导入完成，新增 ${result.imported_nodes} 个节点，跳过 ${result.skipped_nodes} 个重复节点。`,
      });
      setParseResult((previous) =>
        previous
          ? {
              ...previous,
              source_config: {
                ...previous.source_config,
                status: 'imported',
                imported_nodes: result.imported_nodes,
                duplicate_nodes: result.skipped_nodes,
              },
            }
          : previous,
      );
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="配置导入"
        description="上传 Clash/Mihomo YAML，服务端会完成解析、标准化和去重预检，再将确认后的节点导入节点池。"
      />

      {feedback ? (
        <InlineMessage tone={feedback.tone} message={feedback.message} />
      ) : null}

      <AppCard
        title="上传并解析"
        description="当前最小版本支持从 `proxies` 列表读取节点，并按节点指纹做批内去重与库内去重预检。"
      >
        <div className="space-y-4">
          <ResourceField label="YAML 文件">
            <input
              type="file"
              accept=".yaml,.yml,text/yaml,text/x-yaml,application/x-yaml"
              onChange={(event) =>
                setSelectedFile(event.target.files?.[0] ?? null)
              }
              className="block w-full rounded-2xl border border-[var(--border-default)] bg-[var(--surface-base)] px-3 py-2 text-sm text-[var(--foreground-primary)]"
            />
          </ResourceField>

          <div className="flex flex-wrap gap-3">
            <PrimaryButton
              type="button"
              onClick={() => {
                setFeedback(null);
                parseMutation.mutate();
              }}
              disabled={parseMutation.isPending}
            >
              {parseMutation.isPending ? '解析中...' : '开始解析'}
            </PrimaryButton>
            <SecondaryButton
              type="button"
              onClick={() => {
                setSelectedFile(null);
                setParseResult(null);
                setFeedback(null);
              }}
            >
              清空
            </SecondaryButton>
          </div>
        </div>
      </AppCard>

      {parseMutation.isError ? (
        <ErrorState
          title="解析失败"
          description={getErrorMessage(parseMutation.error)}
        />
      ) : null}

      {parseResult ? (
        <>
          <div className="grid gap-4 lg:grid-cols-5">
            <AppCard title="总节点数" description="含有效节点与解析失败条目。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {parseResult.summary.total_nodes}
              </p>
            </AppCard>
            <AppCard title="有效节点" description="满足最小字段要求的可识别节点。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {parseResult.summary.valid_nodes}
              </p>
            </AppCard>
            <AppCard title="错误条目" description="无法标准化的配置项。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {parseResult.summary.invalid_nodes}
              </p>
            </AppCard>
            <AppCard title="重复节点" description="批内重复或库内已存在。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {parseResult.summary.duplicate_nodes}
              </p>
            </AppCard>
            <AppCard title="可导入" description="确认后会写入节点池的节点数。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {parseResult.summary.importable_nodes}
              </p>
            </AppCard>
          </div>

          <AppCard
            title="导入确认"
            description={`当前导入记录 #${parseResult.source_config.id}，文件 ${parseResult.source_config.filename}`}
          >
            <div className="flex flex-wrap gap-3">
              <PrimaryButton
                type="button"
                onClick={() => {
                  setFeedback(null);
                  importMutation.mutate();
                }}
                disabled={
                  importMutation.isPending ||
                  parseResult.summary.importable_nodes === 0
                }
              >
                {importMutation.isPending ? '导入中...' : '导入节点池'}
              </PrimaryButton>
              <SecondaryButton
                type="button"
                onClick={() => setFeedback(null)}
              >
                保留当前预览
              </SecondaryButton>
            </div>
          </AppCard>

          <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
            <AppCard
              title="节点预览"
              description="默认最多展示前 100 个解析结果。"
            >
              <div className="space-y-3">
                {parseResult.nodes.map((node, index) => (
                  <div
                    key={`${node.name}-${node.server}-${node.port}-${index}`}
                    className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                  >
                    <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
                      <div className="space-y-1">
                        <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                          {node.name}
                        </p>
                        <p className="text-sm text-[var(--foreground-secondary)]">
                          {node.type.toUpperCase()} · {node.server}:{node.port}
                        </p>
                      </div>
                      <span className="inline-flex min-h-9 items-center rounded-full border border-[var(--border-default)] px-3 text-xs text-[var(--foreground-secondary)]">
                        {duplicateScopeLabel(node.duplicate_scope)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </AppCard>

            <AppCard
              title="解析错误"
              description="服务端会保留错误信息，便于后续补充更细的兼容规则。"
            >
              <div className="space-y-3">
                {parseResult.errors.length === 0 ? (
                  <p className="text-sm text-[var(--foreground-secondary)]">
                    当前没有解析错误。
                  </p>
                ) : (
                  parseResult.errors.map((issue, index) => (
                    <div
                      key={`${issue.index}-${index}`}
                      className="rounded-2xl border border-[var(--status-danger-border)] bg-[var(--status-danger-soft)] p-4 text-sm text-[var(--status-danger-foreground)]"
                    >
                      <p className="font-medium">
                        条目 #{issue.index + 1}
                        {issue.name ? ` · ${issue.name}` : ''}
                      </p>
                      <p className="mt-1">{issue.message}</p>
                    </div>
                  ))
                )}
              </div>
            </AppCard>
          </div>
        </>
      ) : null}
    </div>
  );
}

