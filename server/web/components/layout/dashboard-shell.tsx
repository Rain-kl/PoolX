import type { ReactNode } from 'react';

import { DashboardSidebar } from '@/components/layout/dashboard-sidebar';
import { DashboardTopbar } from '@/components/layout/dashboard-topbar';

interface DashboardShellProps {
  children: ReactNode;
}

export function DashboardShell({ children }: DashboardShellProps) {
  return (
    <div className='flex min-h-screen bg-transparent text-[var(--foreground-primary)]'>
      <DashboardSidebar />
      <div className='flex min-w-0 flex-1 flex-col'>
        <DashboardTopbar />
        <main className='flex-1 px-4 py-6 md:px-8 md:py-8'>{children}</main>
      </div>
    </div>
  );
}
