'use client';

import { useMemo } from 'react';
import type { EChartsOption } from 'echarts';
import ReactECharts from 'echarts-for-react';

type RankChartItem = {
  label: string;
  value: number;
};

type RankChartProps = {
  items: RankChartItem[];
  color: string;
  valueFormatter?: (value: number) => string;
  emptyMessage?: string;
};

const defaultFormatter = (value: number) => value.toLocaleString('zh-CN');

function getChartValue(params: unknown) {
  if (typeof params !== 'object' || params === null || !('value' in params)) {
    return 0;
  }
  const rawValue = (params as { value?: unknown }).value;
  if (Array.isArray(rawValue)) {
    const candidate = rawValue[0];
    return typeof candidate === 'number' ? candidate : 0;
  }
  return typeof rawValue === 'number' ? rawValue : 0;
}

export function RankChart({
                            items,
                            color,
                            valueFormatter = defaultFormatter,
                            emptyMessage = '暂无分布数据',
                          }: RankChartProps) {
  const option = useMemo<EChartsOption>(
      () => ({
        animationDuration: 400,
        grid: {
          left: 16,
          right: 24,
          top: 12,
          bottom: 12,
          containLabel: true,
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'shadow',
          },
          backgroundColor: 'rgba(15, 23, 42, 0.92)',
          borderWidth: 0,
          textStyle: {
            color: '#e2e8f0',
            fontSize: 12,
          },
          formatter: (params: unknown) => {
            const item = Array.isArray(params) ? params[0] : params;
            const data = item as { name?: string; value?: number };
            return `${data.name ?? ''}<br/>${valueFormatter(data.value ?? 0)}`;
          },
        },
        xAxis: {
          type: 'value',
          axisLabel: {
            color: '#94a3b8',
          },
          splitLine: {
            lineStyle: {
              color: 'rgba(148, 163, 184, 0.16)',
              type: 'dashed',
            },
          },
        },
        yAxis: {
          type: 'category',
          data: items.map((item) => item.label),
          axisTick: { show: false },
          axisLine: { show: false },
          axisLabel: {
            color: '#cbd5e1',
            width: 120,
            overflow: 'truncate',
          },
        },
        series: [
          {
            type: 'bar',
            data: items.map((item) => item.value),
            barWidth: 12,
            showBackground: true,
            backgroundStyle: {
              color: 'rgba(148, 163, 184, 0.12)',
              borderRadius: 999,
            },
            itemStyle: {
              color,
              borderRadius: 999,
            },
            label: {
              show: true,
              position: 'right',
              color: '#e2e8f0',
              formatter: (params: unknown) => valueFormatter(getChartValue(params)),
            },
          },
        ],
      }),
      [color, items, valueFormatter],
  );

  if (items.length === 0) {
    return (
        <div className="flex h-[220px] items-center justify-center rounded-3xl border border-dashed border-[var(--border-default)] bg-[var(--surface-muted)] text-sm text-[var(--foreground-secondary)]">
          {emptyMessage}
        </div>
    );
  }

  return (
      <ReactECharts
          option={option}
          notMerge
          lazyUpdate
          style={{ height: Math.max(220, items.length * 44), width: '100%' }}
      />
  );
}
