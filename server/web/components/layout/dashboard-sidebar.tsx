'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

import { dashboardNavigation } from '@/lib/constants/navigation';
import { cn } from '@/lib/utils/cn';
import { isNavigationItemActive } from '@/lib/utils/navigation';
import { useAppShellStore } from '@/store/app-shell';
import type { NavigationIconKey, NavigationItem } from '@/types/navigation';

function SidebarIcon({ icon }: { icon: NavigationIconKey }) {
  const commonProps = {
    className: 'h-[18px] w-[18px]',
    fill: 'none',
    stroke: 'currentColor',
    strokeWidth: 1.8,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
    viewBox: '0 0 24 24',
  };

  switch (icon) {
    case 'home':
      return (
        <svg {...commonProps}>
          <path d="M3 10.5 12 3l9 7.5" />
          <path d="M5.5 9.5V21h13V9.5" />
          <path d="M9.5 21v-6h5v6" />
        </svg>
      );
    case 'file':
      return (
        <svg {...commonProps}>
          <path d="M7 4.5h7l4 4V19a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6.5a2 2 0 0 1 2-2Z" />
          <path d="M14 4.5V9h4" />
          <path d="M8.5 13H15.5" />
          <path d="M8.5 16.5H13" />
        </svg>
      );
    case 'log':
      return (
        <svg {...commonProps}>
          <path d="M6 6.5h12" />
          <path d="M6 12h12" />
          <path d="M6 17.5h8" />
          <path d="M4.5 4.5h15v15h-15z" />
        </svg>
      );
    case 'user':
      return (
        <svg {...commonProps}>
          <circle cx="12" cy="8" r="3.25" />
          <path d="M5 19.5c1.7-3 4.1-4.5 7-4.5s5.3 1.5 7 4.5" />
        </svg>
      );
    case 'setting':
      return (
        <svg {...commonProps}>
          <circle cx="12" cy="12" r="3.5" />
          <path d="M12 3.5v2.2M12 18.3v2.2M20.5 12h-2.2M5.7 12H3.5M18.1 5.9l-1.6 1.6M7.5 16.5l-1.6 1.6M18.1 18.1l-1.6-1.6M7.5 7.5 5.9 5.9" />
        </svg>
      );
  }
}

function SidebarNavItem({
  item,
  currentPath,
  isSidebarCollapsed,
  forceExpanded,
  onNavigate,
  depth = 0,
}: {
  item: NavigationItem;
  currentPath: string;
  isSidebarCollapsed: boolean;
  forceExpanded?: boolean;
  onNavigate?: () => void;
  depth?: number;
}) {
  const active = isNavigationItemActive(currentPath, item);
  const hasChildren = Boolean(item.children?.length);
  const showLabel = forceExpanded || !isSidebarCollapsed;

  return (
    <div className="space-y-2">
      <Link
        href={item.href}
        onClick={onNavigate}
        className={cn(
          'flex min-h-[50px] items-center gap-3 rounded-2xl border px-3 py-2.5 transition-colors',
          depth > 0 && 'ml-3 rounded-xl',
          active
            ? 'border-[var(--border-strong)] bg-[var(--accent-soft)] text-[var(--foreground-primary)]'
            : 'border-transparent text-[var(--foreground-secondary)] hover:border-[var(--border-default)] hover:bg-[var(--surface-muted)] hover:text-[var(--foreground-primary)]',
        )}
      >
        <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-[var(--control-background)] text-[var(--foreground-primary)]">
          <SidebarIcon icon={item.icon} />
        </span>
        {showLabel ? (
          <span className="min-w-0 flex-1 text-sm font-medium">
            {item.label}
          </span>
        ) : null}
      </Link>
      {showLabel && hasChildren ? (
        <div className="space-y-2">
          {item.children?.map((child) => (
            <SidebarNavItem
              key={child.href}
              item={child}
              currentPath={currentPath}
              isSidebarCollapsed={isSidebarCollapsed}
              forceExpanded={forceExpanded}
              onNavigate={onNavigate}
              depth={depth + 1}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
}

function SidebarContent({
  currentPath,
  isSidebarCollapsed,
  forceExpanded = false,
  onNavigate,
}: {
  currentPath: string;
  isSidebarCollapsed: boolean;
  forceExpanded?: boolean;
  onNavigate?: () => void;
}) {
  const showLabel = forceExpanded || !isSidebarCollapsed;

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex items-center gap-3 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-muted)] px-3 py-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-[var(--brand-primary-soft)] text-sm font-semibold text-[var(--brand-primary)]">
          AF
        </div>
        {showLabel ? (
          <div>
            <p className="text-sm font-semibold text-[var(--foreground-primary)]">
            GinNext
            </p>
            <p className="text-xs text-[var(--foreground-secondary)]">Template</p>
          </div>
        ) : null}
      </div>

      <nav className="flex-1 space-y-2">
        <div className="flex max-h-full min-h-0 flex-col gap-2 overflow-y-auto pr-1">
          {dashboardNavigation.map((item) => (
            <SidebarNavItem
              key={item.href}
              item={item}
              currentPath={currentPath}
              isSidebarCollapsed={isSidebarCollapsed}
              forceExpanded={forceExpanded}
              onNavigate={onNavigate}
            />
          ))}
        </div>
      </nav>
    </div>
  );
}

export function DashboardSidebar() {
  const pathname = usePathname();
  const currentPath = pathname ?? '/';
  const isSidebarCollapsed = useAppShellStore(
    (state) => state.isSidebarCollapsed,
  );
  const isMobileSidebarOpen = useAppShellStore(
    (state) => state.isMobileSidebarOpen,
  );
  const setMobileSidebarOpen = useAppShellStore(
    (state) => state.setMobileSidebarOpen,
  );

  useEffect(() => {
    setMobileSidebarOpen(false);
  }, [currentPath, setMobileSidebarOpen]);

  return (
    <>
      <div
        className={cn(
          'fixed inset-0 z-30 bg-black/35 transition-opacity duration-200 min-[1000px]:hidden',
          isMobileSidebarOpen ? 'opacity-100' : 'pointer-events-none opacity-0',
        )}
        onClick={() => setMobileSidebarOpen(false)}
        aria-hidden="true"
      />

      <aside
        className={cn(
          'fixed top-0 left-0 z-40 h-screen w-[200px] overflow-hidden border-r border-[var(--border-default)] bg-[var(--surface-panel)]/95 px-3 py-5 backdrop-blur transition-transform duration-200 min-[1000px]:hidden',
          isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <SidebarContent
          currentPath={currentPath}
          isSidebarCollapsed={false}
          forceExpanded
          onNavigate={() => setMobileSidebarOpen(false)}
        />
      </aside>

      <aside
        className={cn(
          'sticky top-0 z-10 hidden h-screen shrink-0 overflow-hidden border-r border-[var(--border-default)] bg-[var(--surface-panel)]/95 px-3 py-5 backdrop-blur min-[1000px]:block',
          isSidebarCollapsed ? 'w-[76px]' : 'w-[200px]',
        )}
      >
        <SidebarContent
          currentPath={currentPath}
          isSidebarCollapsed={isSidebarCollapsed}
        />
      </aside>
    </>
  );
}
