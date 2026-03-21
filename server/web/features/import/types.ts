export interface SourceConfigRecord {
  id: number;
  filename: string;
  content_hash: string;
  status: string;
  total_nodes: number;
  valid_nodes: number;
  invalid_nodes: number;
  duplicate_nodes: number;
  imported_nodes: number;
  uploaded_by: string;
  uploaded_by_id: number;
  created_at: string;
  updated_at: string;
}

export interface ParsedNodePreview {
  name: string;
  type: string;
  server: string;
  port: number;
  fingerprint: string;
  duplicate_scope: 'none' | 'batch' | 'database';
}

export interface ParseIssue {
  index: number;
  name?: string;
  message: string;
}

export interface ParseSummary {
  total_nodes: number;
  valid_nodes: number;
  invalid_nodes: number;
  duplicate_nodes: number;
  importable_nodes: number;
}

export interface SourceParseResult {
  source_config: SourceConfigRecord;
  summary: ParseSummary;
  nodes: ParsedNodePreview[];
  errors: ParseIssue[];
}

export interface SourceImportResult {
  source_config_id: number;
  imported_nodes: number;
  skipped_nodes: number;
}

export interface ParsedNodeTestResult {
  node_id: number;
  node_name: string;
  status: string;
  latency_ms?: number;
  error_message?: string;
  test_url?: string;
  dial_address: string;
  started_at: string;
  finished_at: string;
  last_tested_at?: string;
}
