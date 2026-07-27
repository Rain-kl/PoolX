import type { PeriodValue } from '@/shared/lib/period';

export type DashboardPeriod = PeriodValue;

export type DashboardUsageDTO = {
  requests: number;
  successfulRequests: number;
  failedRequests: number;
  inputTokens: number;
  cachedInputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  tokens: number;
  billedCostUsdTicks: number;
  successRate: number;
};

export type DashboardDTO = {
  period: DashboardPeriod;
  generatedAt: string;
  range: { start: string; end: string };
  resources: {
    activeResources: number;
    totalResources: number;
    primaryResources: number;
    edgeResources: number;
    backupResources: number;
    enabledServices: number;
    totalServices: number;
    activeClients: number;
    totalClients: number;
  };
  usage: DashboardUsageDTO;
  series: Array<{
    start: string;
    end: string;
    requests: number;
    inputTokens: number;
    cachedInputTokens: number;
    outputTokens: number;
    reasoningTokens: number;
    tokens: number;
    billedCostUsdTicks: number;
  }>;
  activity: Array<{ start: string; requests: number }>;
  topServices: Array<{
    service: string;
    requests: number;
    inputTokens: number;
    cachedInputTokens: number;
    outputTokens: number;
    reasoningTokens: number;
    tokens: number;
    billedCostUsdTicks: number;
  }>;
  providers: Array<{
    provider: string;
    requests: number;
    successfulRequests: number;
    tokens: number;
  }>;
};
