import { apiRequest } from '@/lib/api/client';
import { AppLog } from '@/lib/logging/app-log';

import type { FileItem } from '@/features/files/types';

export function getFiles(page: number) {
  return apiRequest<FileItem[]>(`/file/?p=${page}`);
}

export function searchFiles(keyword: string) {
  return apiRequest<FileItem[]>(
    `/file/search?keyword=${encodeURIComponent(keyword)}`,
  );
}

export function uploadFiles(files: File[], description: string) {
  const formData = new FormData();
  for (const file of files) {
    formData.append('file', file);
  }
  formData.append('description', description);

  return apiRequest<void>('/file/', {
    method: 'POST',
    body: formData,
  }).then((result) => {
    void AppLog.push('business', 'info', `frontend file upload completed | files=${files.length} | description=${description || '—'}`);
    return result;
  });
}

export function deleteFile(id: number) {
  return apiRequest<void>(`/file/${id}/delete`, {
    method: 'POST',
  }).then((result) => {
    void AppLog.push('business', 'info', `frontend file delete completed | file_id=${id}`);
    return result;
  });
}
