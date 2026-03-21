import type { Metadata } from 'next';
import Script from 'next/script';
import type { ReactNode } from 'react';

import { AppProviders } from '@/components/providers/app-providers';
import { getThemeInitScript } from '@/lib/theme/theme';

import './globals.css';

export const metadata: Metadata = {
  title: {
    default: 'PoolX 控制台',
    template: '%s | PoolX',
  },
  description: 'PoolX 代理池控制台',
  applicationName: 'PoolX',
};

interface RootLayoutProps {
  children: ReactNode;
}

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html lang='zh-CN' suppressHydrationWarning>
      <body>
        <Script id='theme-init' strategy='beforeInteractive'>
          {getThemeInitScript()}
        </Script>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
