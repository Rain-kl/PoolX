'use client';

import { useMemo } from 'react';
import type { EChartsOption } from 'echarts';
import ReactECharts from 'echarts-for-react';

import { calculateNiceAxisMax, formatCompactNumber } from '@/lib/utils/metrics';

type TrendChartSeries = {
  label: string;
  color: string;
  fillColor?: string;
  values: number[];
  variant?: 'line' | 'area';
  valueFormatter?: (value: number) => string;
};

type TrendChartProps = {
  labels: string[];
  series: TrendChartSeries[];
  height?: number;
  yAxisValueFormatter?: (value: number) => string;
};

type TooltipParam = {
  axisValueLabel?: string;
  color?: string;
  seriesName?: string;
  value?: number | string | Array<number | string>;
};

const defaultFormatter = (value: number) => formatCompactNumber(value);

export function TrendChart({
                             labels,
                             series,
                             height = 220,
                             yAxisValueFormatter,
                           }: TrendChartProps) {
  const option = useMemo<EChartsOption>(() => {
    const axisFormatter = yAxisValueFormatter ?? defaultFormatter;
    const maxValue = calculateNiceAxisMax(
        series.flatMap((item) => item.values),
    );

    return {
      animationDuration: 500,
      animationEasing: 'cubicOut',
      grid: {
        left: 16,
        right: 16,
        top: 20,
        bottom: 20,
        containLabel: true,
      },
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(15, 23, 42, 0.92)',
        borderWidth: 0,
        textStyle: {
          color: '#e2e8f0',
          fontSize: 12,
        },
        formatter: (params: unknown) => {
          const items = Array.isArray(params) ? (params as TooltipParam[]) : [];
          if (items.length === 0) {
            return '';
          }

          const header = items[0]?.axisValueLabel ?? '';
          const rows = items.map((item) => {
            const matchedSeries = series.find(
                (seriesItem) => seriesItem.label === item.seriesName,
            );
            const formatter =
                matchedSeries?.valueFormatter ??
                yAxisValueFormatter ??
                defaultFormatter;
            const rawValue = Array.isArray(item.value)
                ? item.value[1]
                : item.value;
            const numericValue =
                typeof rawValue === 'number' ? rawValue : Number(rawValue ?? 0);

            return [
              '<span style="display:inline-flex;align-items:center;gap:8px;">',
              `<span style="display:inline-block;width:8px;height:8px;border-radius:9999px;background:${item.color ?? '#94a3b8'};"></span>`,
              `<span>${item.seriesName ?? ''}</span>`,
              `<strong style="margin-left:8px;">${formatter(numericValue)}</strong>`,
              '</span>',
            ].join('');
          });

          return [header, ...rows].join('<br/>');
        },
      },
      legend: {
        show: false,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: labels,
        axisLine: {
          lineStyle: {
            color: 'rgba(148, 163, 184, 0.24)',
          },
        },
        axisTick: {
          show: false,
        },
        axisLabel: {
          color: '#94a3b8',
          margin: 14,
        },
      },
      yAxis: {
        type: 'value',
        min: 0,
        max: maxValue,
        splitNumber: 4,
        axisLabel: {
          color: '#94a3b8',
          formatter: (value: number) => axisFormatter(value),
        },
        splitLine: {
          lineStyle: {
            color: 'rgba(148, 163, 184, 0.16)',
            type: 'dashed',
          },
        },
      },
      series: series.map((item) => ({
        name: item.label,
        type: 'line',
        smooth: true,
        showSymbol: false,
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: {
          color: item.color,
          width: 3,
        },
        itemStyle: {
          color: item.color,
        },
        areaStyle:
            item.variant === 'area'
                ? {
                  color: item.fillColor ?? `${item.color}33`,
                }
                : undefined,
        emphasis: {
          focus: 'series',
          scale: true,
        },
        data: item.values,
      })),
    };
  }, [labels, series, yAxisValueFormatter]);

  if (labels.length === 0 || series.length === 0) {
    return (
        <div className="flex h-[220px] items-center justify-center rounded-3xl border border-dashed border-[var(--border-default)] bg-[var(--surface-muted)] text-sm text-[var(--foreground-secondary)]">
          暂无趋势数据
        </div>
    );
  }

  return (
      <div className="space-y-4">
        <div className="flex flex-wrap gap-3">
          {series.map((item) => {
            const latestValue = item.values[item.values.length - 1] ?? 0;
            const formatter = item.valueFormatter ?? defaultFormatter;
            return (
                <div
                    key={item.label}
                    className="min-w-[140px] rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3"
                >
                  <div className="flex items-center gap-2">
                <span
                    className="h-2.5 w-2.5 rounded-full"
                    style={{ backgroundColor: item.color }}
                />
                    <p className="text-xs tracking-[0.18em] text-[var(--foreground-muted)] uppercase">
                      {item.label}
                    </p>
                  </div>
                  <p className="mt-2 text-lg font-semibold text-[var(--foreground-primary)]">
                    {formatter(latestValue)}
                  </p>
                </div>
            );
          })}
        </div>

        <div className="overflow-hidden rounded-[28px] border border-[var(--border-default)] bg-[linear-gradient(180deg,rgba(255,255,255,0.03),rgba(255,255,255,0))] px-4 py-4">
          <ReactECharts
              option={option}
              notMerge
              lazyUpdate
              style={{ height, width: '100%' }}
          />
        </div>
      </div>
  );
}
