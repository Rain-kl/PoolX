import Link from 'next/link';

import { AppCard } from '@/components/ui/app-card';
import { PageHeader } from '@/components/layout/page-header';

const quickLinks = [
  {
    href: '/user',
    title: '用户管理',
    description: '查看、搜索和维护模板工程中的账号。',
  },
  {
    href: '/file',
    title: '文件管理',
    description: '上传附件、查看下载次数并执行删除。',
  },
  {
    href: '/setting',
    title: '系统设置',
    description: '维护登录开关、邮箱配置、安全策略和版本升级。',
  },
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="模板总览"
        description="当前管理端已经收敛为模板工程口径，只保留用户、文件、设置与升级等通用能力。"
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
