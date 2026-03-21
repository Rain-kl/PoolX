import type { AppLogItem, LogClassification, LogLevel } from '@/features/logs/types';
import { formatDateTime, formatDateTimeWithMilliseconds } from '@/lib/utils/date';

type ClassificationMeta = { label: string; tone: string; dot: string };
type LevelMeta = { label: string };

const classificationMeta: Record<LogClassification, ClassificationMeta> = {
  system: {
    label: '系统日志',
    tone: 'bg-sky-100 text-sky-700',
    dot: 'bg-sky-300',
  },
  business: {
    label: '业务日志',
    tone: 'bg-emerald-100 text-emerald-700',
    dot: 'bg-emerald-300',
  },
  security: {
    label: '安全日志',
    tone: 'bg-rose-100 text-rose-700',
    dot: 'bg-rose-300',
  },
};

const levelMeta: Record<LogLevel, LevelMeta> = {
  debug: { label: 'DEBUG' },
  info: { label: 'INFO' },
  warn: { label: 'WARN' },
  error: { label: 'ERROR' },
};

const unknownClassificationMeta: ClassificationMeta = {
  label: '未知分类',
  tone: 'bg-slate-100 text-slate-700',
  dot: 'bg-slate-300',
};

const unknownLevelMeta: LevelMeta = {
  label: 'UNKNOWN',
};

function isLogClassification(value: string): value is LogClassification {
  return value in classificationMeta;
}

function isLogLevel(value: string): value is LogLevel {
  return value in levelMeta;
}

export function getClassificationMeta(classification: string) {
  if (isLogClassification(classification)) {
    return classificationMeta[classification];
  }

  return {
    ...unknownClassificationMeta,
    label: classification ? `${unknownClassificationMeta.label}(${classification})` : unknownClassificationMeta.label,
  };
}

export function getLevelMeta(level: string) {
  if (isLogLevel(level)) {
    return levelMeta[level];
  }

  return {
    label: level ? `${unknownLevelMeta.label}(${level.toUpperCase()})` : unknownLevelMeta.label,
  };
}

export function formatLogLine(item: AppLogItem) {
  return `${formatDateTimeWithMilliseconds(item.created_at)} | ${getClassificationMeta(item.classification).label} | ${getLevelMeta(item.level).label} | ${item.message}`;
}

export function matchesLogKeyword(item: AppLogItem, keyword: string) {
  const normalizedKeyword = keyword.trim().toLowerCase();
  if (!normalizedKeyword) {
    return true;
  }

  const classificationLabel = getClassificationMeta(item.classification).label.toLowerCase();
  const levelLabel = getLevelMeta(item.level).label.toLowerCase();

  return (
    item.message.toLowerCase().includes(normalizedKeyword) ||
    item.classification.toLowerCase().includes(normalizedKeyword) ||
    classificationLabel.includes(normalizedKeyword) ||
    item.level.toLowerCase().includes(normalizedKeyword) ||
    levelLabel.includes(normalizedKeyword) ||
    formatDateTime(item.created_at).toLowerCase().includes(normalizedKeyword) ||
    formatDateTimeWithMilliseconds(item.created_at).toLowerCase().includes(normalizedKeyword)
  );
}
