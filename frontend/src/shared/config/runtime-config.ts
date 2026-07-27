const apiBaseUrl =
  window.__POOLX_RUNTIME_CONFIG__?.apiBaseUrl?.replace(/\/$/, '') ?? '';
const configuredPublicApiBaseUrl =
  window.__POOLX_RUNTIME_CONFIG__?.publicApiBaseUrl?.replace(/\/$/, '') ?? '';
const developmentApiBaseUrl =
  typeof __POOLX_DEV_API_TARGET__ === 'string'
    ? __POOLX_DEV_API_TARGET__.replace(/\/$/, '')
    : '';

export const runtimeConfig = {
  apiBaseUrl,
  publicApiBaseUrl:
    configuredPublicApiBaseUrl ||
    apiBaseUrl ||
    developmentApiBaseUrl ||
    window.location.origin,
};
