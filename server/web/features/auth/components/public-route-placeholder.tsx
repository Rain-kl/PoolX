import { FeaturePlaceholder } from '@/components/feedback/feature-placeholder';

interface PublicRoutePlaceholderProps {
  title: string;
  description: string;
}

export function PublicRoutePlaceholder({
  title,
  description,
}: PublicRoutePlaceholderProps) {
  return (
    <FeaturePlaceholder
      title={title}
      description={description}
      milestones={[
        '认证页路由入口已建立，便于阶段 2 接入真实表单。',
        '公共布局已统一，可复用品牌区、说明区与回跳入口。',
        '后续将在此处补齐表单校验、Session 兼容与 OAuth 回调处理。',
      ]}
    />
  );
}
