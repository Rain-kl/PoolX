'use client';

import { useMutation, useQuery } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';

import { useAuth } from '@/components/providers/auth-provider';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { getPublicStatus } from '@/features/auth/api/public';
import {
  createUpgradeLogsWebSocket,
  confirmManualServerUpgrade,
  getLatestRelease,
  parseUpgradeStreamSnapshot,
  upgradeServer,
  uploadServerBinary,
} from '@/features/update/api/update';
import { VersionUpgradeModal } from '@/features/update/components/version-upgrade-modal';
import type {
  LatestReleaseInfo,
  ReleaseChannel,
  UpgradeStreamSnapshot,
  UploadedServerBinaryInfo,
} from '@/features/update/types';
import { publicEnv } from '@/lib/env/public-env';
import { useAppShellStore } from '@/store/app-shell';

export function DashboardTopbar() {
  const router = useRouter();
  const { logout, user } = useAuth();
  const toggleSidebar = useAppShellStore((state) => state.toggleSidebar);
  const isMobileSidebarOpen = useAppShellStore(
    (state) => state.isMobileSidebarOpen,
  );
  const setMobileSidebarOpen = useAppShellStore(
    (state) => state.setMobileSidebarOpen,
  );
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const [isVersionModalOpen, setIsVersionModalOpen] = useState(false);
  const [selectedReleaseChannel, setSelectedReleaseChannel] =
    useState<ReleaseChannel>('stable');
  const [versionFeedback, setVersionFeedback] = useState<string | null>(null);
  const [manualUpgradeStatus, setManualUpgradeStatus] = useState<string | null>(
    null,
  );
  const [manualUpgradeError, setManualUpgradeError] = useState<string | null>(
    null,
  );
  const [uploadedBinary, setUploadedBinary] =
    useState<UploadedServerBinaryInfo | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [upgradeStream, setUpgradeStream] =
    useState<UpgradeStreamSnapshot | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const isRoot = (user?.role ?? 0) >= 100;
  const upgradeStatusPollInterval = 3000;

  const publicStatusQuery = useQuery({
    queryKey: ['public-status'],
    queryFn: getPublicStatus,
  });

  const stableReleaseQuery = useQuery({
    queryKey: ['update', 'latest-release', 'stable'],
    queryFn: () => getLatestRelease('stable'),
    enabled: isRoot,
    refetchInterval: (query) => {
      const release = query.state.data;
      if (isVersionModalOpen && release?.in_progress) {
        return upgradeStatusPollInterval;
      }
      return 60 * 60 * 1000;
    },
  });

  const previewReleaseQuery = useQuery({
    queryKey: ['update', 'latest-release', 'preview'],
    queryFn: () => getLatestRelease('preview'),
    enabled: false,
    refetchInterval: (query) => {
      const release = query.state.data;
      if (isVersionModalOpen && release?.in_progress) {
        return upgradeStatusPollInterval;
      }
      return false;
    },
  });

  const upgradeMutation = useMutation({
    mutationFn: (channel: ReleaseChannel) => upgradeServer(channel),
    onSuccess: (release) => {
      setUploadedBinary(null);
      setManualUpgradeStatus(null);
      setManualUpgradeError(null);
      setVersionFeedback(
        `服务升级任务已启动，目标版本 ${release.tag_name}（${release.channel === 'preview' ? '预览版' : '正式版'}）。页面可能短暂不可用。`,
      );
      void stableReleaseQuery.refetch();
      if (release.channel === 'preview') {
        void previewReleaseQuery.refetch();
      }
    },
    onError: (error) => {
      setVersionFeedback(
        error instanceof Error ? error.message : '升级失败，请稍后重试。',
      );
    },
  });

  const uploadBinaryMutation = useMutation({
    mutationFn: (binary: File) =>
      uploadServerBinary(binary, (progress) => {
        setUploadProgress(progress);
      }),
    onSuccess: (candidate) => {
      setUploadProgress(0);
      setVersionFeedback(null);
      setManualUpgradeError(null);
      setUploadedBinary(candidate);
      setManualUpgradeStatus(candidate.comparison_message);
    },
    onError: (error) => {
      setUploadProgress(0);
      setUploadedBinary(null);
      setManualUpgradeStatus(null);
      setManualUpgradeError(
        error instanceof Error ? error.message : '上传升级包失败，请稍后重试。',
      );
    },
  });

  const confirmManualUpgradeMutation = useMutation({
    mutationFn: confirmManualServerUpgrade,
    onSuccess: (candidate) => {
      setVersionFeedback(null);
      setManualUpgradeError(null);
      setUploadedBinary(candidate);
      setManualUpgradeStatus(
        `手动升级任务已启动，目标版本 ${candidate.detected_version}。页面可能短暂不可用。`,
      );
      void stableReleaseQuery.refetch();
      void previewReleaseQuery.refetch();
    },
    onError: (error) => {
      setManualUpgradeStatus(null);
      setManualUpgradeError(
        error instanceof Error
          ? error.message
          : '确认手动升级失败，请稍后重试。',
      );
    },
  });

  useEffect(() => {
    if (!isVersionModalOpen || !isRoot) {
      setUpgradeStream(null);
      return;
    }

    let closed = false;
    let reconnectTimer: number | null = null;
    let socket: WebSocket | null = null;

    const connect = () => {
      if (closed) {
        return;
      }

      socket = createUpgradeLogsWebSocket();
      if (!socket) {
        return;
      }

      socket.onmessage = (event) => {
        const snapshot = parseUpgradeStreamSnapshot(String(event.data));
        if (snapshot) {
          setUpgradeStream(snapshot);
        }
      };

      socket.onclose = () => {
        if (!closed) {
          reconnectTimer = window.setTimeout(connect, 1500);
        }
      };
    };

    connect();

    return () => {
      closed = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      socket?.close();
    };
  }, [isRoot, isVersionModalOpen]);

  useEffect(() => {
    if (!isUserMenuOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        setIsUserMenuOpen(false);
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsUserMenuOpen(false);
      }
    };

    window.addEventListener('mousedown', handlePointerDown);
    window.addEventListener('keydown', handleEscape);

    return () => {
      window.removeEventListener('mousedown', handlePointerDown);
      window.removeEventListener('keydown', handleEscape);
    };
  }, [isUserMenuOpen]);

  const handleLogout = async () => {
    setIsLoggingOut(true);
    setIsUserMenuOpen(false);
    await logout();
    router.replace('/login');
  };

  const handleSidebarToggle = () => {
    if (window.innerWidth < 1000) {
      setMobileSidebarOpen(!isMobileSidebarOpen);
      return;
    }

    toggleSidebar();
  };

  const handleOpenVersionModal = () => {
    setSelectedReleaseChannel('stable');
    setVersionFeedback(null);
    setManualUpgradeStatus(null);
    setManualUpgradeError(null);
    setIsVersionModalOpen(true);
    if (isRoot) {
      void stableReleaseQuery.refetch();
    }
  };

  const handleUpgrade = () => {
    setVersionFeedback(null);
    setManualUpgradeStatus(null);
    setManualUpgradeError(null);
    upgradeMutation.mutate(selectedReleaseChannel);
  };

  const handleCheckRelease = () => {
    setVersionFeedback(null);
    if (isRoot) {
      if (selectedReleaseChannel === 'preview') {
        void previewReleaseQuery.refetch();
      } else {
        void stableReleaseQuery.refetch();
      }
    }
  };

  const handleReleaseChannelChange = (channel: ReleaseChannel) => {
    setSelectedReleaseChannel(channel);
    setVersionFeedback(null);
  };

  const handleUploadBinary = (binary: File) => {
    setUploadProgress(0);
    setManualUpgradeStatus(null);
    setManualUpgradeError(null);
    uploadBinaryMutation.mutate(binary);
  };

  const handleConfirmManualUpgrade = () => {
    if (!uploadedBinary?.upload_token) {
      setManualUpgradeStatus(null);
      setManualUpgradeError('请先上传并检查升级包。');
      return;
    }
    setVersionFeedback(null);
    setManualUpgradeStatus(null);
    setManualUpgradeError(null);
    confirmManualUpgradeMutation.mutate(uploadedBinary.upload_token);
  };

  const selectedRelease =
    selectedReleaseChannel === 'preview'
      ? previewReleaseQuery.data
      : stableReleaseQuery.data;
  const releaseWithStream = mergeReleaseWithUpgradeStream(
    selectedRelease,
    upgradeStream,
  );
  const selectedReleaseError =
    selectedReleaseChannel === 'preview'
      ? previewReleaseQuery.error
      : stableReleaseQuery.error;
  const isSelectedReleaseError =
    selectedReleaseChannel === 'preview'
      ? previewReleaseQuery.isError
      : stableReleaseQuery.isError;
  const hasUpdate = Boolean(isRoot && stableReleaseQuery.data?.has_update);
  const currentVersion = publicStatusQuery.data?.version || 'unknown';
  const versionLabel = hasUpdate
    ? `版本 ${publicEnv.appVersion} · 可升级`
    : `版本 ${publicEnv.appVersion}`;
  const versionButtonClassName = hasUpdate
    ? 'border-[var(--status-warning-border)] bg-[var(--status-warning-soft)] text-[var(--status-warning-foreground)]'
    : 'border-[var(--border-default)]';
  const versionErrorMessage =
    versionFeedback ||
    (isSelectedReleaseError
      ? selectedReleaseError instanceof Error
        ? selectedReleaseError.message
        : '版本检查失败，请稍后重试。'
      : undefined);
  const manualUpgradeErrorMessage = manualUpgradeError ?? undefined;

  return (
    <>
      <header className="sticky top-0 z-20 border-b border-[var(--border-default)] bg-[var(--surface-panel)]/88 px-4 py-4 backdrop-blur md:px-8">
        <div className="flex items-center justify-between gap-3">
          <button
            type="button"
            onClick={handleSidebarToggle}
            className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-[var(--border-default)] bg-[var(--control-background)] text-lg text-[var(--foreground-primary)] transition hover:bg-[var(--control-background-hover)]"
            aria-label="切换侧边栏"
          >
            ☰
          </button>

          <div className="flex items-center gap-3 text-sm text-[var(--foreground-secondary)]">
            <button
              type="button"
              onClick={handleOpenVersionModal}
              className={[
                'inline-flex rounded-full border px-3 py-1.5 transition',
                versionButtonClassName,
              ].join(' ')}
            >
              <span className="sm:hidden">版本</span>
              <span className="hidden sm:inline">{versionLabel}</span>
            </button>
            <ThemeToggle />
            <div className="relative" ref={menuRef}>
              <button
                type="button"
                onClick={() => setIsUserMenuOpen((value) => !value)}
                className="inline-flex h-11 items-center gap-2 rounded-2xl border border-[var(--border-default)] bg-[var(--control-background)] px-3 text-[var(--foreground-primary)] transition hover:bg-[var(--control-background-hover)]"
                aria-expanded={isUserMenuOpen}
                aria-haspopup="menu"
              >
                <span className="inline-flex h-7 w-7 items-center justify-center rounded-full bg-[var(--accent-soft)] text-xs font-semibold">
                  {(user?.display_name || user?.username || 'U')
                    .slice(0, 1)
                    .toUpperCase()}
                </span>
                <span className="hidden sm:inline">
                  {user?.display_name || user?.username || '用户'}
                </span>
              </button>

              {isUserMenuOpen ? (
                <div className="absolute top-[calc(100%+0.5rem)] right-0 w-52 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-panel)] p-2 shadow-[var(--shadow-lg)]">
                  <div className="rounded-xl px-3 py-2">
                    <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                      {user?.display_name || user?.username || '用户'}
                    </p>
                    {user?.username ? (
                      <p className="mt-1 text-xs text-[var(--foreground-secondary)]">
                        @{user.username}
                      </p>
                    ) : null}
                  </div>
                  <button
                    type="button"
                    onClick={() => void handleLogout()}
                    disabled={isLoggingOut}
                    className="flex w-full items-center rounded-xl px-3 py-2 text-left text-sm text-[var(--status-danger-foreground)] transition hover:bg-[var(--status-danger-soft)] disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isLoggingOut ? '退出中...' : '退出登录'}
                  </button>
                </div>
              ) : null}
            </div>
          </div>
        </div>
      </header>

      <VersionUpgradeModal
        isOpen={isVersionModalOpen}
        onClose={() => setIsVersionModalOpen(false)}
        currentVersion={currentVersion}
        release={releaseWithStream}
        selectedChannel={selectedReleaseChannel}
        uploadedBinary={uploadedBinary}
        isLoading={
          (selectedReleaseChannel === 'preview'
            ? previewReleaseQuery.isLoading && !previewReleaseQuery.data
            : stableReleaseQuery.isLoading && !stableReleaseQuery.data) &&
          isRoot
        }
        releaseErrorMessage={versionErrorMessage}
        manualStatusMessage={manualUpgradeStatus ?? undefined}
        manualErrorMessage={manualUpgradeErrorMessage}
        canUpgrade={isRoot}
        isChecking={
          selectedReleaseChannel === 'preview'
            ? previewReleaseQuery.isFetching
            : stableReleaseQuery.isFetching
        }
        isUpgrading={upgradeMutation.isPending}
        isUploadingBinary={uploadBinaryMutation.isPending}
        uploadProgress={uploadProgress}
        isConfirmingManualUpgrade={confirmManualUpgradeMutation.isPending}
        onChannelChange={handleReleaseChannelChange}
        onCheck={handleCheckRelease}
        onUpgrade={handleUpgrade}
        onUploadBinary={handleUploadBinary}
        onConfirmManualUpgrade={handleConfirmManualUpgrade}
      />
    </>
  );
}

function mergeReleaseWithUpgradeStream(
  release: LatestReleaseInfo | null | undefined,
  stream: UpgradeStreamSnapshot | null,
) {
  if (!release || !stream) {
    return release;
  }

  return {
    ...release,
    in_progress: stream.in_progress,
    upgrade_status: stream.upgrade_status,
    upgrade_logs: stream.upgrade_logs,
  };
}
