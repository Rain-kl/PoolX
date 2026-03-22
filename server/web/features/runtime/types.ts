export interface KernelInstanceItem {
  id: number;
  kernel_type: string;
  status: string;
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

export interface RuntimeListenerItem {
  profile_id: number;
  profile_name: string;
  name: string;
  type: string;
  listen: string;
  port: number;
  proxy_group_name: string;
}

export interface RuntimeStatus {
  instance: KernelInstanceItem;
  running: boolean;
  api_healthy: boolean;
  api_version?: string;
  profile_count: number;
  listener_count: number;
  listeners: RuntimeListenerItem[];
  rendered_config_preview?: string;
}

export interface RuntimeLogItem {
  seq: number;
  stream: string;
  level: string;
  message: string;
  created_at: string;
}

export interface RuntimeLogList {
  items: RuntimeLogItem[];
}
