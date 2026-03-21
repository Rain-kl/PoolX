import type { AuthUser } from '@/types/auth';

export interface OptionItem {
  key: string;
  value: string;
}

export interface GeoIPPreviewResult {
  provider: string;
  ip: string;
  iso_code: string;
  name: string;
  latitude?: number;
  longitude?: number;
}

export interface KernelBinaryInstallResult {
  kernel_type: string;
  install_path: string;
  binary_source: 'upload' | 'download' | string;
  detected_version: string;
  file_name: string;
  release_tag?: string;
  installed_at: string;
}

export interface UpdateSelfPayload {
  username: string;
  display_name: string;
  password: string;
}

export type SettingsProfile = AuthUser;
