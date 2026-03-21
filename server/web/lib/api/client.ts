import { publicEnv } from '@/lib/env/public-env';
import type { ApiEnvelope } from '@/types/api';

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export function getApiUrl(path: string) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${publicEnv.apiBaseUrl}${normalizedPath}`;
}

export async function apiRequest<T>(path: string, init?: RequestInit) {
  const headers = new Headers(init?.headers ?? {});
  const method = init?.method?.toUpperCase() ?? 'GET';

  if (!(init?.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(getApiUrl(path), {
    credentials: 'include',
    headers,
    cache: method === 'GET' ? 'no-store' : init?.cache,
    ...init,
  });

  let payload: ApiEnvelope<T> | null = null;

  try {
    payload = (await response.json()) as ApiEnvelope<T>;
  } catch {
    payload = null;
  }

  if (!response.ok) {
    throw new ApiError(
      payload?.message || `请求失败（${response.status}）`,
      response.status,
    );
  }

  if (!payload) {
    throw new ApiError('响应格式无效', response.status);
  }

  if (!payload.success) {
    throw new ApiError(payload.message || '请求失败', response.status);
  }

  return payload.data;
}
