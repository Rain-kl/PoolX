'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { useAuth } from '@/components/providers/auth-provider';
import { PageHeader } from '@/components/layout/page-header';
import { AppModal } from '@/components/ui/app-modal';
import { AppCard } from '@/components/ui/app-card';
import { StatusBadge } from '@/components/ui/status-badge';
import {
  createUser,
  getUser,
  getUsers,
  manageUser,
  searchUsers,
  updateUser,
} from '@/features/users/api/users';
import type {
  ManageUserAction,
  UserItem,
  UserMutationPayload,
} from '@/features/users/types';
import {
  DangerButton,
  PrimaryButton,
  ResourceField,
  ResourceInput,
  SecondaryButton,
} from '@/features/shared/components/resource-primitives';

const ITEMS_PER_PAGE = 10;
const usersListQueryKey = ['users', 'list'] as const;

const userSchema = z.object({
  username: z.string().trim().min(1, '请输入用户名').max(12, '用户名不能超过 12 个字符'),
  display_name: z.string().trim().max(20, '显示名称不能超过 20 个字符'),
  password: z
    .string()
    .max(20, '密码不能超过 20 个字符')
    .refine((value) => value.length === 0 || value.length >= 8, '密码至少需要 8 个字符'),
});

type UserFormValues = z.infer<typeof userSchema>;

type FeedbackState = {
  tone: 'info' | 'success' | 'danger';
  message: string;
};

const defaultValues: UserFormValues = {
  username: '',
  display_name: '',
  password: '',
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败，请稍后重试。';
}

function getRoleMeta(role: number) {
  if (role >= 100) {
    return { label: '超级管理员', variant: 'warning' as const };
  }

  if (role >= 10) {
    return { label: '管理员', variant: 'info' as const };
  }

  return { label: '普通用户', variant: 'success' as const };
}

function getStatusMeta(status: number) {
  if (status === 1) {
    return { label: '已激活', variant: 'success' as const };
  }

  return { label: '已封禁', variant: 'danger' as const };
}

function toPayload(values: UserFormValues): UserMutationPayload {
  return {
    username: values.username.trim(),
    display_name: values.display_name.trim(),
    password: values.password,
  };
}

export function UsersPage() {
  const queryClient = useQueryClient();
  const { user: currentUser } = useAuth();
  const [page, setPage] = useState(0);
  const [searchInput, setSearchInput] = useState('');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [editingUserId, setEditingUserId] = useState<number | null>(null);
  const [isEditorOpen, setIsEditorOpen] = useState(false);
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);

  const form = useForm<UserFormValues>({
    resolver: zodResolver(userSchema),
    defaultValues,
  });

  const isAdmin = (currentUser?.role ?? 0) >= 10;
  const isRoot = (currentUser?.role ?? 0) >= 100;

  const usersQuery = useQuery({
    queryKey: [...usersListQueryKey, page],
    queryFn: () => getUsers(page),
    enabled: isAdmin && searchKeyword.length === 0,
  });

  const searchQuery = useQuery({
    queryKey: [...usersListQueryKey, 'search', searchKeyword],
    queryFn: () => searchUsers(searchKeyword),
    enabled: isAdmin && searchKeyword.length > 0,
  });

  const editingUserQuery = useQuery({
    queryKey: ['users', 'detail', editingUserId],
    queryFn: () => getUser(editingUserId as number),
    enabled: isAdmin && editingUserId !== null,
  });

  useEffect(() => {
    if (editingUserQuery.data) {
      form.reset({
        username: editingUserQuery.data.username,
        display_name: editingUserQuery.data.display_name || '',
        password: '',
      });
    }
  }, [editingUserQuery.data, form]);

  const saveMutation = useMutation({
    mutationFn: async (values: UserFormValues) => {
      if (!editingUserId && values.password.length === 0) {
        throw new Error('创建用户时必须填写密码。');
      }

      const payload = toPayload(values);

      if (editingUserId) {
        await updateUser({ id: editingUserId, ...payload });
        return '用户信息已更新。';
      }

      await createUser(payload);
      return '用户已创建。';
    },
    onSuccess: async (message) => {
      setFeedback({ tone: 'success', message });
      setEditingUserId(null);
      setIsEditorOpen(false);
      form.reset(defaultValues);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: usersListQueryKey }),
        queryClient.invalidateQueries({ queryKey: ['users', 'detail'] }),
      ]);
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const manageMutation = useMutation({
    mutationFn: async ({ username, action }: { username: string; action: ManageUserAction }) => {
      await manageUser(username, action);
      return action;
    },
    onSuccess: async (action) => {
      const actionLabel = {
        promote: '提升权限',
        demote: '降级权限',
        delete: '删除用户',
        disable: '禁用用户',
        enable: '启用用户',
      }[action];

      setFeedback({ tone: 'success', message: `${actionLabel}操作已完成。` });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: usersListQueryKey }),
        queryClient.invalidateQueries({ queryKey: ['users', 'detail'] }),
      ]);
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const activeQuery = searchKeyword.length > 0 ? searchQuery : usersQuery;
  const users = useMemo(() => activeQuery.data ?? [], [activeQuery.data]);

  const handleSearchSubmit = () => {
    setFeedback(null);
    setPage(0);
    setSearchKeyword(searchInput.trim());
  };

  const handleResetSearch = () => {
    setSearchInput('');
    setSearchKeyword('');
    setPage(0);
    setFeedback(null);
  };

  const handleResetForm = () => {
    setEditingUserId(null);
    setIsEditorOpen(false);
    setFeedback(null);
    form.reset(defaultValues);
  };

  const handleCreate = () => {
    setEditingUserId(null);
    setFeedback(null);
    form.reset(defaultValues);
    setIsEditorOpen(true);
  };

  const handleEdit = (targetUser: UserItem) => {
    setFeedback(null);
    setEditingUserId(targetUser.id);
    setIsEditorOpen(true);
  };

  const handleManage = (targetUser: UserItem, action: ManageUserAction) => {
    const actionText = {
      promote: '提升',
      demote: '降级',
      delete: '删除',
      disable: '禁用',
      enable: '启用',
    }[action];

    if (action === 'delete' && !window.confirm(`确认删除用户“${targetUser.username}”吗？`)) {
      return;
    }

    if ((action === 'disable' || action === 'enable') && !window.confirm(`确认${actionText}用户“${targetUser.username}”吗？`)) {
      return;
    }

    setFeedback(null);
    manageMutation.mutate({ username: targetUser.username, action });
  };

  const handleSubmit = form.handleSubmit((values) => {
    setFeedback(null);
    saveMutation.mutate(values);
  });

  if (!isAdmin) {
    return (
      <div className='space-y-6'>
        <PageHeader
          title='用户管理'
          description='阶段 4 已迁移用户管理页面，但当前账户没有管理员权限。'
        />
        <EmptyState
          title='权限不足'
          description='用户列表、编辑与账户管理动作仅对管理员开放，请使用管理员账号登录后继续。'
        />
      </div>
    );
  }

  return (
    <>
    <div className='space-y-6'>
      <PageHeader
        title='用户管理'
        description='支持搜索、分页、创建、编辑以及启用、封禁、权限调整等常见账户管理动作。'
        action={
          <PrimaryButton type='button' onClick={handleCreate}>
            新建用户
          </PrimaryButton>
        }
      />

      {feedback ? <InlineMessage tone={feedback.tone} message={feedback.message} /> : null}

      <AppCard
        title='用户列表'
        action={
          <div className='flex flex-wrap gap-2'>
            <SecondaryButton
              type='button'
              onClick={() => void queryClient.invalidateQueries({ queryKey: usersListQueryKey })}
            >
              刷新
            </SecondaryButton>
          </div>
        }
      >
        <div className='space-y-5'>
          <div className='flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between'>
            <div className='flex w-full flex-col gap-3 md:flex-row'>
              <ResourceInput
                value={searchInput}
                onChange={(event) => setSearchInput(event.target.value)}
                placeholder='搜索用户 ID、用户名、显示名称或邮箱地址'
                className='md:max-w-xl'
              />
              <div className='flex flex-wrap gap-2'>
                <PrimaryButton type='button' onClick={handleSearchSubmit}>
                  搜索
                </PrimaryButton>
                <SecondaryButton type='button' onClick={handleResetSearch}>
                  清空
                </SecondaryButton>
              </div>
            </div>

            {searchKeyword.length === 0 ? (
              <div className='flex flex-wrap gap-2'>
                <SecondaryButton type='button' onClick={() => setPage((value) => Math.max(value - 1, 0))} disabled={page === 0 || activeQuery.isLoading}>
                  上一页
                </SecondaryButton>
                <SecondaryButton
                  type='button'
                  onClick={() => setPage((value) => value + 1)}
                  disabled={activeQuery.isLoading || users.length < ITEMS_PER_PAGE}
                >
                  下一页
                </SecondaryButton>
              </div>
            ) : null}
          </div>

          {activeQuery.isLoading ? (
            <LoadingState />
          ) : activeQuery.isError ? (
            <ErrorState title='用户列表加载失败' description={getErrorMessage(activeQuery.error)} />
          ) : users.length === 0 ? (
            <EmptyState title='暂无用户' description='当前条件下没有可展示的用户记录。' />
          ) : (
            <div className='overflow-x-auto'>
              <table className='min-w-full divide-y divide-[var(--border-default)] text-left text-sm'>
                <thead>
                  <tr className='text-[var(--foreground-secondary)]'>
                    <th className='px-3 py-3 font-medium'>用户</th>
                    <th className='px-3 py-3 font-medium'>邮箱</th>
                    <th className='px-3 py-3 font-medium'>角色</th>
                    <th className='px-3 py-3 font-medium'>状态</th>
                    <th className='px-3 py-3 font-medium'>操作</th>
                  </tr>
                </thead>
                <tbody className='divide-y divide-[var(--border-default)]'>
                  {users.map((targetUser) => {
                    const roleMeta = getRoleMeta(targetUser.role);
                    const statusMeta = getStatusMeta(targetUser.status);
                    const canManage = (currentUser?.role ?? 0) > targetUser.role;

                    return (
                      <tr key={targetUser.id} className='align-top'>
                        <td className='px-3 py-4'>
                          <div className='space-y-1'>
                            <p className='font-medium text-[var(--foreground-primary)]'>{targetUser.username}</p>
                            <p className='text-xs text-[var(--foreground-secondary)]'>
                              显示名称：{targetUser.display_name || '未设置'}
                            </p>
                            <p className='text-xs text-[var(--foreground-secondary)]'>ID：{targetUser.id}</p>
                          </div>
                        </td>
                        <td className='px-3 py-4 text-[var(--foreground-secondary)]'>
                          {targetUser.email || '未绑定'}
                        </td>
                        <td className='px-3 py-4'>
                          <StatusBadge label={roleMeta.label} variant={roleMeta.variant} />
                        </td>
                        <td className='px-3 py-4'>
                          <StatusBadge label={statusMeta.label} variant={statusMeta.variant} />
                        </td>
                        <td className='px-3 py-4'>
                          <div className='flex flex-wrap gap-2'>
                            <SecondaryButton type='button' onClick={() => handleEdit(targetUser)} className='px-3 py-2 text-xs' disabled={!canManage}>
                              编辑
                            </SecondaryButton>
                            <SecondaryButton
                              type='button'
                              onClick={() => handleManage(targetUser, targetUser.status === 1 ? 'disable' : 'enable')}
                              className='px-3 py-2 text-xs'
                              disabled={!canManage || manageMutation.isPending}
                            >
                              {targetUser.status === 1 ? '禁用' : '启用'}
                            </SecondaryButton>
                            <SecondaryButton
                              type='button'
                              onClick={() => handleManage(targetUser, targetUser.role >= 10 ? 'demote' : 'promote')}
                              className='px-3 py-2 text-xs'
                              disabled={!canManage || manageMutation.isPending || (!isRoot && targetUser.role < 10)}
                            >
                              {targetUser.role >= 10 ? '降级' : '提升'}
                            </SecondaryButton>
                            <DangerButton
                              type='button'
                              onClick={() => handleManage(targetUser, 'delete')}
                              className='px-3 py-2 text-xs'
                              disabled={!canManage || manageMutation.isPending}
                            >
                              删除
                            </DangerButton>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </AppCard>
    </div>
    <AppModal
      isOpen={isEditorOpen}
      onClose={handleResetForm}
      title={editingUserId ? '编辑用户' : '创建新用户'}
      description={editingUserId ? '可更新用户名、显示名称与密码。留空密码表示不修改。' : '新建用户默认创建为普通用户。'}
      footer={
        <div className='flex flex-wrap justify-end gap-3'>
          <SecondaryButton type='button' onClick={handleResetForm} disabled={saveMutation.isPending}>
            取消
          </SecondaryButton>
          <PrimaryButton type='submit' form='user-editor-form' disabled={saveMutation.isPending}>
            {saveMutation.isPending ? '保存中...' : editingUserId ? '保存修改' : '创建用户'}
          </PrimaryButton>
        </div>
      }
    >
      {editingUserId && editingUserQuery.isLoading ? (
        <LoadingState />
      ) : editingUserId && editingUserQuery.isError ? (
        <ErrorState title='用户详情加载失败' description={getErrorMessage(editingUserQuery.error)} />
      ) : (
        <form id='user-editor-form' className='space-y-5' onSubmit={handleSubmit}>
          <ResourceField label='用户名' error={form.formState.errors.username?.message}>
            <ResourceInput placeholder='请输入用户名' {...form.register('username')} />
          </ResourceField>

          <ResourceField label='显示名称' error={form.formState.errors.display_name?.message}>
            <ResourceInput placeholder='请输入显示名称' {...form.register('display_name')} />
          </ResourceField>

          <ResourceField
            label={editingUserId ? '新密码' : '密码'}
            hint={editingUserId ? '留空表示保持原密码不变。' : '密码长度需为 8 到 20 个字符。'}
            error={form.formState.errors.password?.message}
          >
            <ResourceInput type='password' placeholder='请输入密码' {...form.register('password')} />
          </ResourceField>

          {editingUserId && editingUserQuery.data ? (
            <div className='grid gap-4 md:grid-cols-2'>
              <div className='rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4'>
                <p className='text-xs uppercase tracking-[0.2em] text-[var(--foreground-muted)]'>绑定邮箱</p>
                <p className='mt-2 break-all text-sm text-[var(--foreground-primary)]'>
                  {editingUserQuery.data.email || '未绑定'}
                </p>
              </div>
              <div className='rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-4'>
                <p className='text-xs uppercase tracking-[0.2em] text-[var(--foreground-muted)]'>第三方账号</p>
                <p className='mt-2 text-sm text-[var(--foreground-primary)]'>
                  GitHub：{editingUserQuery.data.github_id || '未绑定'}
                </p>
                <p className='mt-1 text-sm text-[var(--foreground-primary)]'>
                  微信：{editingUserQuery.data.wechat_id || '未绑定'}
                </p>
              </div>
            </div>
          ) : null}
        </form>
      )}
    </AppModal>
    </>
  );
}
