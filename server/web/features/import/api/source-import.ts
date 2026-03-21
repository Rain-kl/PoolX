import { apiRequest } from '@/lib/api/client';

import type {
  SourceImportResult,
  SourceParseResult,
} from '@/features/import/types';

export function parseSourceConfig(file: File) {
  const formData = new FormData();
  formData.append('file', file);

  return apiRequest<SourceParseResult>('/source-configs/parse', {
    method: 'POST',
    body: formData,
  });
}

export function importSourceConfig(sourceConfigId: number) {
  return apiRequest<SourceImportResult>('/source-configs/import', {
    method: 'POST',
    body: JSON.stringify({ source_config_id: sourceConfigId }),
  });
}

