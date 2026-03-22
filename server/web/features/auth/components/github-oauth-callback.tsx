'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';

import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { useAuth } from '@/components/providers/auth-provider';
import { AppCard } from '@/components/ui/app-card';
import { exchangeGitHubCode } from '@/features/auth/api/auth';

export function GitHubOAuthCallback() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setUser } = useAuth();
  const [prompt, setPrompt] = useState('正在处理 GitHub 授权结果...');
  const [message, setMessage] = useState<{ tone: 'danger' | 'success'; text: string } | null>(null);

  const mutation = useMutation({
    mutationFn: exchangeGitHubCode,
    onSuccess: (user) => {
      setUser(user);
      setMessage({ tone: 'success', text: '登录成功，正在跳转...' });
      router.replace('/');
    },
    onError: (error: Error) => {
      setPrompt('授权处理失败');
      setMessage({ tone: 'danger', text: error.message || 'GitHub 授权失败，请稍后重试。' });
    },
  });

  useEffect(() => {
    const code = searchParams?.get('code');
    if (!code) {
      setPrompt('缺少授权 code');
      setMessage({ tone: 'danger', text: '未收到 GitHub 授权参数，请返回登录页重试。' });
      return;
    }

    mutation.mutate(code);
  }, [mutation, searchParams]);

  return (
    <AppCard title='GitHub OAuth 回调' description={prompt}>
      <div className='space-y-4'>
        {mutation.isPending ? <LoadingState /> : null}
        {message ? <InlineMessage tone={message.tone} message={message.text} /> : null}
      </div>
    </AppCard>
  );
}
