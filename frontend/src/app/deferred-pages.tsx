import {
  lazy,
  Suspense,
  type ComponentType,
  type LazyExoticComponent,
} from 'react';

import { Spinner } from '@/components/ui/spinner';

const AppShell = lazyNamed(() => import('@/app/app-shell'), 'AppShell');
const ExampleComponentPage = lazyNamed(
  () => import('@/features/example/component-page'),
  'ExampleComponentPage',
);
const ExampleChatPage = lazyNamed(
  () => import('@/features/example/pages/chat-page'),
  'ExampleChatPage',
);
const ExampleDashboardPage = lazyNamed(
  () => import('@/features/example/pages/dashboard-page'),
  'ExampleDashboardPage',
);
const ExampleTablePage = lazyNamed(
  () => import('@/features/example/pages/table-page'),
  'ExampleTablePage',
);
const SettingsPage = lazyNamed(
  () => import('@/features/settings/settings-page'),
  'SettingsPage',
);

function lazyNamed<T extends Record<K, ComponentType>, K extends keyof T>(
  loader: () => Promise<T>,
  exportName: K,
): LazyExoticComponent<T[K]> {
  return lazy(async () => ({ default: (await loader())[exportName] }));
}

function DeferredPage({ page: Page }: { page: ComponentType }) {
  return (
    <Suspense fallback={<PageLoadingFallback />}>
      <Page />
    </Suspense>
  );
}

export function DeferredAppShell() {
  return (
    <Suspense fallback={<PageLoadingFallback fullScreen />}>
      <AppShell />
    </Suspense>
  );
}

export function DeferredExampleComponentPage() {
  return <DeferredPage page={ExampleComponentPage} />;
}

export function DeferredExampleDashboardPage() {
  return <DeferredPage page={ExampleDashboardPage} />;
}

export function DeferredExampleTablePage() {
  return <DeferredPage page={ExampleTablePage} />;
}

export function DeferredExampleChatPage() {
  return <DeferredPage page={ExampleChatPage} />;
}

export function DeferredSettingsPage() {
  return <DeferredPage page={SettingsPage} />;
}

function PageLoadingFallback({ fullScreen = false }: { fullScreen?: boolean }) {
  return (
    <div
      className={
        fullScreen
          ? 'flex min-h-screen items-center justify-center bg-background'
          : 'flex min-h-[calc(100vh-7rem)] items-center justify-center lg:min-h-[calc(100vh-10rem)]'
      }
    >
      <Spinner className='size-5' />
    </div>
  );
}
