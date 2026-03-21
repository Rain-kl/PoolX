import Link from 'next/link';

import { AppCard } from '@/components/ui/app-card';
import { PageHeader } from '@/components/layout/page-header';

const quickLinks = [
  {
    href: '/import',
    title: '配置导入',
    description: '上传 YAML、查看解析摘要，并确认导入节点池。',
  },
  {
    href: '/nodes',
    title: '节点池',
    description: '分页查看节点、筛选状态并执行连通性测试。',
  },
  {
    href: '/setting',
    title: '系统设置',
    description: '维护内核路径、系统参数、安全策略和升级能力。',
  },
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="PoolX 总览"
        description="当前阶段以 Phase 1 为主线，先打通 YAML 导入、节点池管理和节点测试的最小可用闭环。"
      />

      <div className="grid gap-4 lg:grid-cols-3">
        {quickLinks.map((item) => (
          <Link key={item.href} href={item.href} className="block">
            <AppCard title={item.title} description={item.description}>
              <p className="text-sm text-[var(--foreground-secondary)]">
                点击进入 {item.title}
              </p>
            </AppCard>
          </Link>
        ))}
      </div>
    </div>
  );
}
