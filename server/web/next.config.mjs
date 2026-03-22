const devBackendUrl =
  process.env.NEXT_DEV_BACKEND_URL?.replace(/\/+$/, '') || 'http://127.0.0.1:3000';
const enableDevProxy = process.env.NEXT_DEV_PROXY === 'true';

export default function createNextConfig() {
  const nextConfig = {
    output: 'export',
    pageExtensions: ['ts', 'tsx'],
    images: {
      unoptimized: true,
    },
    eslint: {
      dirs: ['app', 'components', 'features', 'hooks', 'lib', 'store', 'tests', 'types'],
    },
  };

  if (!enableDevProxy) {
    return nextConfig;
  }

  return {
    ...nextConfig,
    async rewrites() {
      return [
        {
          source: '/api/:path*',
          destination: `${devBackendUrl}/api/:path*`,
        },
      ];
    },
  };
}
