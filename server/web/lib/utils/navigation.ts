import { dashboardNavigation } from '@/lib/constants/navigation';
import type { NavigationItem } from '@/types/navigation';

const navigationPathAliases: Record<string, string> = {};

function normalizeNavigationPath(pathname: string) {
  return navigationPathAliases[pathname] ?? pathname;
}

export function isPathActive(pathname: string, href: string) {
  const normalizedPathname = normalizeNavigationPath(pathname);

  if (href === '/') {
    return normalizedPathname === '/';
  }

  return (
    normalizedPathname === href || normalizedPathname.startsWith(`${href}/`)
  );
}

export function isNavigationItemActive(
  pathname: string,
  item: NavigationItem,
): boolean {
  return (
    isPathActive(pathname, item.href) ||
    item.children?.some((child) => isNavigationItemActive(pathname, child)) ||
    false
  );
}

export function flattenNavigationItems(
  items: NavigationItem[],
): NavigationItem[] {
  return items.flatMap((item) => [
    item,
    ...(item.children ? flattenNavigationItems(item.children) : []),
  ]);
}

export function getCurrentNavigationItem(
  pathname: string,
): NavigationItem | undefined {
  const findMatch = (items: NavigationItem[]): NavigationItem | undefined => {
    for (const item of items) {
      if (item.children) {
        const childMatch = findMatch(item.children);
        if (childMatch) {
          return childMatch;
        }
      }

      if (isPathActive(pathname, item.href)) {
        return item;
      }
    }

    return undefined;
  };

  return findMatch(dashboardNavigation);
}
