import Link from 'next/link';

import { AppCard } from '@/components/ui/app-card';

export default function NotFound() {
  return (
    <main className='flex min-h-screen items-center justify-center px-4'>
      <AppCard title='页面不存在' description='当前请求的页面尚未接入新版管理端路由。'>
        <div className='space-y-4 text-sm leading-6 text-[var(--foreground-secondary)]'>
          <p>请返回总览页，或继续通过侧边导航访问已初始化的模块入口。</p>
          <Link href='/' className='inline-flex rounded-full bg-[var(--brand-primary)] px-4 py-2 font-medium text-[var(--foreground-inverse)] transition hover:opacity-90'>
            返回总览
          </Link>
        </div>
      </AppCard>
    </main>
  );
}
