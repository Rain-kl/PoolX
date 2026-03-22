import type {NavigationItem} from '@/types/navigation';

export const dashboardNavigation: NavigationItem[] = [
    {
        href: '/',
        label: '内核',
        icon: 'runtime',
    },
    {
        href: '/workspace',
        label: '编排',
        icon: 'workspace',
    },
    {
        href: '/nodes',
        label: '节点',
        icon: 'node',
    },
    {
        href: '/log',
        label: '日志',
        icon: 'log',
    },
    {
        href: '/user',
        label: '用户',
        icon: 'user',
    },
    {
        href: '/setting',
        label: '设置',
        icon: 'setting',
    },
];
