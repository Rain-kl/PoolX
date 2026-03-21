'use client';

import { useMutation } from '@tanstack/react-query';
import { useMemo, useState } from 'react';

import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import {
  importSourceConfig,
  parseSourceConfig,
  testParsedNodes,
} from '@/features/import/api/source-import';
import type {
  ParsedNodePreview,
  ParsedNodeTestResult,
  SourceParseResult,
} from '@/features/import/types';
import {
  PrimaryButton,
  ResourceField,
  ResourceInput,
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
  const [previewNodes, setPreviewNodes] = useState<ParsedNodePreview[]>([]);
  const [selectedFingerprints, setSelectedFingerprints] = useState<string[]>([]);
  const [testResults, setTestResults] = useState<ParsedNodeTestResult[]>([]);
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [testUrl, setTestUrl] = useState('https://cp.cloudflare.com/generate_204');
  const [timeoutMs, setTimeoutMs] = useState('8000');

  const importableNodes = useMemo(
    () => previewNodes.filter((node) => node.duplicate_scope === 'none'),
    [previewNodes],
  );
  const selectedSet = useMemo(
    () => new Set(selectedFingerprints),
    [selectedFingerprints],
  );

  const parseMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFile) {
        throw new Error('请先选择 YAML 文件。');
      }
      return parseSourceConfig(selectedFile);
    },
    onSuccess: (result) => {
      setParseResult(result);
      setPreviewNodes(result.nodes);
      setSelectedFingerprints(
        result.nodes
          .filter((node) => node.duplicate_scope === 'none')
          .map((node) => node.fingerprint),
      );
      setTestResults([]);
      setFeedback({
        tone: 'success',
        message: `解析完成，识别到 ${result.summary.valid_nodes} 个有效节点。`,
      });
    },
    onError: (error) => {
      setParseResult(null);
      setPreviewNodes([]);
      setSelectedFingerprints([]);
      setTestResults([]);
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const importMutation = useMutation({
    mutationFn: async () => {
      if (!parseResult) {
        throw new Error('请先完成解析。');
      }
      return importSourceConfig(parseResult.source_config.id, selectedFingerprints);
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

  const testMutation = useMutation({
    mutationFn: async () => {
      if (!parseResult) {
        throw new Error('请先完成解析。');
      }
      if (selectedFingerprints.length === 0) {
        throw new Error('请先选择要测速的节点。');
      }
      return testParsedNodes({
        sourceConfigId: parseResult.source_config.id,
        fingerprints: selectedFingerprints,
        timeoutMs: Number.parseInt(timeoutMs, 10) || 8000,
        testUrl: testUrl.trim(),
      });
    },
    onSuccess: (result) => {
      setTestResults(result);
      setFeedback({
        tone: 'success',
        message: `测速完成，共返回 ${result.length} 条结果。`,
      });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const summary = useMemo(() => {
    if (!parseResult) {
      return null;
    }

    return {
      total_nodes: previewNodes.length + parseResult.errors.length,
      valid_nodes: previewNodes.length,
      invalid_nodes: parseResult.errors.length,
      duplicate_nodes: previewNodes.filter((node) => node.duplicate_scope !== 'none').length,
      importable_nodes: importableNodes.length,
    };
  }, [importableNodes.length, parseResult, previewNodes]);

  const handleToggleNode = (fingerprint: string, checked: boolean) => {
    setSelectedFingerprints((previous) =>
      checked
        ? Array.from(new Set([...previous, fingerprint]))
        : previous.filter((item) => item !== fingerprint),
    );
  };

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
                setPreviewNodes([]);
                setSelectedFingerprints([]);
                setTestResults([]);
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

      {parseResult && summary ? (
        <>
          <div className="grid gap-4 lg:grid-cols-5">
            <AppCard title="总节点数" description="含有效节点与解析失败条目。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {summary.total_nodes}
              </p>
            </AppCard>
            <AppCard title="有效节点" description="满足最小字段要求的可识别节点。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {summary.valid_nodes}
              </p>
            </AppCard>
            <AppCard title="错误条目" description="无法标准化的配置项。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {summary.invalid_nodes}
              </p>
            </AppCard>
            <AppCard title="重复节点" description="批内重复或库内已存在。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {summary.duplicate_nodes}
              </p>
            </AppCard>
            <AppCard title="可导入" description="确认后会写入节点池的节点数。">
              <p className="text-2xl font-semibold text-[var(--foreground-primary)]">
                {summary.importable_nodes}
              </p>
            </AppCard>
          </div>

          <AppCard
            title="导入确认"
            description={`当前导入记录 #${parseResult.source_config.id}，文件 ${parseResult.source_config.filename}。当前已选择 ${selectedFingerprints.length} 个节点。`}
          >
            <div className="flex flex-wrap gap-3">
              <PrimaryButton
                type="button"
                onClick={() => {
                  setFeedback(null);
                  testMutation.mutate();
                }}
                disabled={testMutation.isPending || selectedFingerprints.length === 0}
              >
                {testMutation.isPending ? '测速中...' : '测速选中节点'}
              </PrimaryButton>
              <PrimaryButton
                type="button"
                onClick={() => {
                  setFeedback(null);
                  importMutation.mutate();
                }}
                disabled={
                  importMutation.isPending ||
                  selectedFingerprints.length === 0
                }
              >
                {importMutation.isPending ? '导入中...' : '导入节点池'}
              </PrimaryButton>
              <SecondaryButton
                type="button"
                onClick={() =>
                  setSelectedFingerprints(
                    previewNodes
                      .filter((node) => node.duplicate_scope === 'none')
                      .map((node) => node.fingerprint),
                  )
                }
              >
                选择全部可导入
              </SecondaryButton>
              <SecondaryButton
                type="button"
                onClick={() => setSelectedFingerprints([])}
              >
                清空选择
              </SecondaryButton>
              <SecondaryButton
                type="button"
                onClick={() => {
                  if (selectedFingerprints.length === 0) {
                    setFeedback({ tone: 'danger', message: '请先选择要删除的预览节点。' });
                    return;
                  }
                  setPreviewNodes((previous) =>
                    previous.filter(
                      (node) => !selectedSet.has(node.fingerprint),
                    ),
                  );
                  setSelectedFingerprints([]);
                  setTestResults([]);
                  setFeedback({
                    tone: 'info',
                    message: '已从当前预览中移除选中节点，本次导入和测速都不会再包含它们。',
                  });
                }}
              >
                删除选中预览
              </SecondaryButton>
            </div>
          </AppCard>

          <AppCard
            title="测速参数"
            description="会在导入前临时拉起内核，对当前选择的解析节点做真实代理请求测试。"
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

          <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
            <AppCard
              title="节点预览"
              description="默认最多展示前 100 个解析结果，支持勾选后批量测速或从本次预览移除。"
            >
              <div className="space-y-3">
                {previewNodes.map((node, index) => (
                  <div
                    key={`${node.fingerprint}-${index}`}
                    className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                  >
                    <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
                      <div className="flex gap-3">
                        <input
                          type="checkbox"
                          checked={selectedSet.has(node.fingerprint)}
                          disabled={node.duplicate_scope !== 'none'}
                          onChange={(event) =>
                            handleToggleNode(node.fingerprint, event.target.checked)
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
                        </div>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                          {selectedSet.has(node.fingerprint) ? '已选择' : '未选择'}
                        </p>
                        <span className="inline-flex min-h-9 items-center rounded-full border border-[var(--border-default)] px-3 text-xs text-[var(--foreground-secondary)]">
                          {duplicateScopeLabel(node.duplicate_scope)}
                        </span>
                      </div>
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

            <AppCard
              title="测速结果"
              description="导入前测速结果仅用于当前预览，不会写入节点池最近状态。"
            >
              <div className="space-y-3">
                {testResults.length === 0 ? (
                  <p className="text-sm text-[var(--foreground-secondary)]">
                    选择节点后点击“测速选中节点”即可查看结果。
                  </p>
                ) : (
                  testResults.map((item, index) => (
                    <div
                      key={`${item.node_name}-${index}`}
                      className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                    >
                      <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                        {item.node_name}
                      </p>
                      <div className="mt-2 text-xs leading-6 text-[var(--foreground-secondary)]">
                        <p>结果：{item.status === 'success' ? '成功' : '失败'}</p>
                        <p>目标：{item.dial_address}</p>
                        <p>
                          耗时：
                          {item.latency_ms !== undefined
                            ? ` ${item.latency_ms} ms`
                            : ' 未记录'}
                        </p>
                        {item.error_message ? <p>错误：{item.error_message}</p> : null}
                      </div>
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
