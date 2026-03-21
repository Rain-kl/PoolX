import { apiRequest } from '@/lib/api/client';

import type {
  GeoIPPreviewResult,
  OptionItem,
  SettingsProfile,
  UpdateSelfPayload,
} from '@/features/settings/types';

export function getOptions() {
  return apiRequest<OptionItem[]>('/option/');
}

export function updateOption(key: string, value: string) {
  return apiRequest<void>('/option/update', {
    method: 'POST',
    body: JSON.stringify({ key, value }),
  });
}

export function previewGeoIP(provider: string, ip: string) {
  return apiRequest<GeoIPPreviewResult>('/option/geoip/preview', {
    method: 'POST',
    body: JSON.stringify({ provider, ip }),
  });
}

export function getSettingsProfile() {
  return apiRequest<SettingsProfile>('/user/self');
}

export function updateSelf(payload: UpdateSelfPayload) {
  return apiRequest<void>('/user/self/update', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function generateAccessToken() {
  return apiRequest<string>('/user/token');
}

export function bindWeChat(code: string) {
  return apiRequest<void>(
    `/oauth/wechat/bind?code=${encodeURIComponent(code)}`,
  );
}

export function bindEmail(email: string, code: string) {
  const searchParams = new URLSearchParams({ email, code });
  return apiRequest<void>(`/oauth/email/bind?${searchParams.toString()}`);
}

export function getAboutContent() {
  return apiRequest<string>('/about');
}
