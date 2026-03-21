'use client';

import type { ReactNode } from 'react';
import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

import { LoadingState } from '@/components/feedback/loading-state';
import { useAuth } from '@/components/providers/auth-provider';

interface PublicAuthGuardProps {
  children: ReactNode;
}

export function PublicAuthGuard({ children }: PublicAuthGuardProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      const redirect = searchParams?.get('redirect');
      router.replace(redirect || '/');
    }
  }, [isAuthenticated, isLoading, router, searchParams]);

  if (isLoading) {
    return <LoadingState />;
  }

  if (isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
