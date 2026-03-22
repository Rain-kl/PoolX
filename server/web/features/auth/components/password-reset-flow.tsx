'use client';

import { useSearchParams } from 'next/navigation';

import { PasswordResetConfirmForm } from '@/features/auth/components/password-reset-confirm-form';
import { PasswordResetRequestForm } from '@/features/auth/components/password-reset-request-form';

export function PasswordResetFlow() {
  const searchParams = useSearchParams();
  const hasConfirmationParams = Boolean(searchParams?.get('email') && searchParams?.get('token'));

  return hasConfirmationParams ? <PasswordResetConfirmForm /> : <PasswordResetRequestForm />;
}