import { apiRequest } from '@/lib/api/client';

import type { KernelCapability } from '@/features/capability/types';

export function getKernelCapability() {
  return apiRequest<KernelCapability>('/capabilities');
}
