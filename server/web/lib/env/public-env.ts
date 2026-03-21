import { z } from 'zod';

const publicEnvSchema = z.object({
  NEXT_PUBLIC_API_BASE_URL: z.string().default('/api'),
  NEXT_PUBLIC_APP_VERSION: z.string().default('dev'),
});

const rawEnv = {
  NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
  NEXT_PUBLIC_APP_VERSION: process.env.NEXT_PUBLIC_APP_VERSION,
};

const parsedEnv = publicEnvSchema.parse(rawEnv);

function normalizeBaseUrl(value: string) {
  if (value === '/') {
    return '';
  }

  return value.replace(/\/+$/, '');
}

export const publicEnv = {
  apiBaseUrl: normalizeBaseUrl(parsedEnv.NEXT_PUBLIC_API_BASE_URL),
  appVersion: parsedEnv.NEXT_PUBLIC_APP_VERSION,
};
