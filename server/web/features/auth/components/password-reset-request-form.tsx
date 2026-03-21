'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { InlineMessage } from '@/components/feedback/inline-message';
import { TurnstileWidget } from '@/components/forms/turnstile-widget';
import { AppCard } from '@/components/ui/app-card';
import { sendPasswordResetEmail } from '@/features/auth/api/auth';
import { getPublicStatus } from '@/features/auth/api/public';
import {
  AuthButton,
  AuthFormField,
  AuthInput,
} from '@/features/auth/components/auth-form-primitives';
import { PublicAuthGuard } from '@/features/auth/components/public-auth-guard';

const resetRequestSchema = z.object({
  email: z.string().email('请输入有效邮箱地址'),
});

type ResetRequestFormValues = z.infer<typeof resetRequestSchema>;

export function PasswordResetRequestForm() {
  const [turnstileToken, setTurnstileToken] = useState('');
  const [message, setMessage] = useState<{ tone: 'success' | 'danger' | 'info'; text: string } | null>(null);

  const form = useForm<ResetRequestFormValues>({
    resolver: zodResolver(resetRequestSchema),
    defaultValues: { email: '' },
  });

  const statusQuery = useQuery({
    queryKey: ['public-status'],
    queryFn: getPublicStatus,
  });

  const mutation = useMutation({
    mutationFn: (values: ResetRequestFormValues) =>
      sendPasswordResetEmail(values.email, turnstileToken || undefined),
    onSuccess: () => {
      setMessage({ tone: 'success', text: '重置邮件发送成功，请检查邮箱。' });
      form.reset();
    },
    onError: (error: Error) => {
      setMessage({ tone: 'danger', text: error.message || '重置邮件发送失败，请稍后重试。' });
    },
  });

  const handleSubmit = form.handleSubmit((values) => {
    setMessage(null);
    if (statusQuery.data?.turnstile_check && !turnstileToken) {
      setMessage({ tone: 'info', text: '请先完成人机验证。' });
      return;
    }
    mutation.mutate(values);
  });

  return (
    <PublicAuthGuard>
      <AppCard title='密码重置' description='提交后，系统会向你的注册邮箱发送重置链接。'>
        <form className='space-y-4' onSubmit={handleSubmit}>
          <AuthFormField label='邮箱地址'>
            <AuthInput type='email' placeholder='请输入邮箱地址' {...form.register('email')} />
            {form.formState.errors.email ? (
              <span className='text-xs text-[var(--status-danger-foreground)]'>
                {form.formState.errors.email.message}
              </span>
            ) : null}
          </AuthFormField>

          {statusQuery.data?.turnstile_check && statusQuery.data.turnstile_site_key ? (
            <TurnstileWidget
              siteKey={statusQuery.data.turnstile_site_key}
              onVerify={(token) => setTurnstileToken(token)}
              onExpire={() => setTurnstileToken('')}
              onError={() => setTurnstileToken('')}
            />
          ) : null}

          {message ? <InlineMessage tone={message.tone} message={message.text} /> : null}

          <AuthButton type='submit' disabled={mutation.isPending}>
            {mutation.isPending ? '提交中...' : '发送重置邮件'}
          </AuthButton>
        </form>

        <div className='mt-6 text-sm text-[var(--foreground-secondary)]'>
          想起密码了？
          <Link href='/login' className='ml-2 text-[var(--brand-primary)] transition hover:opacity-80'>
            返回登录
          </Link>
        </div>
      </AppCard>
    </PublicAuthGuard>
  );
}
