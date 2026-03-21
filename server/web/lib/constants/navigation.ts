import type {NavigationItem} from '@/types/navigation';

export const dashboardNavigation: NavigationItem[] = [
    {
        href: '/',
        label: '运行状态',
        icon: 'runtime',
    },
    {
        href: '/workspace',
        label: '工作台',
        icon: 'workspace',
    },
    {
        href: '/nodes',
        label: '节点池',
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
