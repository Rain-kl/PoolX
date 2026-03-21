import { ApiError, apiRequest, getApiUrl } from '@/lib/api/client';
import type { ApiEnvelope } from '@/types/api';

import type {
  LatestReleaseInfo,
  ReleaseChannel,
  UpgradeStreamSnapshot,
  UploadedServerBinaryInfo,
} from '@/features/update/types';

export function getLatestRelease(channel: ReleaseChannel = 'stable') {
  return apiRequest<LatestReleaseInfo>(
    `/update/latest-release?channel=${channel}`,
  );
}

export function upgradeServer(channel: ReleaseChannel = 'stable') {
  return apiRequest<LatestReleaseInfo>('/update/upgrade', {
    method: 'POST',
    body: JSON.stringify({ channel }),
  });
}

export function uploadServerBinary(
  binary: File,
  onProgress?: (progress: number) => void,
): Promise<UploadedServerBinaryInfo> {
  const formData = new FormData();
  formData.append('binary', binary);

  if (!onProgress) {
    return apiRequest<UploadedServerBinaryInfo>('/update/manual-upload', {
      method: 'POST',
      body: formData,
    });
  }

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', getApiUrl('/update/manual-upload'));
    xhr.withCredentials = true;

    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    });

    xhr.addEventListener('load', () => {
      let payload: ApiEnvelope<UploadedServerBinaryInfo> | null = null;
      try {
        payload = JSON.parse(
          xhr.responseText,
        ) as ApiEnvelope<UploadedServerBinaryInfo>;
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

export function confirmManualServerUpgrade(uploadToken: string) {
  return apiRequest<UploadedServerBinaryInfo>('/update/manual-upgrade', {
    method: 'POST',
    body: JSON.stringify({ upload_token: uploadToken }),
  });
}

export function createUpgradeLogsWebSocket() {
  if (typeof window === 'undefined') {
    return null;
  }

  const apiUrl = getApiUrl('/update/logs/ws');
  const resolvedUrl = apiUrl.startsWith('http://')
    ? `ws://${apiUrl.slice('http://'.length)}`
    : apiUrl.startsWith('https://')
      ? `wss://${apiUrl.slice('https://'.length)}`
      : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${apiUrl}`;

  return new WebSocket(resolvedUrl);
}

export function parseUpgradeStreamSnapshot(
  rawMessage: string,
): UpgradeStreamSnapshot | null {
  try {
    const parsed = JSON.parse(rawMessage) as Partial<UpgradeStreamSnapshot>;
    if (
      typeof parsed.in_progress !== 'boolean' ||
      typeof parsed.upgrade_status !== 'string' ||
      !Array.isArray(parsed.upgrade_logs)
    ) {
      return null;
    }
    return {
      in_progress: parsed.in_progress,
      upgrade_status: parsed.upgrade_status,
      upgrade_logs: parsed.upgrade_logs,
    };
  } catch {
    return null;
  }
}
