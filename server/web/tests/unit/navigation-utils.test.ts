import { describe, expect, it } from 'vitest';

import { getCurrentNavigationItem, isPathActive } from '@/lib/utils/navigation';

describe('navigation utils', () => {
  it('marks root path as active only for home', () => {
    expect(isPathActive('/', '/')).toBe(true);
    expect(isPathActive('/user', '/')).toBe(false);
  });

  it('resolves current navigation item for nested paths', () => {
    expect(getCurrentNavigationItem('/user/abc')?.label).toBe('用户');
    expect(getCurrentNavigationItem('/file')?.label).toBe('文件');
    expect(getCurrentNavigationItem('/setting')?.label).toBe('设置');
  });
});
