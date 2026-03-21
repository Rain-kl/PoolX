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
  deleteFile,
  getFiles,
  searchFiles,
  uploadFiles,
} from '@/features/files/api/files';
import type { FileItem } from '@/features/files/types';
import {
  DangerButton,
  PrimaryButton,
  ResourceField,
  ResourceInput,
  SecondaryButton,
  ResourceTextarea,
} from '@/features/shared/components/resource-primitives';

const filesListQueryKey = ['files', 'list'] as const;

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

export function FilesPage() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [searchInput, setSearchInput] = useState('');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [description, setDescription] = useState('');
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);

  const listQuery = useQuery({
    queryKey: [...filesListQueryKey, page],
    queryFn: () => getFiles(page),
    enabled: searchKeyword.length === 0,
  });

  const searchQuery = useQuery({
    queryKey: [...filesListQueryKey, 'search', searchKeyword],
    queryFn: () => searchFiles(searchKeyword),
    enabled: searchKeyword.length > 0,
  });

  const uploadMutation = useMutation({
    mutationFn: () => uploadFiles(selectedFiles, description.trim()),
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '文件已上传。' });
      setSelectedFiles([]);
      setDescription('');
      await queryClient.invalidateQueries({ queryKey: filesListQueryKey });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteFile,
    onSuccess: async () => {
      setFeedback({ tone: 'success', message: '文件已删除。' });
      await queryClient.invalidateQueries({ queryKey: filesListQueryKey });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const activeQuery = searchKeyword ? searchQuery : listQuery;
  const files = useMemo(() => activeQuery.data ?? [], [activeQuery.data]);

  const handleSearch = () => {
    setPage(0);
    setFeedback(null);
    setSearchKeyword(searchInput.trim());
  };

  const handleReset = () => {
    setSearchInput('');
    setSearchKeyword('');
    setPage(0);
    setFeedback(null);
  };

  const handleUpload = () => {
    if (selectedFiles.length === 0) {
      setFeedback({ tone: 'danger', message: '请先选择要上传的文件。' });
      return;
    }
    setFeedback(null);
    uploadMutation.mutate();
  };

  const handleDelete = (file: FileItem) => {
    if (!window.confirm(`确认删除文件“${file.filename}”吗？`)) {
      return;
    }
    setFeedback(null);
    deleteMutation.mutate(file.id);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="文件管理"
        description="上传、检索和删除模板工程中的附件文件。"
      />

      {feedback ? (
        <InlineMessage tone={feedback.tone} message={feedback.message} />
      ) : null}

      <div className="grid gap-6 xl:grid-cols-1">
        <AppCard
          title="上传文件"
          description="支持一次选择多个文件，并为本次上传附加统一描述。"
        >
          <div className="space-y-4">
            <ResourceField label="文件">
              <input
                type="file"
                multiple
                onChange={(event) =>
                  setSelectedFiles(Array.from(event.target.files ?? []))
                }
                className="block w-full rounded-2xl border border-[var(--border-default)] bg-[var(--surface-base)] px-3 py-2 text-sm text-[var(--foreground-primary)]"
              />
            </ResourceField>
            <ResourceField label="描述">
              <ResourceTextarea
                rows={4}
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                placeholder="可选，留空时会写入默认描述。"
              />
            </ResourceField>
            <div className="text-xs text-[var(--foreground-secondary)]">
              {selectedFiles.length > 0
                ? `已选择 ${selectedFiles.length} 个文件`
                : '尚未选择文件'}
            </div>
            <PrimaryButton
              type="button"
              onClick={handleUpload}
              disabled={uploadMutation.isPending}
            >
              {uploadMutation.isPending ? '上传中...' : '上传'}
            </PrimaryButton>
          </div>
        </AppCard>

        <AppCard
          title="文件列表"
          description="支持按文件名、上传者或上传者 ID 搜索。"
        >
          <div className="space-y-4">
            <div className="flex flex-col gap-3 md:flex-row">
              <div className="min-w-0 flex-1">
                <ResourceInput
                  value={searchInput}
                  onChange={(event) => setSearchInput(event.target.value)}
                  placeholder="输入关键字"
                />
              </div>
              <div className="flex gap-2">
                <PrimaryButton type="button" onClick={handleSearch}>
                  搜索
                </PrimaryButton>
                <SecondaryButton type="button" onClick={handleReset}>
                  重置
                </SecondaryButton>
              </div>
            </div>

            {activeQuery.isLoading ? <LoadingState /> : null}
            {activeQuery.isError ? (
              <ErrorState
                title="加载文件失败"
                description={getErrorMessage(activeQuery.error)}
              />
            ) : null}
            {!activeQuery.isLoading && !activeQuery.isError && files.length === 0 ? (
              <EmptyState
                title="暂无文件"
                description="上传后的文件会显示在这里。"
              />
            ) : null}

            {!activeQuery.isLoading && !activeQuery.isError && files.length > 0 ? (
              <div className="space-y-3">
                {files.map((file) => (
                  <div
                    key={file.id}
                    className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] p-4"
                  >
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                      <div className="min-w-0 space-y-1">
                        <p className="truncate text-sm font-semibold text-[var(--foreground-primary)]">
                          {file.filename}
                        </p>
                        <p className="text-sm text-[var(--foreground-secondary)]">
                          {file.description || '无描述信息'}
                        </p>
                        <div className="text-xs text-[var(--foreground-secondary)]">
                          <p>上传者：{file.uploader || '未知'}</p>
                          <p>上传时间：{file.upload_time || '未知'}</p>
                          <p>下载次数：{file.download_counter}</p>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <a
                          href={`/upload/${file.link}`}
                          target="_blank"
                          rel="noreferrer"
                          className="inline-flex min-h-10 items-center justify-center rounded-2xl border border-[var(--border-default)] px-4 text-sm font-medium text-[var(--foreground-primary)] transition hover:bg-[var(--surface-base)]"
                        >
                          下载
                        </a>
                        <DangerButton
                          type="button"
                          onClick={() => handleDelete(file)}
                          disabled={deleteMutation.isPending}
                        >
                          删除
                        </DangerButton>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : null}

            {searchKeyword.length === 0 ? (
              <div className="flex justify-end gap-2">
                <SecondaryButton
                  type="button"
                  onClick={() => setPage((previous) => Math.max(previous - 1, 0))}
                  disabled={page === 0}
                >
                  上一页
                </SecondaryButton>
                <PrimaryButton type="button" onClick={() => setPage((previous) => previous + 1)}>
                  下一页
                </PrimaryButton>
              </div>
            ) : null}
          </div>
        </AppCard>
      </div>
    </div>
  );
}
