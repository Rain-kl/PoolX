import { Suspense } from 'react';

import { LoadingState } from '@/components/feedback/loading-state';
import { PasswordResetFlow } from '@/features/auth/components/password-reset-flow';

export default function LegacyResetPasswordPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <PasswordResetFlow />
    </Suspense>
  );
}
