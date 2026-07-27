import { apiRequest } from '@/shared/api/client';
import type { ApiDecoder } from '@/shared/api/decoder';

export interface SourceConfig {
  id: number;
  source_type: 'upload' | 'subscription_url';
  source_url: string;
  content_type: string;
  fetched_at?: string;
  filename: string;
  content_hash: string;
  status: 'parsed' | 'imported';
  total_nodes: number;
  valid_nodes: number;
  invalid_nodes: number;
  duplicate_nodes: number;
  imported_nodes: number;
  uploaded_by: string;
  uploaded_by_id: number;
  created_at: string;
  updated_at: string;
}

export interface ParseIssue {
  index: number;
  name?: string;
  message: string;
}

export interface ParsedNode {
  name: string;
  type: string;
  server: string;
  port: number;
  fingerprint: string;
  metadata_json: string;
}

export interface ParseResult {
  nodes: ParsedNode[];
  issues: ParseIssue[];
}

export interface UploadSourceConfigResponse {
  config: SourceConfig;
  parse_result: ParseResult;
}

export interface ProxyNode {
  id: number;
  source_config_id: number;
  source_config_name: string;
  name: string;
  type: string;
  server: string;
  port: number;
  tags: string;
  metadata_json: string;
  enabled: boolean;
  last_test_status: 'unknown' | 'success' | 'failed';
  last_latency_ms?: number;
  last_test_error?: string;
  last_tested_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TestNodeResult {
  node_id: number;
  node_name: string;
  success: boolean;
  latency_ms: number;
  error_message: string;
}

export interface PortProfileProxySettings {
  strategy_type: 'select' | 'url-test' | 'fallback' | 'load-balance';
  test_url: string;
  test_interval_seconds: number;
  load_balance_strategy: 'consistent-hashing' | 'round-robin';
  load_balance_lazy?: boolean;
  load_balance_disable_udp?: boolean;
  udp_enabled: boolean;
  auth_enabled: boolean;
  auth_username?: string;
  auth_password?: string;
}

export interface PortProfile {
  id: number;
  name: string;
  listen_host: string;
  mixed_port: number;
  socks_port: number;
  http_port: number;
  proxy_settings: PortProfileProxySettings;
  include_in_runtime: boolean;
  kernel_type: string;
  created_at: string;
  updated_at: string;
}

export interface RuntimeConfig {
  id: number;
  port_profile_id: number;
  kernel_type: string;
  checksum: string;
  rendered_config: string;
  created_at: string;
  updated_at: string;
}

export interface PortProfileWithNodes {
  profile: PortProfile;
  node_ids: number[];
  nodes?: ProxyNode[];
  runtime_config?: RuntimeConfig;
}

export interface PortProfileTemplate {
  id: number;
  name: string;
  description: string;
  mixed_port: number;
  proxy_settings: PortProfileProxySettings;
  created_at: string;
  updated_at: string;
}

export interface KernelInstance {
  id: number;
  kernel_type: string;
  status:
    | 'stopped'
    | 'starting'
    | 'running'
    | 'stopping'
    | 'error'
    | 'reloading';
  pid?: number;
  work_dir: string;
  config_path: string;
  controller_address: string;
  active_config_checksum: string;
  active_profile_count: number;
  active_listener_count: number;
  last_action: string;
  last_error: string;
  last_started_at?: string;
  last_stopped_at?: string;
  last_reloaded_at?: string;
  created_at: string;
  updated_at: string;
}

export interface RuntimeStatusView {
  kernel_instance: KernelInstance;
  is_process_alive: boolean;
  controller_ok: boolean;
  mihomo_version: string;
  active_profiles: PortProfileWithNodes[];
}

export interface KernelCapability {
  kernel_type: string;
  supports_reload: boolean;
  supports_external_api: boolean;
  supports_health_check: boolean;
  supported_strategies: string[];
}

export interface RenderResult {
  kernel_type: string;
  checksum: string;
  content: string;
}

function decoder<T>(): ApiDecoder<T> {
  return (val: unknown) => val as T;
}

// --- API Methods ---

export async function uploadSourceConfig(
  filename: string,
  rawContent: string,
): Promise<UploadSourceConfigResponse> {
  return apiRequest<UploadSourceConfigResponse>(
    '/api/v1/admin/clash/source-configs/upload',
    {
      method: 'POST',
      body: { filename, raw_content: rawContent },
    },
    decoder<UploadSourceConfigResponse>(),
  );
}

export async function fetchSubscription(
  sourceUrl: string,
): Promise<UploadSourceConfigResponse> {
  return apiRequest<UploadSourceConfigResponse>(
    '/api/v1/admin/clash/source-configs/subscription',
    {
      method: 'POST',
      body: { source_url: sourceUrl },
    },
    decoder<UploadSourceConfigResponse>(),
  );
}

export async function listSourceConfigs(
  page = 1,
  pageSize = 20,
): Promise<{
  items: SourceConfig[];
  total: number;
  page: number;
  page_size: number;
}> {
  return apiRequest<{
    items: SourceConfig[];
    total: number;
    page: number;
    page_size: number;
  }>(
    `/api/v1/admin/clash/source-configs?page=${page}&page_size=${pageSize}`,
    {
      method: 'GET',
    },
    decoder<{
      items: SourceConfig[];
      total: number;
      page: number;
      page_size: number;
    }>(),
  );
}

export async function deleteSourceConfig(
  id: number,
): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(
    `/api/v1/admin/clash/source-configs/${id}`,
    {
      method: 'DELETE',
    },
    decoder<{ deleted: boolean }>(),
  );
}

export async function confirmSourceConfig(
  id: number,
): Promise<{ imported_nodes: number }> {
  return apiRequest<{ imported_nodes: number }>(
    `/api/v1/admin/clash/source-configs/${id}/confirm`,
    {
      method: 'POST',
    },
    decoder<{ imported_nodes: number }>(),
  );
}

export interface SyncSourceResult {
  id: number;
  total_nodes: number;
  valid_nodes: number;
  invalid_nodes: number;
  duplicate_nodes: number;
  imported_nodes: number;
}

export async function refreshSourceConfig(
  id: number,
): Promise<SyncSourceResult> {
  return apiRequest<SyncSourceResult>(
    `/api/v1/admin/clash/source-configs/${id}/refresh`,
    {
      method: 'POST',
    },
    decoder<SyncSourceResult>(),
  );
}

export async function reuploadSourceConfig(
  id: number,
  filename: string,
  rawContent: string,
): Promise<SyncSourceResult> {
  return apiRequest<SyncSourceResult>(
    `/api/v1/admin/clash/source-configs/${id}/reupload`,
    {
      method: 'POST',
      body: { filename, raw_content: rawContent },
    },
    decoder<SyncSourceResult>(),
  );
}

export async function listProxyNodes(params: {
  page?: number;
  pageSize?: number;
  keyword?: string;
  sourceConfigId?: number;
  enabled?: boolean;
}): Promise<{
  items: ProxyNode[];
  total: number;
  page: number;
  page_size: number;
}> {
  const query = new URLSearchParams();
  if (params.page) query.set('page', params.page.toString());
  if (params.pageSize) query.set('page_size', params.pageSize.toString());
  if (params.keyword) query.set('keyword', params.keyword);
  if (params.sourceConfigId)
    query.set('source_config_id', params.sourceConfigId.toString());
  if (params.enabled !== undefined)
    query.set('enabled', params.enabled ? 'true' : 'false');
  return apiRequest<{
    items: ProxyNode[];
    total: number;
    page: number;
    page_size: number;
  }>(
    `/api/v1/admin/clash/nodes?${query.toString()}`,
    {
      method: 'GET',
    },
    decoder<{
      items: ProxyNode[];
      total: number;
      page: number;
      page_size: number;
    }>(),
  );
}

export async function testProxyNodes(
  nodeIds: number[],
  binaryPath?: string,
  testUrl?: string,
  timeoutSeconds?: number,
): Promise<{ results: TestNodeResult[] }> {
  return apiRequest<{ results: TestNodeResult[] }>(
    '/api/v1/admin/clash/nodes/test',
    {
      method: 'POST',
      body: {
        node_ids: nodeIds,
        binary_path: binaryPath,
        test_url: testUrl,
        timeout_seconds: timeoutSeconds,
      },
    },
    decoder<{ results: TestNodeResult[] }>(),
  );
}

export async function deleteProxyNode(
  id: number,
): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(
    `/api/v1/admin/clash/nodes/${id}`,
    {
      method: 'DELETE',
    },
    decoder<{ deleted: boolean }>(),
  );
}

export async function deleteProxyNodesBatch(
  ids: number[],
): Promise<{ deleted_count: number }> {
  return apiRequest<{ deleted_count: number }>(
    '/api/v1/admin/clash/nodes/batch-delete',
    {
      method: 'POST',
      body: { ids },
    },
    decoder<{ deleted_count: number }>(),
  );
}

export async function toggleProxyNodesBatch(
  ids: number[],
  enabled: boolean,
): Promise<{ toggled_count: number; enabled: boolean }> {
  return apiRequest<{ toggled_count: number; enabled: boolean }>(
    '/api/v1/admin/clash/nodes/batch-toggle',
    {
      method: 'POST',
      body: { ids, enabled },
    },
    decoder<{ toggled_count: number; enabled: boolean }>(),
  );
}

export async function listPortProfiles(): Promise<{
  items: PortProfileWithNodes[];
}> {
  return apiRequest<{ items: PortProfileWithNodes[] }>(
    '/api/v1/admin/clash/port-profiles',
    {
      method: 'GET',
    },
    decoder<{ items: PortProfileWithNodes[] }>(),
  );
}

export async function createPortProfile(payload: {
  name: string;
  listen_host?: string;
  mixed_port: number;
  socks_port?: number;
  http_port?: number;
  proxy_settings: PortProfileProxySettings;
  include_in_runtime?: boolean;
  node_ids?: number[];
}): Promise<PortProfileWithNodes> {
  return apiRequest<PortProfileWithNodes>(
    '/api/v1/admin/clash/port-profiles',
    {
      method: 'POST',
      body: payload,
    },
    decoder<PortProfileWithNodes>(),
  );
}

export async function updatePortProfile(
  id: number,
  payload: {
    name: string;
    listen_host?: string;
    mixed_port: number;
    socks_port?: number;
    http_port?: number;
    proxy_settings: PortProfileProxySettings;
    include_in_runtime?: boolean;
    node_ids?: number[];
  },
): Promise<PortProfileWithNodes> {
  return apiRequest<PortProfileWithNodes>(
    `/api/v1/admin/clash/port-profiles/${id}`,
    {
      method: 'PUT',
      body: payload,
    },
    decoder<PortProfileWithNodes>(),
  );
}

export async function deletePortProfile(
  id: number,
): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(
    `/api/v1/admin/clash/port-profiles/${id}`,
    {
      method: 'DELETE',
    },
    decoder<{ deleted: boolean }>(),
  );
}

export async function previewPortProfile(id: number): Promise<RenderResult> {
  return apiRequest<RenderResult>(
    `/api/v1/admin/clash/port-profiles/${id}/preview`,
    {
      method: 'GET',
    },
    decoder<RenderResult>(),
  );
}

export async function listPortProfileTemplates(): Promise<{
  items: PortProfileTemplate[];
}> {
  return apiRequest<{ items: PortProfileTemplate[] }>(
    '/api/v1/admin/clash/port-profile-templates',
    {
      method: 'GET',
    },
    decoder<{ items: PortProfileTemplate[] }>(),
  );
}

export async function getRuntimeStatus(
  kernelType = 'mihomo',
): Promise<RuntimeStatusView> {
  return apiRequest<RuntimeStatusView>(
    `/api/v1/admin/clash/runtime/status?kernel_type=${kernelType}`,
    {
      method: 'GET',
    },
    decoder<RuntimeStatusView>(),
  );
}

export async function startKernel(payload: {
  kernel_type?: string;
  binary_path?: string;
  work_dir?: string;
  allow_lan?: boolean;
  mode?: string;
  controller_address?: string;
  controller_secret?: string;
}): Promise<KernelInstance> {
  return apiRequest<KernelInstance>(
    '/api/v1/admin/clash/runtime/start',
    {
      method: 'POST',
      body: payload,
    },
    decoder<KernelInstance>(),
  );
}

export async function stopKernel(
  kernelType = 'mihomo',
): Promise<{ stopped: boolean }> {
  return apiRequest<{ stopped: boolean }>(
    '/api/v1/admin/clash/runtime/stop',
    {
      method: 'POST',
      body: { kernel_type: kernelType },
    },
    decoder<{ stopped: boolean }>(),
  );
}

export async function reloadKernel(payload: {
  kernel_type?: string;
  allow_lan?: boolean;
  mode?: string;
  controller_address?: string;
  controller_secret?: string;
}): Promise<KernelInstance> {
  return apiRequest<KernelInstance>(
    '/api/v1/admin/clash/runtime/reload',
    {
      method: 'POST',
      body: payload,
    },
    decoder<KernelInstance>(),
  );
}

export async function getActiveRuntimeConfig(
  kernelType = 'mihomo',
): Promise<RenderResult> {
  return apiRequest<RenderResult>(
    `/api/v1/admin/clash/runtime/config?kernel_type=${kernelType}`,
    {
      method: 'GET',
    },
    decoder<RenderResult>(),
  );
}

export async function getKernelLogs(
  kernelType = 'mihomo',
): Promise<{ logs: string[] }> {
  return apiRequest<{ logs: string[] }>(
    `/api/v1/admin/clash/runtime/logs?kernel_type=${kernelType}`,
    {
      method: 'GET',
    },
    decoder<{ logs: string[] }>(),
  );
}

export type InstalledKernelBinary = {
  kernel_type: string;
  install_path: string;
  binary_source: string;
  detected_version: string;
  file_name: string;
  release_tag?: string;
  installed_at: string;
};

export async function getKernelCapabilities(): Promise<{
  capabilities: KernelCapability[];
}> {
  return apiRequest<{ capabilities: KernelCapability[] }>(
    '/api/v1/admin/clash/kernels/capabilities',
    {
      method: 'GET',
    },
    decoder<{ capabilities: KernelCapability[] }>(),
  );
}

export async function inspectKernelBinary(
  installPath?: string,
): Promise<InstalledKernelBinary> {
  return apiRequest<InstalledKernelBinary>(
    '/api/v1/admin/clash/kernels/inspect',
    {
      method: 'POST',
      body: { install_path: installPath },
    },
    decoder<InstalledKernelBinary>(),
  );
}

export async function downloadKernelBinary(
  installPath?: string,
): Promise<InstalledKernelBinary> {
  return apiRequest<InstalledKernelBinary>(
    '/api/v1/admin/clash/kernels/download',
    {
      method: 'POST',
      body: { install_path: installPath },
    },
    decoder<InstalledKernelBinary>(),
  );
}

export async function uploadKernelBinary(
  file: File,
  installPath?: string,
): Promise<InstalledKernelBinary> {
  const formData = new FormData();
  formData.append('binary', file);
  if (installPath) {
    formData.append('install_path', installPath);
  }
  return apiRequest<InstalledKernelBinary>(
    '/api/v1/admin/clash/kernels/upload',
    {
      method: 'POST',
      body: formData,
    },
    decoder<InstalledKernelBinary>(),
  );
}
