import type {NavigationItem} from '@/types/navigation';

export const dashboardNavigation: NavigationItem[] = [
    {
        href: '/',
        label: '概览',
        icon: 'home',
    },
    {
        href: '/import',
        label: '配置导入',
        icon: 'import',
    },
    {
        href: '/nodes',
        label: '节点池',
        icon: 'node',
    },
    {
        href: '/file',
        label: '文件',
        icon: 'file',
    },
    {
        href: '/log',
        label: '日志',
        icon: 'log',
    },
    {
        href: '/setting',
        label: '设置',
        icon: 'setting',
    },
    {
        href: '/user',
        label: '用户',
        icon: 'user',
    },
];
