import { apiRequest } from '@/lib/api/client';

import type {
  NodeTestExecution,
  ProxyNodeItem,
} from '@/features/nodes/types';

export function getProxyNodes(params: {
  page: number;
  keyword?: string;
  enabled?: string;
}) {
  const searchParams = new URLSearchParams();
  searchParams.set('p', String(params.page));
  if (params.keyword?.trim()) {
    searchParams.set('keyword', params.keyword.trim());
  }
  if (params.enabled && params.enabled !== 'all') {
    searchParams.set('enabled', params.enabled);
  }

  return apiRequest<ProxyNodeItem[]>(`/proxy-nodes?${searchParams.toString()}`);
}

export function updateProxyNodeStatus(id: number, enabled: boolean) {
  return apiRequest<void>(`/proxy-nodes/${id}/status`, {
    method: 'POST',
    body: JSON.stringify({ enabled }),
  });
}

export function deleteProxyNode(id: number) {
  return apiRequest<void>(`/proxy-nodes/${id}/delete`, {
    method: 'POST',
  });
}

export function deleteProxyNodes(nodeIds: number[]) {
  return apiRequest<{ deleted: number }>('/proxy-nodes/delete', {
    method: 'POST',
    body: JSON.stringify({ node_ids: nodeIds }),
  });
}

export function testProxyNodes(input: {
  nodeIds: number[];
  timeoutMs: number;
  testUrl: string;
}) {
  return apiRequest<NodeTestExecution[]>('/proxy-nodes/test', {
    method: 'POST',
    body: JSON.stringify({
      node_ids: input.nodeIds,
      timeout_ms: input.timeoutMs,
      test_url: input.testUrl,
    }),
  });
}
