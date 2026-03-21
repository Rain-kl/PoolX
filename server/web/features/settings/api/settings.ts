import { ApiError, apiRequest, getApiUrl } from '@/lib/api/client';
import type { ApiEnvelope } from '@/types/api';

import type {
  GeoIPPreviewResult,
  KernelBinaryInstallResult,
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

export function uploadMihomoBinary(
  binary: File,
  installPath: string,
  onProgress?: (progress: number) => void,
) {
  const formData = new FormData();
  formData.append('binary', binary);
  formData.append('install_path', installPath);

  if (!onProgress) {
    return apiRequest<KernelBinaryInstallResult>('/kernel/mihomo/upload', {
      method: 'POST',
      body: formData,
    });
  }

  return new Promise<KernelBinaryInstallResult>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', getApiUrl('/kernel/mihomo/upload'));
    xhr.withCredentials = true;

    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    });

    xhr.addEventListener('load', () => {
      let payload: ApiEnvelope<KernelBinaryInstallResult> | null = null;
      try {
        payload = JSON.parse(
          xhr.responseText,
        ) as ApiEnvelope<KernelBinaryInstallResult>;
      } catch {
        payload = null;
      }
      if (xhr.status < 200 || xhr.status >= 300) {
        reject(
          new ApiError(
            payload?.message || `请求失败（${xhr.status}）`,
            xhr.status,
          ),
        );
        return;
      }
      if (!payload) {
        reject(new ApiError('响应格式无效', xhr.status));
        return;
      }
      if (!payload.success) {
        reject(new ApiError(payload.message || '请求失败', xhr.status));
        return;
      }
      resolve(payload.data);
    });

    xhr.addEventListener('error', () => {
      reject(new ApiError('上传过程中网络连接中断，请检查网络后重试', 0));
    });

    xhr.send(formData);
  });
}

export function inspectMihomoBinary(installPath: string) {
  return apiRequest<KernelBinaryInstallResult>('/kernel/mihomo/inspect', {
    method: 'POST',
    body: JSON.stringify({ install_path: installPath }),
  });
}

export function downloadMihomoBinary(installPath: string) {
  return apiRequest<KernelBinaryInstallResult>('/kernel/mihomo/download', {
    method: 'POST',
    body: JSON.stringify({ install_path: installPath }),
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
