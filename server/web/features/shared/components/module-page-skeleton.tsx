import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { FeaturePlaceholder } from '@/components/feedback/feature-placeholder';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';

interface ModulePageSkeletonProps {
  title: string;
  description: string;
}

export function ModulePageSkeleton({
  title,
  description,
}: ModulePageSkeletonProps) {
  return (
    <div className='space-y-6'>
      <PageHeader title={title} description={description} />

      <FeaturePlaceholder
        title={`${title} 页面骨架`}
        description='当前阶段先完成页面入口、统一头部和反馈组件接线，真实列表、表单与联动能力将在后续业务迁移阶段接入。'
        milestones={[
          '路由入口已创建，可纳入统一导航。',
          '页面标题区、说明区与内容容器已统一。',
          '加载态、空态、错误态组件已建立，可在真实查询接入后直接复用。',
        ]}
      />

      <div className='grid gap-4 xl:grid-cols-3'>
        <LoadingState />
        <EmptyState title='空态骨架' description='后续接入查询后，可将资源为空时的提示统一沉淀为此组件。' />
        <ErrorState title='错误态骨架' description='后续请求失败、鉴权失效与重试提示可在此基础上扩展。' />
      </div>
    </div>
  );
}
