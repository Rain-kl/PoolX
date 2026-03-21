import { apiRequest } from '@/lib/api/client';

import type {
  NodeTestExecution,
  NodeTestResultItem,
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

export function testProxyNodes(nodeIds: number[], timeoutMs = 3000) {
  return apiRequest<NodeTestExecution[]>('/proxy-nodes/test', {
    method: 'POST',
    body: JSON.stringify({
      node_ids: nodeIds,
      timeout_ms: timeoutMs,
    }),
  });
}

export function getNodeTestResults(proxyNodeId: number, limit = 10) {
  const searchParams = new URLSearchParams({
    proxy_node_id: String(proxyNodeId),
    limit: String(limit),
  });
  return apiRequest<NodeTestResultItem[]>(
    `/node-test-results?${searchParams.toString()}`,
  );
}

