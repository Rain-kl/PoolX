import path from 'node:path';

import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

/** Backend API target for Vite dev proxy (production still uses embedded static from PoolX). */
const devApiTarget = process.env.VITE_DEV_API_TARGET ?? 'http://127.0.0.1:8000';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  define: {
    // Empty by default: browser uses same-origin relative paths, Vite proxies to backend.
    // Set VITE_DEV_API_TARGET only when you need the client to call the backend absolute URL.
    __POOLX_DEV_API_TARGET__: JSON.stringify(
      process.env.VITE_DEV_API_TARGET ?? '',
    ),
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '127.0.0.1',
    port: 8010,
    strictPort: true,
    headers: {
      'X-Frame-Options': 'SAMEORIGIN',
    },
    proxy: {
      '/api': { target: devApiTarget, changeOrigin: true },
      '/healthz': { target: devApiTarget, changeOrigin: true },
      '/readyz': { target: devApiTarget, changeOrigin: true },
      '/swagger': { target: devApiTarget, changeOrigin: true },
      '/zashboard': {
        target: devApiTarget,
        changeOrigin: true,
        // Ensure the backend's X-Frame-Options header is preserved, not overwritten
        configure: (proxy) => {
          proxy.on('proxyRes', (proxyRes) => {
            proxyRes.headers['x-frame-options'] = 'SAMEORIGIN';
          });
        },
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
