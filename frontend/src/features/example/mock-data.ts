import type { DashboardDTO } from '@/features/example/dashboard-types';

export type ExampleRow = {
  id: string;
  name: string;
  status: 'active' | 'paused' | 'archived';
  owner: string;
  category: string;
  requests: number;
  updatedAt: string;
};

export type ExampleChatMessage = {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
};

export type ExampleGalleryItem = {
  id: string;
  title: string;
  accent: string;
  meta: string;
};

function daysAgo(days: number): string {
  const date = new Date();
  date.setHours(12, 0, 0, 0);
  date.setDate(date.getDate() - days);
  return date.toISOString();
}

function hoursAgo(hours: number): string {
  return new Date(Date.now() - hours * 3_600_000).toISOString();
}

export const MOCK_TABLE_ROWS: ExampleRow[] = [
  {
    id: 'ex-1001',
    name: 'Northwind gateway',
    status: 'active',
    owner: 'Ava Chen',
    category: 'API',
    requests: 12840,
    updatedAt: hoursAgo(2),
  },
  {
    id: 'ex-1002',
    name: 'Atlas billing',
    status: 'active',
    owner: 'Noah Park',
    category: 'Billing',
    requests: 9420,
    updatedAt: hoursAgo(5),
  },
  {
    id: 'ex-1003',
    name: 'Harbor chat',
    status: 'paused',
    owner: 'Mia Lopez',
    category: 'Chat',
    requests: 6102,
    updatedAt: hoursAgo(9),
  },
  {
    id: 'ex-1004',
    name: 'Lumen analytics',
    status: 'active',
    owner: 'Eli Brooks',
    category: 'Analytics',
    requests: 15880,
    updatedAt: hoursAgo(12),
  },
  {
    id: 'ex-1005',
    name: 'Cedar docs',
    status: 'archived',
    owner: 'Sophia Kim',
    category: 'Docs',
    requests: 2201,
    updatedAt: daysAgo(2),
  },
  {
    id: 'ex-1006',
    name: 'Orbit media',
    status: 'active',
    owner: 'Liam Ortiz',
    category: 'Media',
    requests: 7344,
    updatedAt: daysAgo(1),
  },
  {
    id: 'ex-1007',
    name: 'Sable auth',
    status: 'paused',
    owner: 'Emma Wright',
    category: 'Auth',
    requests: 4011,
    updatedAt: daysAgo(3),
  },
  {
    id: 'ex-1008',
    name: 'Pulse monitors',
    status: 'active',
    owner: 'Owen Blake',
    category: 'Ops',
    requests: 11203,
    updatedAt: hoursAgo(18),
  },
  {
    id: 'ex-1009',
    name: 'Nova templates',
    status: 'active',
    owner: 'Zoe Reed',
    category: 'UI',
    requests: 2890,
    updatedAt: daysAgo(4),
  },
  {
    id: 'ex-1010',
    name: 'Quill notes',
    status: 'archived',
    owner: 'Ryan Vale',
    category: 'Content',
    requests: 990,
    updatedAt: daysAgo(6),
  },
  {
    id: 'ex-1011',
    name: 'Forge workers',
    status: 'active',
    owner: 'Ivy Grant',
    category: 'Jobs',
    requests: 8340,
    updatedAt: hoursAgo(7),
  },
  {
    id: 'ex-1012',
    name: 'Beacon alerts',
    status: 'active',
    owner: 'Kai Moss',
    category: 'Ops',
    requests: 5602,
    updatedAt: hoursAgo(3),
  },
  {
    id: 'ex-1013',
    name: 'Mesa storage',
    status: 'paused',
    owner: 'Nora Hill',
    category: 'Storage',
    requests: 3104,
    updatedAt: daysAgo(5),
  },
  {
    id: 'ex-1014',
    name: 'River search',
    status: 'active',
    owner: 'Jude Lane',
    category: 'Search',
    requests: 14220,
    updatedAt: hoursAgo(1),
  },
  {
    id: 'ex-1015',
    name: 'Willow forms',
    status: 'active',
    owner: 'Ada Frost',
    category: 'Forms',
    requests: 1880,
    updatedAt: daysAgo(1),
  },
  {
    id: 'ex-1016',
    name: 'Comet charts',
    status: 'active',
    owner: 'Ben Cole',
    category: 'Charts',
    requests: 6720,
    updatedAt: hoursAgo(14),
  },
  {
    id: 'ex-1017',
    name: 'Drift queues',
    status: 'archived',
    owner: 'Cara West',
    category: 'Jobs',
    requests: 740,
    updatedAt: daysAgo(8),
  },
  {
    id: 'ex-1018',
    name: 'Summit keys',
    status: 'active',
    owner: 'Drew Moss',
    category: 'Auth',
    requests: 4520,
    updatedAt: hoursAgo(6),
  },
  {
    id: 'ex-1019',
    name: 'Pine gallery',
    status: 'active',
    owner: 'Elle Hart',
    category: 'Media',
    requests: 3911,
    updatedAt: hoursAgo(20),
  },
  {
    id: 'ex-1020',
    name: 'Canvas shell',
    status: 'paused',
    owner: 'Finn Vale',
    category: 'UI',
    requests: 2666,
    updatedAt: daysAgo(2),
  },
  {
    id: 'ex-1021',
    name: 'Echo webhooks',
    status: 'active',
    owner: 'Gia Moon',
    category: 'API',
    requests: 9888,
    updatedAt: hoursAgo(4),
  },
  {
    id: 'ex-1022',
    name: 'Lattice grid',
    status: 'active',
    owner: 'Hugo Dane',
    category: 'Tables',
    requests: 5110,
    updatedAt: hoursAgo(11),
  },
  {
    id: 'ex-1023',
    name: 'Marble theme',
    status: 'archived',
    owner: 'Iris Snow',
    category: 'UI',
    requests: 420,
    updatedAt: daysAgo(10),
  },
  {
    id: 'ex-1024',
    name: 'Nest profiles',
    status: 'active',
    owner: 'Jon Apex',
    category: 'Users',
    requests: 7204,
    updatedAt: hoursAgo(8),
  },
];

export const MOCK_GALLERY_ITEMS: ExampleGalleryItem[] = [
  {
    id: 'g1',
    title: 'Aurora panel',
    accent: 'from-sky-500/30 to-cyan-500/10',
    meta: 'Card · Soft gradient',
  },
  {
    id: 'g2',
    title: 'Ember tile',
    accent: 'from-orange-500/30 to-amber-500/10',
    meta: 'Card · Warm accent',
  },
  {
    id: 'g3',
    title: 'Moss block',
    accent: 'from-emerald-500/30 to-teal-500/10',
    meta: 'Card · Calm surface',
  },
  {
    id: 'g4',
    title: 'Violet frame',
    accent: 'from-violet-500/30 to-fuchsia-500/10',
    meta: 'Card · Contrast',
  },
  {
    id: 'g5',
    title: 'Slate board',
    accent: 'from-slate-500/30 to-zinc-500/10',
    meta: 'Card · Neutral',
  },
  {
    id: 'g6',
    title: 'Coral shelf',
    accent: 'from-rose-500/30 to-pink-500/10',
    meta: 'Card · Highlight',
  },
];

export const MOCK_CHAT_SEED: ExampleChatMessage[] = [
  {
    id: 'm1',
    role: 'assistant',
    content:
      'Welcome to the PoolX chat page template. Messages stay in local state so you can explore the layout without backend APIs.',
    createdAt: hoursAgo(1),
  },
  {
    id: 'm2',
    role: 'user',
    content: 'Show me how a short product update would look in this thread.',
    createdAt: hoursAgo(1),
  },
  {
    id: 'm3',
    role: 'assistant',
    content:
      'Ship checklist: component gallery, dashboard template, table template, and this chat shell. All example routes use mock or local data.',
    createdAt: hoursAgo(1),
  },
];

function buildSeries(days: number): DashboardDTO['series'] {
  return Array.from({ length: days }, (_, index) => {
    const start = new Date();
    start.setHours(0, 0, 0, 0);
    start.setDate(start.getDate() - (days - 1 - index));
    const end = new Date(start);
    end.setDate(end.getDate() + 1);
    const requests = 180 + ((index * 37) % 90) + (index % 5) * 12;
    const tokens = requests * (42 + (index % 7));
    return {
      start: start.toISOString(),
      end: end.toISOString(),
      requests,
      inputTokens: Math.round(tokens * 0.62),
      cachedInputTokens: Math.round(tokens * 0.18),
      outputTokens: Math.round(tokens * 0.16),
      reasoningTokens: Math.round(tokens * 0.04),
      tokens,
      billedCostUsdTicks: requests * 12_500_000,
    };
  });
}

function buildActivity(): DashboardDTO['activity'] {
  return Array.from({ length: 180 }, (_, index) => {
    const start = new Date();
    start.setHours(12, 0, 0, 0);
    start.setDate(start.getDate() - (179 - index));
    const requests = Math.max(
      0,
      Math.round(20 + 40 * Math.sin(index / 9) + ((index * 13) % 30)),
    );
    return { start: start.toISOString(), requests };
  });
}

export function getMockDashboard(
  period: '24h' | '7d' | '30d' | '90d' = '30d',
): DashboardDTO {
  const dayCount =
    period === '24h' ? 24 : period === '7d' ? 7 : period === '90d' ? 90 : 30;
  const series =
    period === '24h'
      ? Array.from({ length: 24 }, (_, index) => {
          const start = new Date();
          start.setMinutes(0, 0, 0);
          start.setHours(start.getHours() - (23 - index));
          const end = new Date(start);
          end.setHours(end.getHours() + 1);
          const requests = 8 + ((index * 5) % 17);
          const tokens = requests * 50;
          return {
            start: start.toISOString(),
            end: end.toISOString(),
            requests,
            inputTokens: Math.round(tokens * 0.6),
            cachedInputTokens: Math.round(tokens * 0.2),
            outputTokens: Math.round(tokens * 0.15),
            reasoningTokens: Math.round(tokens * 0.05),
            tokens,
            billedCostUsdTicks: requests * 11_000_000,
          };
        })
      : buildSeries(dayCount);

  const usageTotals = series.reduce(
    (acc, bucket) => ({
      requests: acc.requests + bucket.requests,
      inputTokens: acc.inputTokens + bucket.inputTokens,
      cachedInputTokens: acc.cachedInputTokens + bucket.cachedInputTokens,
      outputTokens: acc.outputTokens + bucket.outputTokens,
      reasoningTokens: acc.reasoningTokens + bucket.reasoningTokens,
      tokens: acc.tokens + bucket.tokens,
      billedCostUsdTicks: acc.billedCostUsdTicks + bucket.billedCostUsdTicks,
    }),
    {
      requests: 0,
      inputTokens: 0,
      cachedInputTokens: 0,
      outputTokens: 0,
      reasoningTokens: 0,
      tokens: 0,
      billedCostUsdTicks: 0,
    },
  );
  const successfulRequests = Math.round(usageTotals.requests * 0.974);
  const failedRequests = usageTotals.requests - successfulRequests;
  const now = new Date().toISOString();

  return {
    period,
    generatedAt: now,
    range: {
      start: series[0]?.start ?? now,
      end: series[series.length - 1]?.end ?? now,
    },
    resources: {
      activeResources: 42,
      totalResources: 48,
      primaryResources: 18,
      edgeResources: 20,
      backupResources: 10,
      enabledServices: 14,
      totalServices: 16,
      activeClients: 9,
      totalClients: 11,
    },
    usage: {
      ...usageTotals,
      successfulRequests,
      failedRequests,
      successRate:
        usageTotals.requests > 0
          ? (successfulRequests / usageTotals.requests) * 100
          : 0,
    },
    series,
    activity: buildActivity(),
    topServices: [
      {
        service: 'poolx-chat-pro',
        requests: 4200,
        inputTokens: 1_820_000,
        cachedInputTokens: 420_000,
        outputTokens: 510_000,
        reasoningTokens: 88_000,
        tokens: 2_838_000,
        billedCostUsdTicks: 58_000_000_000,
      },
      {
        service: 'poolx-vision',
        requests: 1880,
        inputTokens: 640_000,
        cachedInputTokens: 90_000,
        outputTokens: 120_000,
        reasoningTokens: 0,
        tokens: 850_000,
        billedCostUsdTicks: 31_000_000_000,
      },
      {
        service: 'poolx-fast',
        requests: 3510,
        inputTokens: 990_000,
        cachedInputTokens: 210_000,
        outputTokens: 280_000,
        reasoningTokens: 12_000,
        tokens: 1_492_000,
        billedCostUsdTicks: 22_000_000_000,
      },
      {
        service: 'poolx-reasoner',
        requests: 960,
        inputTokens: 720_000,
        cachedInputTokens: 60_000,
        outputTokens: 190_000,
        reasoningTokens: 240_000,
        tokens: 1_210_000,
        billedCostUsdTicks: 19_500_000_000,
      },
      {
        service: 'poolx-embed',
        requests: 2740,
        inputTokens: 1_100_000,
        cachedInputTokens: 0,
        outputTokens: 0,
        reasoningTokens: 0,
        tokens: 1_100_000,
        billedCostUsdTicks: 4_200_000_000,
      },
    ],
    providers: [
      {
        provider: 'primary',
        requests: 9200,
        successfulRequests: 8970,
        tokens: 4_200_000,
      },
      {
        provider: 'edge',
        requests: 5100,
        successfulRequests: 4990,
        tokens: 2_100_000,
      },
      {
        provider: 'backup',
        requests: 2400,
        successfulRequests: 2310,
        tokens: 980_000,
      },
    ],
  };
}
