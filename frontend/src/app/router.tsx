import { Navigate, createBrowserRouter } from 'react-router-dom';

import { AnonymousBoundary, AuthBoundary } from '@/app/auth-boundary';
import {
  DeferredAppShell,
  DeferredExampleChatPage,
  DeferredExampleComponentPage,
  DeferredExampleDashboardPage,
  DeferredExampleTablePage,
  DeferredSettingsPage,
} from '@/app/deferred-pages';
import { LoginPage } from '@/features/auth/login-page';
import { ClashDashboardPage } from '@/features/clash/dashboard/clash-dashboard-page';
import { SourceConfigsPage } from '@/features/clash/source-configs/source-configs-page';
import { NodePoolPage } from '@/features/clash/nodes/node-pool-page';
import { PortProfilesPage } from '@/features/clash/port-profiles/port-profiles-page';

const defaultAuthedPath = '/clash/dashboard';

export const router = createBrowserRouter([
  {
    element: <AnonymousBoundary />,
    children: [{ path: '/login', element: <LoginPage /> }],
  },
  {
    element: <AuthBoundary />,
    children: [
      {
        element: <DeferredAppShell />,
        children: [
          { index: true, element: <Navigate to={defaultAuthedPath} replace /> },
          { path: '/clash/dashboard', element: <ClashDashboardPage /> },
          { path: '/clash/source-configs', element: <SourceConfigsPage /> },
          { path: '/clash/nodes', element: <NodePoolPage /> },
          { path: '/clash/port-profiles', element: <PortProfilesPage /> },
          {
            path: '/clash/runtime',
            element: <Navigate to='/clash/dashboard' replace />,
          },
          {
            path: '/clash/settings',
            element: <Navigate to='/settings' replace />,
          },
          {
            path: '/example/component',
            element: <DeferredExampleComponentPage />,
          },
          {
            path: '/example/page/dashboard',
            element: <DeferredExampleDashboardPage />,
          },
          {
            path: '/example/page/table',
            element: <DeferredExampleTablePage />,
          },
          { path: '/example/page/chat', element: <DeferredExampleChatPage /> },
          {
            path: '/dashboard',
            element: <Navigate to={defaultAuthedPath} replace />,
          },
          {
            path: '/accounts',
            element: <Navigate to={defaultAuthedPath} replace />,
          },
          {
            path: '/models',
            element: <Navigate to={defaultAuthedPath} replace />,
          },
          {
            path: '/creative-console',
            element: <Navigate to='/example/page/chat' replace />,
          },
          {
            path: '/client-keys',
            element: <Navigate to={defaultAuthedPath} replace />,
          },
          {
            path: '/gallery',
            element: <Navigate to='/example/component' replace />,
          },
          {
            path: '/video-gallery',
            element: <Navigate to='/example/component' replace />,
          },
          {
            path: '/request-audits',
            element: <Navigate to='/example/page/table' replace />,
          },
          {
            path: '/docs',
            element: <Navigate to={defaultAuthedPath} replace />,
          },
          {
            path: '/docs/:category/:endpoint',
            element: <Navigate to={defaultAuthedPath} replace />,
          },
          { path: '/settings', element: <DeferredSettingsPage /> },
        ],
      },
    ],
  },
  { path: '*', element: <Navigate to={defaultAuthedPath} replace /> },
]);
