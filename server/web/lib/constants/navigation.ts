import type {NavigationItem} from '@/types/navigation';

export const dashboardNavigation: NavigationItem[] = [
    {
        href: '/',
        label: '模板',
        icon: 'home',
    },
    {
        href: '/user',
        label: '用户',
        icon: 'user',
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
];
