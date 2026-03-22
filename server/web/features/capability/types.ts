export interface KernelCapability {
  kernel_type: string;
  binary_configured: boolean;
  binary_exists: boolean;
  supports_start: boolean;
  supports_stop: boolean;
  supports_reload: boolean;
  supports_templates: boolean;
  supports_node_tags: boolean;
  supports_auto_refresh: boolean;
  supports_node_test_cache: boolean;
  supports_node_test_batch: boolean;
  supported_strategies: string[];
  runtime_controller_type: string;
  message: string;
}
