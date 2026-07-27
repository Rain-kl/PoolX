import { apiRequest } from '@/shared/api/client';
import {
  createValidatedDecoder,
  hasShape,
  isNumber,
  isOptional,
  isString,
  isBoolean,
  type ApiDecoder,
} from '@/shared/api/decoder';

export type ClashSettingsConfig = {
  kernel_type?: string;
  mihomo_binary_path?: string;
  mihomo_binary_version?: string;
  mihomo_binary_source?: string;
  clash_external_controller?: string;
  clash_mode?: string;
  clash_secret?: string;
  clash_allow_lan?: boolean;
  node_test_default_url?: string;
  node_test_default_timeout_ms?: number;
};

export type SettingsSnapshot = {
  config: {
    app: { display_name: string };
    frontend: { public_api_base_url: string };
    clash?: ClashSettingsConfig;
  };
  revision: number;
  updated_at?: string;
  file_public_api_base_url: string;
  effective: {
    display_name: string;
    public_api_base_url: string;
  };
};

export type SettingsUpdateInput = {
  revision: number;
  config: SettingsSnapshot['config'];
};

const decodeSettingsSnapshot: ApiDecoder<SettingsSnapshot> =
  createValidatedDecoder(
    'settings snapshot',
    hasShape({
      config: hasShape({
        app: hasShape({ display_name: isString }),
        frontend: hasShape({ public_api_base_url: isString }),
        clash: isOptional(
          hasShape({
            kernel_type: isOptional(isString),
            mihomo_binary_path: isOptional(isString),
            mihomo_binary_version: isOptional(isString),
            mihomo_binary_source: isOptional(isString),
            clash_external_controller: isOptional(isString),
            clash_mode: isOptional(isString),
            clash_secret: isOptional(isString),
            clash_allow_lan: isOptional(isBoolean),
            node_test_default_url: isOptional(isString),
            node_test_default_timeout_ms: isOptional(isNumber),
          }),
        ),
      }),
      revision: isNumber,
      file_public_api_base_url: isString,
      effective: hasShape({
        display_name: isString,
        public_api_base_url: isString,
      }),
      updated_at: isOptional(isString),
    }),
  );

export async function fetchSettings(): Promise<SettingsSnapshot> {
  return apiRequest(
    '/api/v1/admin/settings',
    { method: 'GET' },
    decodeSettingsSnapshot,
  );
}

export async function updateSettings(
  input: SettingsUpdateInput,
): Promise<SettingsSnapshot> {
  return apiRequest(
    '/api/v1/admin/settings',
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    },
    decodeSettingsSnapshot,
  );
}
