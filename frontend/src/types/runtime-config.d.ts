export {};

declare global {
  const __POOLX_DEV_API_TARGET__: string;

  interface Window {
    __POOLX_RUNTIME_CONFIG__?: {
      apiBaseUrl?: string;
      publicApiBaseUrl?: string;
    };
  }
}
