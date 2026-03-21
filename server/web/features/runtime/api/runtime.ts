import { apiRequest } from '@/lib/api/client';

import type { RuntimeLogList, RuntimeStatus } from '@/features/runtime/types';

export function getRuntimeStatus() {
  return apiRequest<RuntimeStatus>('/runtime/status');
}

export function startRuntime() {
  return apiRequest<RuntimeStatus>('/runtime/start', {
    method: 'POST',
  });
}

export function stopRuntime() {
  return apiRequest<RuntimeStatus>('/runtime/stop', {
    method: 'POST',
  });
}

export function reloadRuntime() {
  return apiRequest<RuntimeStatus>('/runtime/reload', {
    method: 'POST',
  });
}

export function getRuntimeLogs(afterSeq = 0, limit = 100) {
  const searchParams = new URLSearchParams();
  if (afterSeq > 0) {
    searchParams.set('after_seq', String(afterSeq));
  }
  searchParams.set('limit', String(limit));
  return apiRequest<RuntimeLogList>(`/runtime/logs?${searchParams.toString()}`);
}
