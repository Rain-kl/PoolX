import { pushLog } from '@/features/logs/api/logs';
import type { LogClassification, LogLevel } from '@/features/logs/types';

export class AppLog {
  static async push(classification: LogClassification, level: LogLevel, message: string) {
    if (!message.trim()) {
      return;
    }

    try {
      await pushLog(classification, level, message);
    } catch (error) {
      console.warn('push app log failed', error);
    }
  }
}
