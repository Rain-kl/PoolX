import { apiRequest } from '@/lib/api/client';

import type { AppLogList, LogClassification, LogLevel } from '@/features/logs/types';

type GetLogsParams = {
  classification?: LogClassification | 'all';
  afterId?: number;
  limit?: number;
};

export function getLogs({ classification = 'all', afterId = 0, limit = 100 }: GetLogsParams = {}) {
  const searchParams = new URLSearchParams({
    limit: String(limit),
  });

  if (afterId > 0) {
    searchParams.set('after_id', String(afterId));
  }

  if (classification !== 'all') {
    searchParams.set('classification', classification);
  }

  return apiRequest<AppLogList>(`/log/?${searchParams.toString()}`);
}

export function pushLog(classification: LogClassification, level: LogLevel, message: string) {
  return apiRequest<void>('/log/', {
    method: 'POST',
    body: JSON.stringify({
      classification,
      level,
      message,
    }),
  });
}
