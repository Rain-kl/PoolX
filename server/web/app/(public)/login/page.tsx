import { Suspense } from 'react';

import { LoadingState } from '@/components/feedback/loading-state';
import { LoginForm } from '@/features/auth/components/login-form';

export default function LoginPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <LoginForm />
    </Suspense>
  );
}
