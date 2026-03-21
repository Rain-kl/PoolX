export type PortProfileStrategy =
  | 'select'
  | 'url-test'
  | 'fallback'
  | 'load-balance';

export interface RuntimeConfigItem {
  id: number;
  port_profile_id: number;
  kernel_type: string;
  checksum: string;
  rendered_config: string;
  created_at: string;
  updated_at: string;
}

export interface ProxyNodeOption {
  id: number;
  name: string;
  type: string;
  server: string;
  port: number;
  tags: string;
  source_config_name: string;
  enabled: boolean;
  last_test_status: string;
}

export interface PortProfileRecord {
  id: number;
  name: string;
  listen_host: string;
  mixed_port: number;
  socks_port: number;
  http_port: number;
  strategy_type: PortProfileStrategy;
  strategy_group_name: string;
  test_url: string;
  test_interval_seconds: number;
  enabled: boolean;
  include_in_runtime: boolean;
  kernel_type: string;
  created_at: string;
  updated_at: string;
}

export interface PortProfileWithNodes {
  profile: PortProfileRecord;
  node_ids: number[];
  nodes: ProxyNodeOption[];
  runtime?: RuntimeConfigItem;
}

export interface PortProfilePayload {
  name: string;
  listen_host: string;
  mixed_port: number;
  socks_port: number;
  http_port: number;
  strategy_type: PortProfileStrategy;
  strategy_group_name: string;
  test_url: string;
  test_interval_seconds: number;
  enabled: boolean;
  include_in_runtime: boolean;
  node_ids: number[];
}

export interface PortProfilePreview {
  profile: PortProfileRecord;
  node_ids: number[];
  nodes: ProxyNodeOption[];
  kernel_type: string;
  checksum: string;
  content: string;
}

export interface PortProfileTemplateRecord {
  id: number;
  name: string;
  listen_host: string;
  mixed_port: number;
  socks_port: number;
  http_port: number;
  strategy_type: PortProfileStrategy;
  strategy_group_name: string;
  test_url: string;
  test_interval_seconds: number;
  enabled: boolean;
  include_in_runtime: boolean;
  created_at: string;
  updated_at: string;
}

export interface PortProfileTemplateItem {
  template: PortProfileTemplateRecord;
  node_ids: number[];
}
