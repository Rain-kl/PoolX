export type ReleaseChannel = 'stable' | 'preview';

export interface UpgradeLogItem {
  level: string;
  message: string;
  created_at: string;
}

export interface LatestReleaseInfo {
  tag_name: string;
  body: string;
  html_url: string;
  published_at: string;
  channel: ReleaseChannel;
  prerelease: boolean;
  current_version: string;
  has_update: boolean;
  upgrade_supported: boolean;
  in_progress: boolean;
  upgrade_status: 'idle' | 'running' | 'succeeded' | 'failed' | string;
  upgrade_logs: UpgradeLogItem[];
}

export interface UpgradeStreamSnapshot {
  in_progress: boolean;
  upgrade_status: 'idle' | 'running' | 'succeeded' | 'failed' | string;
  upgrade_logs: UpgradeLogItem[];
}

export interface UploadedServerBinaryInfo {
  upload_token: string;
  file_name: string;
  detected_version: string;
  current_version: string;
  has_update: boolean;
  upgrade_supported: boolean;
  ready_to_upgrade: boolean;
  comparison_message: string;
  uploaded_at: string;
}
