import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  parseSourceConfigMock,
  parseSourceConfigByURLMock,
  importSourceConfigMock,
  testParsedNodesMock,
} = vi.hoisted(() => ({
  parseSourceConfigMock: vi.fn(),
  parseSourceConfigByURLMock: vi.fn(),
  importSourceConfigMock: vi.fn(),
  testParsedNodesMock: vi.fn(),
}));

vi.mock('@/features/import/api/source-import', () => ({
  parseSourceConfig: parseSourceConfigMock,
  parseSourceConfigByURL: parseSourceConfigByURLMock,
  importSourceConfig: importSourceConfigMock,
  testParsedNodes: testParsedNodesMock,
}));

import { SourceImportPanel } from '@/features/import/components/source-import-page';

function renderWithQueryClient(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
  );
}

const sourceParseResult = {
  source_config: {
    id: 1,
    filename: 'subscription.yaml',
    content_hash: 'hash-1',
    status: 'parsed',
    total_nodes: 1,
    valid_nodes: 1,
    invalid_nodes: 0,
    duplicate_nodes: 0,
    imported_nodes: 0,
    uploaded_by: 'tester',
    uploaded_by_id: 1,
    created_at: '2026-04-06T12:00:00Z',
    updated_at: '2026-04-06T12:00:00Z',
    source_type: 'subscription_url' as const,
    source_url: 'https://sub.example.com/clash.yaml?token=secret-token',
    content_type: 'text/yaml',
    fetched_at: '2026-04-06T12:00:00Z',
  },
  summary: {
    total_nodes: 1,
    valid_nodes: 1,
    invalid_nodes: 0,
    duplicate_nodes: 0,
    importable_nodes: 1,
  },
  nodes: [
    {
      name: 'Node A',
      type: 'ss',
      server: '1.1.1.1',
      port: 443,
      fingerprint: 'node-a',
      duplicate_scope: 'none' as const,
    },
  ],
  errors: [],
};

describe('SourceImportPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows validation feedback when subscription url is blank', async () => {
    const user = userEvent.setup();

    renderWithQueryClient(<SourceImportPanel />);

    await user.click(screen.getByRole('button', { name: '订阅地址' }));
    await user.click(screen.getByRole('button', { name: '开始解析' }));

    expect(
      await screen.findAllByText('请输入有效的 http/https 订阅地址。'),
    ).not.toHaveLength(0);
  });

  it('submits subscription urls through the dedicated API and renders the result summary', async () => {
    const user = userEvent.setup();
    parseSourceConfigByURLMock.mockResolvedValue(sourceParseResult);

    renderWithQueryClient(<SourceImportPanel />);

    await user.click(screen.getByRole('button', { name: '订阅地址' }));
    await user.type(
      screen.getByLabelText('订阅地址'),
      'https://sub.example.com/clash.yaml',
    );
    await user.click(screen.getByRole('button', { name: '开始解析' }));

    await waitFor(() => {
      expect(parseSourceConfigByURLMock).toHaveBeenCalledWith(
        'https://sub.example.com/clash.yaml',
      );
    });

    expect(
      await screen.findByText('解析完成，识别到 1 个有效节点。'),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/来源订阅 https:\/\/sub\.example\.com\/clash\.yaml\?\*\*\*/),
    ).toBeInTheDocument();
    expect(screen.queryByText(/secret-token/)).not.toBeInTheDocument();
  });
});
