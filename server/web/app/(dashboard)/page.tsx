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
    href: '/workspace',
    title: '工作台',
    description: '维护端口配置、生成可合并片段，并为后续统一启动做准备。',
  },
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="PoolX 总览"
        description="当前阶段进入 Phase 2，开始把节点池衔接到工作台配置、片段预览与后续统一启动。"
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
