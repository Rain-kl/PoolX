import { apiRequest } from '@/lib/api/client';

import type {
  PortProfilePayload,
  PortProfilePreview,
  PortProfileTemplateItem,
  PortProfileWithNodes,
  ProxyNodeOption,
  RuntimeConfigItem,
} from '@/features/workspace/types';

export function getPortProfiles() {
  return apiRequest<PortProfileWithNodes[]>('/port-profiles');
}

export function getPortProfile(id: number) {
  return apiRequest<PortProfileWithNodes>(`/port-profiles/${id}`);
}

export function createPortProfile(payload: PortProfilePayload) {
  return apiRequest<PortProfileWithNodes>('/port-profiles', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updatePortProfile(id: number, payload: PortProfilePayload) {
  return apiRequest<PortProfileWithNodes>(`/port-profiles/${id}`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function deletePortProfile(id: number) {
  return apiRequest<void>(`/port-profiles/${id}/delete`, {
    method: 'POST',
  });
}

export function previewPortProfile(payload: PortProfilePayload) {
  return apiRequest<PortProfilePreview>('/port-profiles/preview', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function previewSavedPortProfile(id: number) {
  return apiRequest<PortProfilePreview>(`/port-profiles/${id}/preview`);
}

export function saveRuntimeConfig(id: number) {
  return apiRequest<RuntimeConfigItem>(`/port-profiles/${id}/runtime/save`, {
    method: 'POST',
  });
}

export function getProxyNodeOptions(keyword = '') {
  const searchParams = new URLSearchParams();
  if (keyword.trim()) {
    searchParams.set('keyword', keyword.trim());
  }
  const query = searchParams.toString();
  return apiRequest<ProxyNodeOption[]>(
    `/proxy-nodes/options${query ? `?${query}` : ''}`,
  );
}

export function getPortProfileTemplates() {
  return apiRequest<PortProfileTemplateItem[]>('/port-profile-templates');
}

export function savePortProfileTemplate(name: string, payload: PortProfilePayload) {
  return apiRequest<PortProfileTemplateItem>('/port-profile-templates', {
    method: 'POST',
    body: JSON.stringify({ name, payload }),
  });
}

export function deletePortProfileTemplate(id: number) {
  return apiRequest<void>(`/port-profile-templates/${id}/delete`, {
    method: 'POST',
  });
}
