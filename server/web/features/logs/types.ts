export type LogClassification = 'system' | 'business' | 'security';
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export interface AppLogItem {
  id: number;
  classification: string;
  level: string;
  message: string;
  created_at: string;
}

export interface AppLogList {
  items: AppLogItem[];
}
