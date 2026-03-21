import { apiRequest } from '@/lib/api/client';
import type { PublicStatus } from '@/types/public-status';

export function getPublicStatus() {
  return apiRequest<PublicStatus>('/status');
}
