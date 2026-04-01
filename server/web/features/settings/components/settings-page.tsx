'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ChangeEvent } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { TurnstileWidget } from '@/components/forms/turnstile-widget';
import { useAuth } from '@/components/providers/auth-provider';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import { StatusBadge } from '@/components/ui/status-badge';
import { sendEmailVerification } from '@/features/auth/api/auth';
import { getPublicStatus } from '@/features/auth/api/public';
import {
  bindEmail,
  bindWeChat,
  downloadMihomoBinary,
  generateAccessToken,
  getOptions,
  getSettingsProfile,
  inspectMihomoBinary,
  previewGeoIP,
  uploadMihomoBinary,
  updateOption,
  updateSelf,
} from '@/features/settings/api/settings';
import type {
  GeoIPPreviewResult,
  KernelBinaryInstallResult,
  OptionItem,
  UpdateSelfPayload,
} from '@/features/settings/types';
import {
  CodeBlock,
  PrimaryButton,
  ResourceField,
  ResourceInput,
  ResourceSelect,
  ResourceTextarea,
  SecondaryButton,
  ToggleField,
} from '@/features/shared/components/resource-primitives';
import { formatDateTime } from '@/lib/utils/date';

const settingsQueryKey = ['settings', 'options'] as const;
const defaultServerUpdateRepo = 'Rain-kl/PoolX';

const defaultSystemFields = {
  ServerAddress: '',
  ServerUpdateRepo: defaultServerUpdateRepo,
  KernelType: 'mihomo',
  MihomoBinaryPath: '',
  MihomoBinaryVersion: '',
  MihomoBinarySource: '',
  ClashAllowLAN: false,
  ClashExternalController: '127.0.0.1:19090',
  ClashMode: 'rule',
  ClashSecret: '3ebc195c9fbe81c01eb9299e3c6bf644',
  NodeTestDefaultURL: 'https://cp.cloudflare.com/generate_204',
  NodeTestDefaultTimeoutMS: '8000',
  GeoIPProvider: 'disabled',
  PasswordLoginEnabled: true,
  PasswordRegisterEnabled: true,
  EmailVerificationEnabled: false,
  GitHubOAuthEnabled: false,
  WeChatAuthEnabled: false,
  TurnstileCheckEnabled: false,
  RegisterEnabled: false,
  SMTPServer: '',
  SMTPPort: '587',
  SMTPAccount: '',
  SMTPToken: '',
  GitHubClientId: '',
  GitHubClientSecret: '',
  WeChatServerAddress: '',
  WeChatServerToken: '',
  WeChatAccountQRCodeImageURL: '',
  TurnstileSiteKey: '',
  TurnstileSecretKey: '',
};

const defaultOperationFields = {
  GlobalApiRateLimitNum: '300',
  GlobalApiRateLimitDuration: '180',
  GlobalWebRateLimitNum: '300',
  GlobalWebRateLimitDuration: '180',
  UploadRateLimitNum: '50',
  UploadRateLimitDuration: '60',
  DownloadRateLimitNum: '50',
  DownloadRateLimitDuration: '60',
  CriticalRateLimitNum: '100',
  CriticalRateLimitDuration: '1200',
  ServerAddress: '',
};

const defaultOtherFields = {
  Notice: '',
  SystemName: '',
  HomePageLink: '',
  About: '',
  Footer: '',
};

const defaultProfileFields: UpdateSelfPayload = {
  username: '',
  display_name: '',
  password: '',
};

type FeedbackState = {
  tone: 'info' | 'success' | 'danger';
  message: string;
};

type SettingsTab = 'personal' | 'operation' | 'clash' | 'system' | 'other';

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

function optionsToMap(options: OptionItem[] | undefined) {
  return (options ?? []).reduce<Record<string, string>>(
    (accumulator, option) => {
      accumulator[option.key] = option.value;
      return accumulator;
    },
    {},
  );
}

function toBoolean(value: string | undefined, fallback: boolean) {
  if (value === undefined) {
    return fallback;
  }

  return value === 'true';
}

function normalizeServerUrl(value: string) {
  return value.trim().replace(/\/+$/, '');
}

function getBrowserOrigin() {
  if (typeof window === 'undefined') {
    return '';
  }

  return normalizeServerUrl(window.location.origin);
}

function defaultMihomoInstallPath() {
  if (typeof navigator !== 'undefined' && navigator.userAgent.includes('Windows')) {
    return '.\\mihomo.exe';
  }
  return './mihomo';
}

function formatSecondsLabel(value: string) {
  const seconds = Number.parseInt(value, 10);
  if (Number.isNaN(seconds)) {
    return value;
  }

  if (seconds >= 3600 && seconds % 3600 === 0) {
    return `${seconds / 3600} 小时`;
  }

  if (seconds >= 60 && seconds % 60 === 0) {
    return `${seconds / 60} 分钟`;
  }

  return `${seconds} 秒`;
}

async function copyToClipboard(value: string) {
  await navigator.clipboard.writeText(value);
}

export function SettingsPage() {
  const queryClient = useQueryClient();
  const { refreshUser, user } = useAuth();
  const [activeTab, setActiveTab] = useState<SettingsTab>('personal');
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [busyKey, setBusyKey] = useState<string | null>(null);
  const [profileFields, setProfileFields] = useState(defaultProfileFields);
  const [systemFields, setSystemFields] = useState(defaultSystemFields);
  const [operationFields, setOperationFields] = useState(
    defaultOperationFields,
  );
  const [otherFields, setOtherFields] = useState(defaultOtherFields);
  const [mihomoUploadProgress, setMihomoUploadProgress] = useState(0);
  const [accessToken, setAccessToken] = useState('');
  const [wechatCode, setWeChatCode] = useState('');
  const [emailAddress, setEmailAddress] = useState('');
  const [emailCode, setEmailCode] = useState('');
  const [emailTurnstileToken, setEmailTurnstileToken] = useState('');
  const [geoIPTestIP, setGeoIPTestIP] = useState('8.8.8.8');
  const [geoIPPreviewResult, setGeoIPPreviewResult] =
    useState<GeoIPPreviewResult | null>(null);
  const mihomoUploadInputRef = useRef<HTMLInputElement | null>(null);
  const isRoot = (user?.role ?? 0) >= 100;

  const publicStatusQuery = useQuery({
    queryKey: ['public-status'],
    queryFn: getPublicStatus,
  });

  const profileQuery = useQuery({
    queryKey: ['settings', 'profile'],
    queryFn: getSettingsProfile,
  });

  const optionsQuery = useQuery({
    queryKey: settingsQueryKey,
    queryFn: getOptions,
    enabled: isRoot,
  });

  useEffect(() => {
    if (profileQuery.data) {
      setProfileFields({
        username: profileQuery.data.username,
        display_name: profileQuery.data.display_name || '',
        password: '',
      });
      setEmailAddress(profileQuery.data.email || '');
    }
  }, [profileQuery.data]);

  useEffect(() => {
    const publicStatus = publicStatusQuery.data;
    if (!publicStatus) {
      return;
    }

    const resolvedServerAddress =
      publicStatus.server_address || getBrowserOrigin();

    setSystemFields((previous) => ({
      ...previous,
      ServerAddress: resolvedServerAddress || previous.ServerAddress,
      GitHubClientId: publicStatus.github_client_id || previous.GitHubClientId,
      WeChatAccountQRCodeImageURL:
        publicStatus.wechat_qrcode || previous.WeChatAccountQRCodeImageURL,
      TurnstileSiteKey:
        publicStatus.turnstile_site_key || previous.TurnstileSiteKey,
    }));
    setOtherFields((previous) => ({
      ...previous,
      SystemName: publicStatus.system_name || previous.SystemName,
      HomePageLink: publicStatus.home_page_link || previous.HomePageLink,
      Footer: publicStatus.footer_html || previous.Footer,
    }));
    setOperationFields((previous) => ({
      ...previous,
      ServerAddress: resolvedServerAddress || previous.ServerAddress,
    }));
  }, [publicStatusQuery.data]);

  useEffect(() => {
    if (!optionsQuery.data) {
      return;
    }

    const optionMap = optionsToMap(optionsQuery.data);
    const resolvedServerAddress =
      optionMap.ServerAddress ||
      publicStatusQuery.data?.server_address ||
      getBrowserOrigin();

    setSystemFields({
      ServerAddress: resolvedServerAddress,
      ServerUpdateRepo: optionMap.ServerUpdateRepo ?? defaultServerUpdateRepo,
      KernelType: optionMap.KernelType ?? 'mihomo',
      MihomoBinaryPath: optionMap.MihomoBinaryPath ?? '',
      MihomoBinaryVersion: optionMap.MihomoBinaryVersion ?? '',
      MihomoBinarySource: optionMap.MihomoBinarySource ?? '',
      ClashAllowLAN: toBoolean(optionMap.ClashAllowLAN, false),
      ClashExternalController:
        optionMap.ClashExternalController ?? '127.0.0.1:19090',
      ClashMode: optionMap.ClashMode ?? 'rule',
      ClashSecret:
        optionMap.ClashSecret ?? '3ebc195c9fbe81c01eb9299e3c6bf644',
      NodeTestDefaultURL:
        optionMap.NodeTestDefaultURL ?? 'https://cp.cloudflare.com/generate_204',
      NodeTestDefaultTimeoutMS: optionMap.NodeTestDefaultTimeoutMS ?? '8000',
      GeoIPProvider: optionMap.GeoIPProvider ?? 'disabled',
      PasswordLoginEnabled: toBoolean(optionMap.PasswordLoginEnabled, true),
      PasswordRegisterEnabled: toBoolean(
        optionMap.PasswordRegisterEnabled,
        true,
      ),
      EmailVerificationEnabled: toBoolean(
        optionMap.EmailVerificationEnabled,
        false,
      ),
      GitHubOAuthEnabled: toBoolean(optionMap.GitHubOAuthEnabled, false),
      WeChatAuthEnabled: toBoolean(optionMap.WeChatAuthEnabled, false),
      TurnstileCheckEnabled: toBoolean(optionMap.TurnstileCheckEnabled, false),
      RegisterEnabled: toBoolean(optionMap.RegisterEnabled, false),
      SMTPServer: optionMap.SMTPServer ?? '',
      SMTPPort: optionMap.SMTPPort ?? '587',
      SMTPAccount: optionMap.SMTPAccount ?? '',
      SMTPToken: '',
      GitHubClientId: optionMap.GitHubClientId ?? '',
      GitHubClientSecret: '',
      WeChatServerAddress: optionMap.WeChatServerAddress ?? '',
      WeChatServerToken: '',
      WeChatAccountQRCodeImageURL: optionMap.WeChatAccountQRCodeImageURL ?? '',
      TurnstileSiteKey: optionMap.TurnstileSiteKey ?? '',
      TurnstileSecretKey: '',
    });

    setOperationFields({
      GlobalApiRateLimitNum: optionMap.GlobalApiRateLimitNum ?? '300',
      GlobalApiRateLimitDuration: optionMap.GlobalApiRateLimitDuration ?? '180',
      GlobalWebRateLimitNum: optionMap.GlobalWebRateLimitNum ?? '300',
      GlobalWebRateLimitDuration: optionMap.GlobalWebRateLimitDuration ?? '180',
      UploadRateLimitNum: optionMap.UploadRateLimitNum ?? '50',
      UploadRateLimitDuration: optionMap.UploadRateLimitDuration ?? '60',
      DownloadRateLimitNum: optionMap.DownloadRateLimitNum ?? '50',
      DownloadRateLimitDuration: optionMap.DownloadRateLimitDuration ?? '60',
      CriticalRateLimitNum: optionMap.CriticalRateLimitNum ?? '100',
      CriticalRateLimitDuration: optionMap.CriticalRateLimitDuration ?? '1200',
      ServerAddress: resolvedServerAddress,
    });

    setOtherFields({
      Notice: optionMap.Notice ?? '',
      SystemName: optionMap.SystemName ?? '',
      HomePageLink: optionMap.HomePageLink ?? '',
      About: optionMap.About ?? '',
      Footer: optionMap.Footer ?? '',
    });
  }, [optionsQuery.data, publicStatusQuery.data?.server_address]);

  const accessTokenMutation = useMutation({
    mutationFn: generateAccessToken,
    onSuccess: (token) => {
      setAccessToken(token);
      setFeedback({
        tone: 'success',
        message: '访问令牌已重置，并已在当前页面展示。',
      });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const geoIPPreviewMutation = useMutation({
    mutationFn: ({ provider, ip }: { provider: string; ip: string }) =>
      previewGeoIP(provider, ip),
    onSuccess: (result) => {
      setGeoIPPreviewResult(result);
      setFeedback({
        tone: 'success',
        message: 'IP 归属解析测试成功。',
      });
    },
    onError: (error) => {
      setGeoIPPreviewResult(null);
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const mihomoUploadMutation = useMutation({
    mutationFn: ({
      binary,
      installPath,
    }: {
      binary: File;
      installPath: string;
    }) =>
      uploadMihomoBinary(binary, installPath, (progress) => {
        setMihomoUploadProgress(progress);
      }),
    onSuccess: async (result) => {
      await queryClient.invalidateQueries({ queryKey: settingsQueryKey });
      setSystemFields((previous) => ({
        ...previous,
        MihomoBinaryPath: result.install_path,
        MihomoBinaryVersion: result.detected_version,
        MihomoBinarySource: result.binary_source,
        KernelType: result.kernel_type,
      }));
      setFeedback({
        tone: 'success',
        message: `Mihomo 二进制已安装，版本为 ${result.detected_version}。`,
      });
      setMihomoUploadProgress(100);
    },
    onError: (error) => {
      setMihomoUploadProgress(0);
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
    onSettled: () => {
      if (mihomoUploadInputRef.current) {
        mihomoUploadInputRef.current.value = '';
      }
    },
  });

  const mihomoDownloadMutation = useMutation({
    mutationFn: (installPath: string) => downloadMihomoBinary(installPath),
    onSuccess: async (result: KernelBinaryInstallResult) => {
      await queryClient.invalidateQueries({ queryKey: settingsQueryKey });
      setSystemFields((previous) => ({
        ...previous,
        MihomoBinaryPath: result.install_path,
        MihomoBinaryVersion: result.detected_version,
        MihomoBinarySource: result.binary_source,
        KernelType: result.kernel_type,
      }));
      setFeedback({
        tone: 'success',
        message: `已安装官方 Mihomo 版本 ${result.detected_version}。`,
      });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const mihomoInspectMutation = useMutation({
    mutationFn: (installPath: string) => inspectMihomoBinary(installPath),
    onSuccess: async (result: KernelBinaryInstallResult) => {
      await queryClient.invalidateQueries({ queryKey: settingsQueryKey });
      setSystemFields((previous) => ({
        ...previous,
        MihomoBinaryPath: result.install_path,
        MihomoBinaryVersion: result.detected_version,
        MihomoBinarySource: result.binary_source,
        KernelType: result.kernel_type,
      }));
      setFeedback({
        tone: 'success',
        message: `Mihomo 二进制检查通过，版本为 ${result.detected_version}。`,
      });
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const tabs = useMemo(
    () => [
      {
        key: 'personal' as const,
        label: '个人设置',
        description: '更新个人资料、绑定账号与访问令牌。',
      },
      ...(isRoot
        ? [
            {
              key: 'clash' as const,
              label: 'Clash 设置',
              description: '代理内核、控制接口与默认测速参数。',
            },
            {
              key: 'system' as const,
              label: '系统设置',
              description: '登录注册、SMTP、OAuth、更新源与归属设置。',
            },
            {
              key: 'other' as const,
              label: '其他设置',
              description: '公告、关于与品牌信息。',
            },
          ]
        : []),
    ],
    [isRoot],
  );

  const runBusyAction = async (key: string, action: () => Promise<void>) => {
    setBusyKey(key);
    setFeedback(null);

    try {
      await action();
    } catch (error) {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    } finally {
      setBusyKey(null);
    }
  };

  const saveOptionEntries = async (
    entries: Array<[string, string]>,
    successMessage: string,
  ) => {
    for (const [key, value] of entries) {
      await updateOption(key, value);
    }

    await queryClient.invalidateQueries({ queryKey: settingsQueryKey });
    await queryClient.invalidateQueries({ queryKey: ['public-status'] });
    setFeedback({ tone: 'success', message: successMessage });
  };

  const handleProfileSave = () => {
    void runBusyAction('profile', async () => {
      await updateSelf({
        username: profileFields.username.trim(),
        display_name: profileFields.display_name.trim(),
        password: profileFields.password,
      });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] }),
        refreshUser(),
      ]);
      setProfileFields((previous) => ({ ...previous, password: '' }));
      setFeedback({ tone: 'success', message: '个人资料已更新。' });
    });
  };

  const handleEmailVerification = () => {
    if (!emailAddress.trim()) {
      setFeedback({ tone: 'danger', message: '请输入要绑定的邮箱地址。' });
      return;
    }

    if (publicStatusQuery.data?.turnstile_check && !emailTurnstileToken) {
      setFeedback({ tone: 'info', message: '请先完成人机验证。' });
      return;
    }

    void runBusyAction('email-send', async () => {
      await sendEmailVerification(
        emailAddress.trim(),
        emailTurnstileToken || undefined,
      );
      setFeedback({ tone: 'success', message: '验证码已发送，请检查邮箱。' });
    });
  };

  const handleBindEmail = () => {
    if (!emailAddress.trim() || !emailCode.trim()) {
      setFeedback({ tone: 'danger', message: '请输入邮箱地址和验证码。' });
      return;
    }

    void runBusyAction('email-bind', async () => {
      await bindEmail(emailAddress.trim(), emailCode.trim());
      setEmailCode('');
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] }),
        refreshUser(),
      ]);
      setFeedback({ tone: 'success', message: '邮箱已绑定。' });
    });
  };

  const handleBindWeChat = () => {
    if (!wechatCode.trim()) {
      setFeedback({ tone: 'danger', message: '请输入微信验证码。' });
      return;
    }

    void runBusyAction('wechat-bind', async () => {
      await bindWeChat(wechatCode.trim());
      setWeChatCode('');
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] }),
        refreshUser(),
      ]);
      setFeedback({ tone: 'success', message: '微信账号已绑定。' });
    });
  };

  const handleToggleOption = (
    key: keyof typeof systemFields,
    nextValue: boolean,
  ) => {
    setSystemFields((previous) => ({ ...previous, [key]: nextValue }));

    void runBusyAction(`toggle-${key}`, async () => {
      await saveOptionEntries([[key, String(nextValue)]], '系统开关已更新。');
    });
  };

  const requireMihomoInstallPath = (allowDefault = false) => {
    const installPath = systemFields.MihomoBinaryPath.trim() || (allowDefault ? defaultMihomoInstallPath() : '');
    if (!installPath) {
      setFeedback({
        tone: 'danger',
        message: '请先填写 Mihomo 二进制文件路径。',
      });
      return '';
    }
    return installPath;
  };

  const handleClashSettingsSave = () => {
    void runBusyAction('clash-settings', async () => {
      const timeout = Number.parseInt(systemFields.NodeTestDefaultTimeoutMS, 10);
      if (Number.isNaN(timeout) || timeout <= 0) {
        throw new Error('默认测速超时必须为大于 0 的整数。');
      }
      if (timeout > 60000) {
        throw new Error('默认测速超时不能超过 60000 毫秒。');
      }
      if (!systemFields.NodeTestDefaultURL.trim()) {
        throw new Error('默认测速 URL 不能为空。');
      }
      if (!systemFields.ClashExternalController.trim()) {
        throw new Error('external-controller 不能为空。');
      }
      if (!/^[^:]+:\d+$/.test(systemFields.ClashExternalController.trim())) {
        throw new Error('external-controller 必须为 host:port 格式。');
      }
      if (!systemFields.ClashSecret.trim()) {
        throw new Error('secret 不能为空。');
      }
      if (!['rule', 'global', 'direct'].includes(systemFields.ClashMode.trim())) {
        throw new Error('mode 仅支持 rule、global 或 direct。');
      }

      await saveOptionEntries(
        [
          ['KernelType', systemFields.KernelType],
          ['ClashAllowLAN', String(systemFields.ClashAllowLAN)],
          ['ClashExternalController', systemFields.ClashExternalController.trim()],
          ['ClashMode', systemFields.ClashMode.trim()],
          ['ClashSecret', systemFields.ClashSecret.trim()],
          ['NodeTestDefaultURL', systemFields.NodeTestDefaultURL.trim()],
          ['NodeTestDefaultTimeoutMS', String(timeout)],
        ],
        'Clash 设置已保存。',
      );
    });
  };

  const handleMihomoUploadSelect = async (event: ChangeEvent<HTMLInputElement>) => {
    const binary = event.target.files?.[0];
    if (!binary) {
      return;
    }
    const installPath = requireMihomoInstallPath(true);
    if (!installPath) {
      event.target.value = '';
      return;
    }
    setSystemFields((previous) => ({ ...previous, MihomoBinaryPath: installPath }));
    setFeedback(null);
    setMihomoUploadProgress(0);
    await mihomoUploadMutation.mutateAsync({ binary, installPath });
  };

  const handleMihomoAutoDownload = () => {
    const installPath = requireMihomoInstallPath(true);
    if (!installPath) {
      return;
    }
    setSystemFields((previous) => ({ ...previous, MihomoBinaryPath: installPath }));
    setFeedback(null);
    mihomoDownloadMutation.mutate(installPath);
  };

  const handleMihomoInspect = () => {
    const installPath = requireMihomoInstallPath(false);
    if (!installPath) {
      return;
    }
    setFeedback(null);
    mihomoInspectMutation.mutate(installPath);
  };

  const renderTabContent = () => {
    if (profileQuery.isLoading || publicStatusQuery.isLoading) {
      return <LoadingState />;
    }

    if (profileQuery.isError) {
      return (
        <ErrorState
          title="个人设置加载失败"
          description={getErrorMessage(profileQuery.error)}
        />
      );
    }

    if (publicStatusQuery.isError) {
      return (
        <ErrorState
          title="系统状态加载失败"
          description={getErrorMessage(publicStatusQuery.error)}
        />
      );
    }

    const publicStatus = publicStatusQuery.data;
    const profile = profileQuery.data;

    if (!publicStatus || !profile) {
      return (
        <EmptyState
          title="设置暂不可用"
          description="未获取到当前用户或系统状态信息。"
        />
      );
    }

    if (activeTab === 'personal') {
      return (
        <div className="space-y-6">
          <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
            <AppCard
              title="个人资料"
              description="可更新用户名、显示名称和密码。留空密码表示保持当前密码不变。"
              action={
                <PrimaryButton
                  type="button"
                  onClick={handleProfileSave}
                  disabled={busyKey === 'profile'}
                >
                  {busyKey === 'profile' ? '保存中...' : '保存资料'}
                </PrimaryButton>
              }
            >
              <div className="space-y-5">
                <ResourceField label="用户名">
                  <ResourceInput
                    value={profileFields.username}
                    onChange={(event) =>
                      setProfileFields((previous) => ({
                        ...previous,
                        username: event.target.value,
                      }))
                    }
                    placeholder="请输入用户名"
                  />
                </ResourceField>

                <ResourceField label="显示名称">
                  <ResourceInput
                    value={profileFields.display_name}
                    onChange={(event) =>
                      setProfileFields((previous) => ({
                        ...previous,
                        display_name: event.target.value,
                      }))
                    }
                    placeholder="请输入显示名称"
                  />
                </ResourceField>

                <ResourceField label="新密码" hint="留空表示不修改密码。">
                  <ResourceInput
                    type="password"
                    value={profileFields.password}
                    onChange={(event) =>
                      setProfileFields((previous) => ({
                        ...previous,
                        password: event.target.value,
                      }))
                    }
                    placeholder="请输入新密码"
                  />
                </ResourceField>
              </div>
            </AppCard>

            <AppCard
              title="访问令牌"
              description="重置后会立即生成新的访问令牌，可用于自动化请求。"
              action={
                <PrimaryButton
                  type="button"
                  onClick={() => accessTokenMutation.mutate()}
                  disabled={accessTokenMutation.isPending}
                >
                  {accessTokenMutation.isPending ? '生成中...' : '重置令牌'}
                </PrimaryButton>
              }
            >
              <div className="space-y-4">
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                    <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                      当前角色
                    </p>
                    <div className="mt-2">
                      <StatusBadge
                        label={
                          user?.role === 100
                            ? '超级管理员'
                            : user?.role === 10
                              ? '管理员'
                              : '普通用户'
                        }
                        variant={
                          user?.role === 100
                            ? 'warning'
                            : user?.role === 10
                              ? 'info'
                              : 'success'
                        }
                      />
                    </div>
                  </div>
                  <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                    <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                      已绑定邮箱
                    </p>
                    <p className="mt-2 text-sm break-all text-[var(--foreground-primary)]">
                      {profile.email || '未绑定'}
                    </p>
                  </div>
                </div>

                {accessToken ? (
                  <div className="space-y-3">
                    <CodeBlock className="break-all whitespace-pre-wrap">
                      {accessToken}
                    </CodeBlock>
                    <SecondaryButton
                      type="button"
                      onClick={() => void copyToClipboard(accessToken)}
                    >
                      复制令牌
                    </SecondaryButton>
                  </div>
                ) : (
                  <EmptyState
                    title="尚未生成令牌"
                    description="点击“重置令牌”后，新的访问令牌会显示在这里。"
                  />
                )}
              </div>
            </AppCard>
          </div>

          <AppCard
            title="账号绑定"
            description="支持绑定 GitHub、微信和邮箱地址，用于统一个人身份入口。"
          >
            <div className="grid gap-6 xl:grid-cols-3">
              <div className="space-y-4 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                <div className="space-y-1">
                  <p className="text-base font-semibold text-[var(--foreground-primary)]">
                    GitHub 账号
                  </p>
                  <p className="text-sm leading-6 text-[var(--foreground-secondary)]">
                    当前状态：
                    {profile.github_id
                      ? `已绑定 ${profile.github_id}`
                      : '未绑定'}
                  </p>
                </div>
                <PrimaryButton
                  type="button"
                  onClick={() =>
                    window.open(
                      `https://github.com/login/oauth/authorize?client_id=${publicStatus.github_client_id}&scope=user:email`,
                      '_blank',
                      'noopener,noreferrer',
                    )
                  }
                  disabled={
                    !publicStatus.github_oauth || !publicStatus.github_client_id
                  }
                >
                  {publicStatus.github_oauth
                    ? '绑定 GitHub'
                    : '未启用 GitHub OAuth'}
                </PrimaryButton>
              </div>

              <div className="space-y-4 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                <div className="space-y-1">
                  <p className="text-base font-semibold text-[var(--foreground-primary)]">
                    微信账号
                  </p>
                  <p className="text-sm leading-6 text-[var(--foreground-secondary)]">
                    当前状态：
                    {profile.wechat_id
                      ? `已绑定 ${profile.wechat_id}`
                      : '未绑定'}
                  </p>
                </div>
                {publicStatus.wechat_login && publicStatus.wechat_qrcode ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={publicStatus.wechat_qrcode}
                    alt="微信绑定二维码"
                    className="h-40 w-40 rounded-2xl border border-[var(--border-default)] object-cover"
                  />
                ) : null}
                <ResourceField
                  label="验证码"
                  hint="扫码关注后输入“验证码”获取绑定码。"
                >
                  <ResourceInput
                    value={wechatCode}
                    onChange={(event) => setWeChatCode(event.target.value)}
                    placeholder="请输入微信验证码"
                  />
                </ResourceField>
                <PrimaryButton
                  type="button"
                  onClick={handleBindWeChat}
                  disabled={
                    !publicStatus.wechat_login || busyKey === 'wechat-bind'
                  }
                >
                  {busyKey === 'wechat-bind' ? '绑定中...' : '绑定微信'}
                </PrimaryButton>
              </div>

              <div className="space-y-4 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                <div className="space-y-1">
                  <p className="text-base font-semibold text-[var(--foreground-primary)]">
                    邮箱地址
                  </p>
                  <p className="text-sm leading-6 text-[var(--foreground-secondary)]">
                    当前状态：
                    {profile.email ? `已绑定 ${profile.email}` : '未绑定'}
                  </p>
                </div>
                <div className="space-y-4">
                  <ResourceField label="邮箱地址">
                    <ResourceInput
                      value={emailAddress}
                      onChange={(event) => setEmailAddress(event.target.value)}
                      placeholder="请输入邮箱地址"
                    />
                  </ResourceField>
                  <ResourceField label="验证码">
                    <ResourceInput
                      value={emailCode}
                      onChange={(event) => setEmailCode(event.target.value)}
                      placeholder="请输入邮箱验证码"
                    />
                  </ResourceField>
                  {publicStatus.turnstile_check ? (
                    publicStatus.turnstile_site_key ? (
                      <TurnstileWidget
                        siteKey={publicStatus.turnstile_site_key}
                        onVerify={(token) => setEmailTurnstileToken(token)}
                        onExpire={() => setEmailTurnstileToken('')}
                        onError={() => setEmailTurnstileToken('')}
                      />
                    ) : (
                      <EmptyState
                        title="Turnstile 配置不完整"
                        description="当前系统已启用 Turnstile，但未配置 Site Key，邮箱绑定暂不可用。"
                      />
                    )
                  ) : null}
                  <div className="flex flex-wrap gap-2">
                    <SecondaryButton
                      type="button"
                      onClick={handleEmailVerification}
                      disabled={busyKey === 'email-send'}
                    >
                      {busyKey === 'email-send' ? '发送中...' : '发送验证码'}
                    </SecondaryButton>
                    <PrimaryButton
                      type="button"
                      onClick={handleBindEmail}
                      disabled={busyKey === 'email-bind'}
                    >
                      {busyKey === 'email-bind' ? '绑定中...' : '绑定邮箱'}
                    </PrimaryButton>
                  </div>
                </div>
              </div>
            </div>
          </AppCard>
        </div>
      );
    }

    if (!isRoot) {
      return (
        <EmptyState
          title="权限不足"
          description="只有超级管理员可以访问系统级设置。"
        />
      );
    }

    if (optionsQuery.isLoading) {
      return <LoadingState />;
    }

    if (optionsQuery.isError) {
      return (
        <ErrorState
          title="设置项加载失败"
          description={getErrorMessage(optionsQuery.error)}
        />
      );
    }

    if (activeTab === 'clash') {
      return (
        <div className="space-y-6">
          <AppCard
            title="代理内核设置"
            description="当前仅支持 Mihomo，Xray 与 sing-box 入口已预留。完成安装后会自动校验版本并写回配置。"
            action={
              <PrimaryButton
                type="button"
                onClick={handleClashSettingsSave}
                disabled={busyKey === 'clash-settings'}
              >
                {busyKey === 'clash-settings' ? '保存中...' : '保存 Clash 设置'}
              </PrimaryButton>
            }
          >
            <div className="grid gap-5 xl:grid-cols-[0.9fr_1.1fr]">
              <div className="space-y-5">
                <ResourceField
                  label="内核类型"
                  hint="当前版本仅开放 Mihomo，其他内核先保留配置入口。"
                >
                  <ResourceSelect
                    value={systemFields.KernelType}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        KernelType: event.target.value,
                      }))
                    }
                  >
                    <option value="mihomo">Mihomo</option>
                    <option value="xray" disabled>
                      Xray（预留）
                    </option>
                    <option value="singbox" disabled>
                      sing-box（预留）
                    </option>
                  </ResourceSelect>
                </ResourceField>

                <ResourceField
                  label="Mihomo 二进制路径"
                  hint="可填写已存在的二进制路径并点击检查；如果留空，上传或自动下载时会默认安装到当前工作目录。"
                >
                  <ResourceInput
                    value={systemFields.MihomoBinaryPath}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        MihomoBinaryPath: event.target.value,
                      }))
                    }
                    placeholder={
                      typeof window !== 'undefined' &&
                      navigator.userAgent.includes('Windows')
                        ? 'C:\\poolx\\bin\\mihomo.exe'
                        : '/usr/local/bin/mihomo'
                    }
                  />
                </ResourceField>

                <input
                  ref={mihomoUploadInputRef}
                  type="file"
                  className="hidden"
                  onChange={(event) => {
                    void handleMihomoUploadSelect(event);
                  }}
                />

                <div className="flex flex-wrap gap-3">
                  <SecondaryButton
                    type="button"
                    onClick={handleMihomoInspect}
                    disabled={
                      mihomoInspectMutation.isPending ||
                      mihomoUploadMutation.isPending ||
                      mihomoDownloadMutation.isPending
                    }
                  >
                    {mihomoInspectMutation.isPending ? '检查中...' : '检查内核'}
                  </SecondaryButton>
                  <SecondaryButton
                    type="button"
                    onClick={() => mihomoUploadInputRef.current?.click()}
                    disabled={
                      mihomoInspectMutation.isPending ||
                      mihomoUploadMutation.isPending ||
                      mihomoDownloadMutation.isPending
                    }
                  >
                    {mihomoUploadMutation.isPending
                      ? '上传校验中...'
                      : '手动上传 Mihomo'}
                  </SecondaryButton>
                  <PrimaryButton
                    type="button"
                    onClick={handleMihomoAutoDownload}
                    disabled={
                      mihomoInspectMutation.isPending ||
                      mihomoUploadMutation.isPending ||
                      mihomoDownloadMutation.isPending
                    }
                  >
                    {mihomoDownloadMutation.isPending
                      ? '下载校验中...'
                      : '自动下载官方发行版'}
                  </PrimaryButton>
                </div>

                {mihomoUploadMutation.isPending ? (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-xs text-[var(--foreground-secondary)]">
                      <span>上传进度</span>
                      <span>{mihomoUploadProgress}%</span>
                    </div>
                    <div className="h-2 overflow-hidden rounded-full bg-[var(--surface-muted)]">
                      <div
                        className="h-full bg-[var(--brand-primary)] transition-all"
                        style={{ width: `${mihomoUploadProgress}%` }}
                      />
                    </div>
                  </div>
                ) : null}
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    当前内核
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    {systemFields.KernelType || 'mihomo'}
                  </p>
                </div>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    已校验版本
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    {systemFields.MihomoBinaryVersion || '尚未安装'}
                  </p>
                </div>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    安装来源
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    {systemFields.MihomoBinarySource === 'upload'
                      ? '手动上传'
                      : systemFields.MihomoBinarySource === 'download'
                        ? '官方自动下载'
                        : systemFields.MihomoBinarySource === 'existing'
                          ? '现有路径检查'
                        : '未设置'}
                  </p>
                </div>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4 md:col-span-2">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    当前路径
                  </p>
                  <p className="mt-2 break-all text-sm text-[var(--foreground-primary)]">
                    {systemFields.MihomoBinaryPath || '尚未配置'}
                  </p>
                </div>
              </div>
            </div>
          </AppCard>

          <AppCard
            title="Clash 运行参数"
            description="这些设置会参与最终 Mihomo 配置渲染，并在启动与热重载时生效。"
          >
            <div className="grid gap-5 xl:grid-cols-2">
              <ResourceField
                label="external-controller"
                hint="控制接口监听地址，格式为 host:port。"
              >
                <ResourceInput
                  value={systemFields.ClashExternalController}
                  onChange={(event) =>
                    setSystemFields((previous) => ({
                      ...previous,
                      ClashExternalController: event.target.value,
                    }))
                  }
                  placeholder="127.0.0.1:19090"
                />
              </ResourceField>
              <ResourceField
                label="mode"
                hint="控制最终 Clash 运行模式。"
              >
                <ResourceSelect
                  value={systemFields.ClashMode}
                  onChange={(event) =>
                    setSystemFields((previous) => ({
                      ...previous,
                      ClashMode: event.target.value,
                    }))
                  }
                >
                  <option value="rule">rule</option>
                  <option value="global">global</option>
                  <option value="direct">direct</option>
                </ResourceSelect>
              </ResourceField>
              <ResourceField
                label="secret"
                hint="控制接口访问密钥，运行控制和 API 探活都会使用它。"
              >
                <ResourceInput
                  value={systemFields.ClashSecret}
                  onChange={(event) =>
                    setSystemFields((previous) => ({
                      ...previous,
                      ClashSecret: event.target.value,
                    }))
                  }
                  placeholder="3ebc195c9fbe81c01eb9299e3c6bf644"
                />
              </ResourceField>
              <ToggleField
                label="allow-lan"
                description="开启后允许局域网设备访问代理监听端口。"
                checked={systemFields.ClashAllowLAN}
                onChange={(checked) =>
                  setSystemFields((previous) => ({
                    ...previous,
                    ClashAllowLAN: checked,
                  }))
                }
              />
            </div>
          </AppCard>

          <AppCard
            title="默认测速参数"
            description="配置导入和节点池的测速操作都会统一使用这里的默认 URL 与超时。"
          >
            <div className="grid gap-5 xl:grid-cols-[minmax(0,1.8fr)_minmax(0,220px)]">
              <ResourceField
                label="默认测速 URL"
                hint="节点测试会通过代理访问该地址，建议使用返回体小、响应稳定的探测地址。"
              >
                <ResourceInput
                  value={systemFields.NodeTestDefaultURL}
                  onChange={(event) =>
                    setSystemFields((previous) => ({
                      ...previous,
                      NodeTestDefaultURL: event.target.value,
                    }))
                  }
                  placeholder="https://cp.cloudflare.com/generate_204"
                />
              </ResourceField>
              <ResourceField
                label="默认超时（毫秒）"
                hint="建议保持在 3000 到 15000 之间。"
              >
                <ResourceInput
                  type="number"
                  value={systemFields.NodeTestDefaultTimeoutMS}
                  onChange={(event) =>
                    setSystemFields((previous) => ({
                      ...previous,
                      NodeTestDefaultTimeoutMS: event.target.value,
                    }))
                  }
                  inputMode="numeric"
                  placeholder="8000"
                />
              </ResourceField>
            </div>
          </AppCard>
        </div>
      );
    }

    if (activeTab === 'system') {
      return (
        <div className="space-y-6">

          <AppCard
            title="登录与注册开关"
            description="切换后立即生效，无需重启服务。"
          >
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <ToggleField
                label="允许密码登录"
                description="关闭后将无法使用用户名密码登录。"
                checked={systemFields.PasswordLoginEnabled}
                onChange={(checked) =>
                  handleToggleOption('PasswordLoginEnabled', checked)
                }
                disabled={busyKey === 'toggle-PasswordLoginEnabled'}
              />
              <ToggleField
                label="允许密码注册"
                description="关闭后新用户不能通过密码方式注册。"
                checked={systemFields.PasswordRegisterEnabled}
                onChange={(checked) =>
                  handleToggleOption('PasswordRegisterEnabled', checked)
                }
                disabled={busyKey === 'toggle-PasswordRegisterEnabled'}
              />
              <ToggleField
                label="注册需要邮箱验证"
                description="开启后，新用户注册必须先完成邮箱验证码校验。"
                checked={systemFields.EmailVerificationEnabled}
                onChange={(checked) =>
                  handleToggleOption('EmailVerificationEnabled', checked)
                }
                disabled={busyKey === 'toggle-EmailVerificationEnabled'}
              />
              <ToggleField
                label="启用 GitHub OAuth"
                description="允许用户通过 GitHub 登录与注册。"
                checked={systemFields.GitHubOAuthEnabled}
                onChange={(checked) =>
                  handleToggleOption('GitHubOAuthEnabled', checked)
                }
                disabled={busyKey === 'toggle-GitHubOAuthEnabled'}
              />
              <ToggleField
                label="启用微信登录"
                description="允许用户通过微信入口登录与注册。"
                checked={systemFields.WeChatAuthEnabled}
                onChange={(checked) =>
                  handleToggleOption('WeChatAuthEnabled', checked)
                }
                disabled={busyKey === 'toggle-WeChatAuthEnabled'}
              />
              <ToggleField
                label="启用 Turnstile"
                description="开启后注册、邮箱验证码等流程需要先通过人机验证。"
                checked={systemFields.TurnstileCheckEnabled}
                onChange={(checked) =>
                  handleToggleOption('TurnstileCheckEnabled', checked)
                }
                disabled={busyKey === 'toggle-TurnstileCheckEnabled'}
              />
              <ToggleField
                label="允许新用户注册"
                description="关闭后将禁止所有新用户注册入口。"
                checked={systemFields.RegisterEnabled}
                onChange={(checked) =>
                  handleToggleOption('RegisterEnabled', checked)
                }
                disabled={busyKey === 'toggle-RegisterEnabled'}
              />
            </div>
          </AppCard>

          <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
            <AppCard
              title="系统更新"
              description="配置服务端版本检查使用的上游仓库。"
              action={
                <PrimaryButton
                  type="button"
                  onClick={() =>
                    void runBusyAction('system-runtime', async () => {
                      await saveOptionEntries(
                        [
                          [
                            'ServerUpdateRepo',
                            systemFields.ServerUpdateRepo.trim(),
                          ],
                        ],
                        '更新设置已保存。',
                      );
                    })
                  }
                  disabled={busyKey === 'system-runtime'}
                >
                  {busyKey === 'system-runtime' ? '保存中...' : '保存更新设置'}
                </PrimaryButton>
              }
            >
              <div className="space-y-5">
                <ResourceField
                  label="上游更新仓库"
                  hint="默认使用 Rain-kl/PoolX，也可按 owner/repo 格式改为你自己的发布仓库。"
                >
                  <ResourceInput
                    value={systemFields.ServerUpdateRepo}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        ServerUpdateRepo: event.target.value,
                      }))
                    }
                    placeholder={defaultServerUpdateRepo}
                  />
                </ResourceField>
              </div>
              <div className="grid gap-4 md:grid-cols-2 mt-5">
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    服务端版本
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    {publicStatus.version}
                  </p>
                </div>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    Server 启动时间
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    {formatDateTime(new Date(publicStatus.start_time * 1000))}
                  </p>
                </div>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    运行模式
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    静态导出 + Go Server 托管
                  </p>
                </div>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4 md:col-span-2">
                  <p className="text-xs tracking-[0.2em] text-[var(--foreground-muted)] uppercase">
                    版本入口
                  </p>
                  <p className="mt-2 text-sm text-[var(--foreground-primary)]">
                    点击顶栏“版本”可检查更新、查看 Release
                    Notes，并直接触发服务端升级。
                  </p>
                </div>
              </div>
            </AppCard>

            <AppCard
              title="通用设置"
              description="服务器地址会影响邮件链接、OAuth 回调和部署命令展示。"
              action={
                <div className="flex flex-wrap gap-2">
                  <SecondaryButton
                    type="button"
                    onClick={() =>
                      window.open(
                        '/swagger/index.html',
                        '_blank',
                        'noopener,noreferrer',
                      )
                    }
                  >
                    打开接口文档
                  </SecondaryButton>
                  <PrimaryButton
                    type="button"
                    onClick={() =>
                      void runBusyAction('system-general', async () => {
                        await saveOptionEntries(
                          [
                            [
                              'ServerAddress',
                              normalizeServerUrl(systemFields.ServerAddress),
                            ],
                            ['GeoIPProvider', systemFields.GeoIPProvider],
                          ],
                          '通用设置已保存。',
                        );
                      })
                    }
                    disabled={busyKey === 'system-general'}
                  >
                    {busyKey === 'system-general'
                      ? '保存中...'
                      : '保存通用设置'}
                  </PrimaryButton>
                </div>
              }
            >
              <ResourceField label="服务器地址">
                <ResourceInput
                  value={systemFields.ServerAddress}
                  onChange={(event) =>
                    setSystemFields((previous) => ({
                      ...previous,
                      ServerAddress: event.target.value,
                    }))
                  }
                  placeholder="https://yourdomain.com"
                />
              </ResourceField>
              <div className="mt-5 space-y-5">
                <ResourceField
                  label="IP 归属方式"
                  hint="disabled 关闭归属解析；mmdb 使用本地数据库；其余选项调用外部 GeoIP 服务。"
                >
                  <ResourceSelect
                    value={systemFields.GeoIPProvider}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        GeoIPProvider: event.target.value,
                      }))
                    }
                  >
                    <option value="disabled">关闭</option>
                    <option value="mmdb">MaxMind mmdb</option>
                    <option value="ip-api">ip-api.com</option>
                    <option value="geojs">geojs.io</option>
                    <option value="ipinfo">ipinfo.io</option>
                  </ResourceSelect>
                </ResourceField>
                <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] p-5">
                  <div className="space-y-4">
                    <ResourceField
                      label="归属解析测试"
                      hint="测试使用当前表单中选择的归属方式，不要求先保存设置。"
                    >
                      <div className="flex flex-col gap-3 lg:flex-row lg:items-end">
                        <div className="min-w-0 flex-1">
                          <ResourceInput
                            value={geoIPTestIP}
                            onChange={(event) =>
                              setGeoIPTestIP(event.target.value)
                            }
                            placeholder="例如 8.8.8.8"
                          />
                        </div>
                        <PrimaryButton
                          type="button"
                          onClick={() => {
                            setFeedback(null);
                            void geoIPPreviewMutation.mutate({
                              provider: systemFields.GeoIPProvider,
                              ip: geoIPTestIP.trim(),
                            });
                          }}
                          disabled={geoIPPreviewMutation.isPending}
                        >
                          {geoIPPreviewMutation.isPending
                            ? '测试中...'
                            : '测试归属解析'}
                        </PrimaryButton>
                      </div>
                    </ResourceField>
                    {geoIPPreviewResult ? (
                      <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] px-4 py-4 text-sm text-[var(--foreground-secondary)]">
                        <p>Provider: {geoIPPreviewResult.provider}</p>
                        <p>IP: {geoIPPreviewResult.ip}</p>
                        <p>
                          归属地: {geoIPPreviewResult.name || '未知'} (
                          {geoIPPreviewResult.iso_code || 'N/A'})
                        </p>
                        <p>
                          坐标:{' '}
                          {geoIPPreviewResult.latitude !== undefined &&
                          geoIPPreviewResult.longitude !== undefined
                            ? `${geoIPPreviewResult.latitude}, ${geoIPPreviewResult.longitude}`
                            : '无'}
                        </p>
                      </div>
                    ) : null}
                  </div>
                </div>
              </div>
            </AppCard>

            <AppCard
              title="SMTP 设置"
              description="用于邮件验证码、密码重置和其他邮件通知发送。"
              action={
                <PrimaryButton
                  type="button"
                  onClick={() =>
                    void runBusyAction('system-smtp', async () => {
                      await saveOptionEntries(
                        [
                          ['SMTPServer', systemFields.SMTPServer.trim()],
                          ['SMTPPort', systemFields.SMTPPort.trim()],
                          ['SMTPAccount', systemFields.SMTPAccount.trim()],
                          ['SMTPToken', systemFields.SMTPToken.trim()],
                        ],
                        'SMTP 设置已保存。',
                      );
                    })
                  }
                  disabled={busyKey === 'system-smtp'}
                >
                  {busyKey === 'system-smtp' ? '保存中...' : '保存 SMTP 设置'}
                </PrimaryButton>
              }
            >
              <div className="grid gap-5 md:grid-cols-2">
                <ResourceField label="SMTP 服务器">
                  <ResourceInput
                    value={systemFields.SMTPServer}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        SMTPServer: event.target.value,
                      }))
                    }
                    placeholder="smtp.qq.com"
                  />
                </ResourceField>
                <ResourceField label="SMTP 端口">
                  <ResourceInput
                    value={systemFields.SMTPPort}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        SMTPPort: event.target.value,
                      }))
                    }
                    placeholder="587"
                  />
                </ResourceField>
                <ResourceField label="SMTP 账户">
                  <ResourceInput
                    value={systemFields.SMTPAccount}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        SMTPAccount: event.target.value,
                      }))
                    }
                    placeholder="name@example.com"
                  />
                </ResourceField>
                <ResourceField
                  label="SMTP 凭证"
                  hint="因安全原因不会回显历史密钥，留空表示不更新。"
                >
                  <ResourceInput
                    type="password"
                    value={systemFields.SMTPToken}
                    onChange={(event) =>
                      setSystemFields((previous) => ({
                        ...previous,
                        SMTPToken: event.target.value,
                      }))
                    }
                    placeholder="请输入新的 SMTP 凭证"
                  />
                </ResourceField>
              </div>
            </AppCard>

            <AppCard
              title="OAuth / WeChat / Turnstile"
              description="敏感密钥不会从后端回显，留空即保持原值。"
              action={
                <PrimaryButton
                  type="button"
                  onClick={() =>
                    void runBusyAction('system-integrations', async () => {
                      await saveOptionEntries(
                        [
                          [
                            'GitHubClientId',
                            systemFields.GitHubClientId.trim(),
                          ],
                          [
                            'GitHubClientSecret',
                            systemFields.GitHubClientSecret.trim(),
                          ],
                          [
                            'WeChatServerAddress',
                            normalizeServerUrl(
                              systemFields.WeChatServerAddress,
                            ),
                          ],
                          [
                            'WeChatServerToken',
                            systemFields.WeChatServerToken.trim(),
                          ],
                          [
                            'WeChatAccountQRCodeImageURL',
                            systemFields.WeChatAccountQRCodeImageURL.trim(),
                          ],
                          [
                            'TurnstileSiteKey',
                            systemFields.TurnstileSiteKey.trim(),
                          ],
                          [
                            'TurnstileSecretKey',
                            systemFields.TurnstileSecretKey.trim(),
                          ],
                        ],
                        '第三方集成设置已保存。',
                      );
                    })
                  }
                  disabled={busyKey === 'system-integrations'}
                >
                  {busyKey === 'system-integrations'
                    ? '保存中...'
                    : '保存集成设置'}
                </PrimaryButton>
              }
            >
              <div className="space-y-5">
                <div className="grid gap-5 md:grid-cols-2">
                  <ResourceField label="GitHub Client ID">
                    <ResourceInput
                      value={systemFields.GitHubClientId}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          GitHubClientId: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="GitHub Client Secret">
                    <ResourceInput
                      type="password"
                      value={systemFields.GitHubClientSecret}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          GitHubClientSecret: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="WeChat Server 地址">
                    <ResourceInput
                      value={systemFields.WeChatServerAddress}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          WeChatServerAddress: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="WeChat Server Token">
                    <ResourceInput
                      type="password"
                      value={systemFields.WeChatServerToken}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          WeChatServerToken: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="公众号二维码链接">
                    <ResourceInput
                      value={systemFields.WeChatAccountQRCodeImageURL}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          WeChatAccountQRCodeImageURL: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="Turnstile Site Key">
                    <ResourceInput
                      value={systemFields.TurnstileSiteKey}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          TurnstileSiteKey: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="Turnstile Secret Key">
                    <ResourceInput
                      type="password"
                      value={systemFields.TurnstileSecretKey}
                      onChange={(event) =>
                        setSystemFields((previous) => ({
                          ...previous,
                          TurnstileSecretKey: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                </div>
              </div>
            </AppCard>
          </div>
          <AppCard
            title="请求限流设置"
            description="按来源 IP 生效，保存后立即影响 Web、API、上传下载及登录注册等敏感接口。时间单位均为秒。"
            action={
              <PrimaryButton
                type="button"
                onClick={() =>
                  void runBusyAction('operation-rate-limit', async () => {
                    const entries = [
                      [
                        'GlobalApiRateLimitNum',
                        operationFields.GlobalApiRateLimitNum,
                      ],
                      [
                        'GlobalApiRateLimitDuration',
                        operationFields.GlobalApiRateLimitDuration,
                      ],
                      [
                        'GlobalWebRateLimitNum',
                        operationFields.GlobalWebRateLimitNum,
                      ],
                      [
                        'GlobalWebRateLimitDuration',
                        operationFields.GlobalWebRateLimitDuration,
                      ],
                      [
                        'UploadRateLimitNum',
                        operationFields.UploadRateLimitNum,
                      ],
                      [
                        'UploadRateLimitDuration',
                        operationFields.UploadRateLimitDuration,
                      ],
                      [
                        'DownloadRateLimitNum',
                        operationFields.DownloadRateLimitNum,
                      ],
                      [
                        'DownloadRateLimitDuration',
                        operationFields.DownloadRateLimitDuration,
                      ],
                      [
                        'CriticalRateLimitNum',
                        operationFields.CriticalRateLimitNum,
                      ],
                      [
                        'CriticalRateLimitDuration',
                        operationFields.CriticalRateLimitDuration,
                      ],
                    ] as const;

                    for (const [key, rawValue] of entries) {
                      const parsedValue = Number.parseInt(rawValue, 10);
                      if (Number.isNaN(parsedValue) || parsedValue <= 0) {
                        throw new Error(`${key} 必须为大于 0 的整数。`);
                      }
                      if (key.endsWith('Duration') && parsedValue > 1200) {
                        throw new Error(`${key} 不能超过 1200 秒。`);
                      }
                    }

                    await saveOptionEntries(
                      entries.map(([key, value]) => [
                        key,
                        String(Number.parseInt(value, 10)),
                      ]),
                      '限流设置已保存。',
                    );
                  })
                }
                disabled={busyKey === 'operation-rate-limit'}
              >
                {busyKey === 'operation-rate-limit'
                  ? '保存中...'
                  : '保存限流设置'}
              </PrimaryButton>
            }
          >
            <div className="grid gap-4 lg:grid-cols-2">
              <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] p-5">
                <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                  全局 API 限流
                </p>
                <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                  作用于 `/api` 下的通用请求。
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <ResourceField label="请求次数">
                    <ResourceInput
                      type="number"
                      value={operationFields.GlobalApiRateLimitNum}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          GlobalApiRateLimitNum: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField
                    label={`时间窗口 (${formatSecondsLabel(operationFields.GlobalApiRateLimitDuration)})`}
                  >
                    <ResourceInput
                      type="number"
                      value={operationFields.GlobalApiRateLimitDuration}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          GlobalApiRateLimitDuration: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                </div>
              </div>

              <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] p-5">
                <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                  全局 Web 限流
                </p>
                <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                  作用于页面和静态资源请求，过低会更容易触发 429。
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <ResourceField label="请求次数">
                    <ResourceInput
                      type="number"
                      value={operationFields.GlobalWebRateLimitNum}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          GlobalWebRateLimitNum: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField
                    label={`时间窗口 (${formatSecondsLabel(operationFields.GlobalWebRateLimitDuration)})`}
                  >
                    <ResourceInput
                      type="number"
                      value={operationFields.GlobalWebRateLimitDuration}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          GlobalWebRateLimitDuration: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                </div>
              </div>

              <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] p-5">
                <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                  上传 / 下载限流
                </p>
                <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                  用于文件上传与下载接口，建议保留相对严格的阈值。
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <ResourceField label="上传请求次数">
                    <ResourceInput
                      type="number"
                      value={operationFields.UploadRateLimitNum}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          UploadRateLimitNum: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField
                    label={`上传窗口 (${formatSecondsLabel(operationFields.UploadRateLimitDuration)})`}
                  >
                    <ResourceInput
                      type="number"
                      value={operationFields.UploadRateLimitDuration}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          UploadRateLimitDuration: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField label="下载请求次数">
                    <ResourceInput
                      type="number"
                      value={operationFields.DownloadRateLimitNum}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          DownloadRateLimitNum: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField
                    label={`下载窗口 (${formatSecondsLabel(operationFields.DownloadRateLimitDuration)})`}
                  >
                    <ResourceInput
                      type="number"
                      value={operationFields.DownloadRateLimitDuration}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          DownloadRateLimitDuration: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                </div>
              </div>

              <div className="rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] p-5">
                <p className="text-sm font-semibold text-[var(--foreground-primary)]">
                  敏感接口限流
                </p>
                <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                  用于登录、注册、验证码、重置密码和 OAuth 等接口。
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <ResourceField label="请求次数">
                    <ResourceInput
                      type="number"
                      value={operationFields.CriticalRateLimitNum}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          CriticalRateLimitNum: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                  <ResourceField
                    label={`时间窗口 (${formatSecondsLabel(operationFields.CriticalRateLimitDuration)})`}
                  >
                    <ResourceInput
                      type="number"
                      value={operationFields.CriticalRateLimitDuration}
                      onChange={(event) =>
                        setOperationFields((previous) => ({
                          ...previous,
                          CriticalRateLimitDuration: event.target.value,
                        }))
                      }
                    />
                  </ResourceField>
                </div>
              </div>
            </div>
          </AppCard>
        </div>
      );
    }

    return (
      <div className="space-y-6">
        <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
          <AppCard
            title="公告与品牌信息"
            description="用于控制首页公告、系统名称、默认首页链接和页脚展示。"
            action={
              <PrimaryButton
                type="button"
                onClick={() =>
                  void runBusyAction('other-brand', async () => {
                    await saveOptionEntries(
                      [
                        ['Notice', otherFields.Notice],
                        ['SystemName', otherFields.SystemName.trim()],
                        ['HomePageLink', otherFields.HomePageLink.trim()],
                        ['Footer', otherFields.Footer],
                      ],
                      '公告与品牌设置已保存。',
                    );
                  })
                }
                disabled={busyKey === 'other-brand'}
              >
                {busyKey === 'other-brand' ? '保存中...' : '保存基础信息'}
              </PrimaryButton>
            }
          >
            <div className="space-y-5">
              <ResourceField label="系统名称">
                <ResourceInput
                  value={otherFields.SystemName}
                  onChange={(event) =>
                    setOtherFields((previous) => ({
                      ...previous,
                      SystemName: event.target.value,
                    }))
                  }
                  placeholder="OpenFlare"
                />
              </ResourceField>
              <ResourceField label="首页链接">
                <ResourceInput
                  value={otherFields.HomePageLink}
                  onChange={(event) =>
                    setOtherFields((previous) => ({
                      ...previous,
                      HomePageLink: event.target.value,
                    }))
                  }
                  placeholder="https://example.com"
                />
              </ResourceField>
              <ResourceField label="公告">
                <ResourceTextarea
                  value={otherFields.Notice}
                  onChange={(event) =>
                    setOtherFields((previous) => ({
                      ...previous,
                      Notice: event.target.value,
                    }))
                  }
                  placeholder="可在此编写首页公告内容"
                />
              </ResourceField>
              <ResourceField label="页脚 HTML">
                <ResourceTextarea
                  value={otherFields.Footer}
                  onChange={(event) =>
                    setOtherFields((previous) => ({
                      ...previous,
                      Footer: event.target.value,
                    }))
                  }
                  placeholder="留空则使用默认页脚"
                />
              </ResourceField>
            </div>
          </AppCard>

          <AppCard
            title="关于页内容"
            description="支持 Markdown / HTML 内容编辑，保存后会同步到公开关于页。"
            action={
              <PrimaryButton
                type="button"
                onClick={() =>
                  void runBusyAction('other-about', async () => {
                    await saveOptionEntries(
                      [['About', otherFields.About]],
                      '关于页内容已保存。',
                    );
                  })
                }
                disabled={busyKey === 'other-about'}
              >
                {busyKey === 'other-about' ? '保存中...' : '保存关于内容'}
              </PrimaryButton>
            }
          >
            <div className="space-y-5">
              <ResourceField
                label="关于内容"
                hint="支持 Markdown 和 HTML，保存后会同步到公开关于页。"
              >
                <ResourceTextarea
                  value={otherFields.About}
                  onChange={(event) =>
                    setOtherFields((previous) => ({
                      ...previous,
                      About: event.target.value,
                    }))
                  }
                  placeholder="在这里编写关于 OpenFlare 的介绍内容"
                  className="min-h-48"
                />
              </ResourceField>
            </div>
          </AppCard>
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader title="设置" />

      {feedback ? (
        <InlineMessage tone={feedback.tone} message={feedback.message} />
      ) : null}

      <div className="flex flex-wrap gap-3">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            type="button"
            onClick={() => setActiveTab(tab.key)}
            className={[
              'rounded-2xl border px-4 py-3 text-left transition',
              activeTab === tab.key
                ? 'border-[var(--border-strong)] bg-[var(--accent-soft)] text-[var(--foreground-primary)]'
                : 'border-[var(--border-default)] bg-[var(--surface-muted)] text-[var(--foreground-secondary)] hover:border-[var(--border-strong)] hover:text-[var(--foreground-primary)]',
            ].join(' ')}
          >
            <p className="text-sm font-semibold">{tab.label}</p>
            <p className="mt-1 text-xs leading-5 text-inherit/80">
              {tab.description}
            </p>
          </button>
        ))}
      </div>

      {renderTabContent()}
    </div>
  );
}
