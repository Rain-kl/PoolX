import { describe, expect, it } from 'vitest';

import {
  calculateNiceAxisMax,
  formatBytes,
  formatBytesPerSecond,
  formatPercent,
} from '@/lib/utils/metrics';

describe('metrics utils', () => {
  it('formats byte values with readable units', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(1536)).toBe('1.5 KB');
    expect(formatBytes(183 * 1024 * 1024)).toBe('183 MB');
    expect(formatBytes(14.7 * 1024 * 1024)).toBe('14.7 MB');
  });

  it('formats byte rates with per-second units', () => {
    expect(formatBytesPerSecond(0, 3600)).toBe('0 B/s');
    expect(formatBytesPerSecond(1630858.9, 3600)).toBe('453 B/s');
    expect(formatBytesPerSecond(22620490956.8, 3600)).toBe('6 MB/s');
  });

  it('formats percent values without long decimals', () => {
    expect(formatPercent(58.36809884639982)).toBe('58.4%');
  });

  it('rounds chart max values to friendly axis steps', () => {
    expect(calculateNiceAxisMax([0, 58.36809884639982])).toBe(80);
    expect(calculateNiceAxisMax([0, 1630858.9])).toBe(2000000);
    expect(calculateNiceAxisMax([0, 22620490956.8])).toBe(25000000000);
  });
});
