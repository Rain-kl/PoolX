import { Suspense } from 'react';

import { LoadingState } from '@/components/feedback/loading-state';
import { GitHubOAuthCallback } from '@/features/auth/components/github-oauth-callback';

export default function GithubOAuthPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <GitHubOAuthCallback />
    </Suspense>
  );
}
