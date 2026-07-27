package clash

import "time"

// SourceConfig Constants
const (
	SourceConfigStatusParsed              = "parsed"
	SourceConfigStatusImported            = "imported"
	SourceConfigSourceTypeUpload          = "upload"
	SourceConfigSourceTypeSubscriptionURL = "subscription_url"
)

// ProxyNode Constants
const (
	NodeTestStatusUnknown = "unknown"
	NodeTestStatusSuccess = "success"
	NodeTestStatusFailed  = "failed"
)

// PortProfile Strategy Constants
const (
	PortProfileStrategySelect      = "select"
	PortProfileStrategyURLTest     = "url-test"
	PortProfileStrategyFallback    = "fallback"
	PortProfileStrategyLoadBalance = "load-balance"
)

// KernelInstance Status Constants
const (
	KernelInstanceStatusStopped   = "stopped"
	KernelInstanceStatusStarting  = "starting"
	KernelInstanceStatusRunning   = "running"
	KernelInstanceStatusStopping  = "stopping"
	KernelInstanceStatusError     = "error"
	KernelInstanceStatusReloading = "reloading"
)

// SourceConfig represents a configuration file upload or subscription URL
type SourceConfig struct {
	ID             int        `json:"id"`
	SourceType     string     `json:"source_type"`
	SourceURL      string     `json:"source_url"`
	ContentType    string     `json:"content_type"`
	FetchedAt      *time.Time `json:"fetched_at"`
	Filename       string     `json:"filename"`
	ContentHash    string     `json:"content_hash"`
	RawContent     string     `json:"-"`
	Status         string     `json:"status"`
	TotalNodes     int        `json:"total_nodes"`
	ValidNodes     int        `json:"valid_nodes"`
	InvalidNodes   int        `json:"invalid_nodes"`
	DuplicateNodes int        `json:"duplicate_nodes"`
	ImportedNodes  int        `json:"imported_nodes"`
	UploadedBy     string     `json:"uploaded_by"`
	UploadedByID   int        `json:"uploaded_by_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ProxyNode represents a proxy node entity
type ProxyNode struct {
	ID               int        `json:"id"`
	SourceConfigID   int        `json:"source_config_id"`
	SourceConfigName string     `json:"source_config_name"`
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Server           string     `json:"server"`
	Port             int        `json:"port"`
	Tags             string     `json:"tags"`
	Fingerprint      string     `json:"-"`
	MetadataJSON     string     `json:"metadata_json"`
	Enabled          bool       `json:"enabled"`
	LastTestStatus   string     `json:"last_test_status"`
	LastLatencyMS    *int       `json:"last_latency_ms,omitempty"`
	LastTestError    string     `json:"last_test_error,omitempty"`
	LastTestedAt     *time.Time `json:"last_tested_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// NodeTestResult represents a test attempt result
type NodeTestResult struct {
	ID           int       `json:"id"`
	NodeID       int       `json:"node_id"`
	TestType     string    `json:"test_type"`
	Success      bool      `json:"success"`
	LatencyMS    int       `json:"latency_ms"`
	ErrorMessage string    `json:"error_message"`
	TestedAt     time.Time `json:"tested_at"`
}

// PortProfileProxySettings holds proxy strategy options for a port profile
type PortProfileProxySettings struct {
	StrategyType          string `json:"strategy_type"`
	TestURL               string `json:"test_url"`
	TestIntervalSeconds   int    `json:"test_interval_seconds"`
	LoadBalanceStrategy   string `json:"load_balance_strategy"`
	LoadBalanceLazy       bool   `json:"load_balance_lazy"`
	LoadBalanceDisableUDP bool   `json:"load_balance_disable_udp"`
	UDPEnabled            bool   `json:"udp_enabled"`
	AuthEnabled           bool   `json:"auth_enabled"`
	AuthUsername          string `json:"auth_username"`
	AuthPassword          string `json:"auth_password"`
}

// DefaultPortProfileProxySettings returns sensible defaults
func DefaultPortProfileProxySettings() PortProfileProxySettings {
	return PortProfileProxySettings{
		StrategyType:        PortProfileStrategySelect,
		TestURL:             "https://cp.cloudflare.com/generate_204",
		TestIntervalSeconds: 300,
		LoadBalanceStrategy: "consistent-hashing",
		UDPEnabled:          true,
		AuthEnabled:         false,
	}
}

// PortProfile represents an inbound listening port profile
type PortProfile struct {
	ID               int                      `json:"id"`
	Name             string                   `json:"name"`
	ListenHost       string                   `json:"listen_host"`
	MixedPort        int                      `json:"mixed_port"`
	SocksPort        int                      `json:"socks_port"`
	HTTPPort         int                      `json:"http_port"`
	ProxySettings    PortProfileProxySettings `json:"proxy_settings"`
	IncludeInRuntime bool                     `json:"include_in_runtime"`
	KernelType       string                   `json:"kernel_type"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

// PortProfileTemplate represents a reusable template for creating port profiles
type PortProfileTemplate struct {
	ID            int                      `json:"id"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	MixedPort     int                      `json:"mixed_port"`
	ProxySettings PortProfileProxySettings `json:"proxy_settings"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

// PortProfileNode represents the binding between a port profile and a node (with ID & fingerprint persistence)
type PortProfileNode struct {
	ID              int       `json:"id"`
	PortProfileID   int       `json:"port_profile_id"`
	ProxyNodeID     int       `json:"proxy_node_id"`
	NodeFingerprint string    `json:"node_fingerprint"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
}

// RuntimeConfig holds the generated configuration snippet for a port profile
type RuntimeConfig struct {
	ID             int       `json:"id"`
	PortProfileID  int       `json:"port_profile_id"`
	KernelType     string    `json:"kernel_type"`
	Checksum       string    `json:"checksum"`
	RenderedConfig string    `json:"rendered_config"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// KernelInstance represents the runtime process & state of Mihomo kernel
type KernelInstance struct {
	ID                   int        `json:"id"`
	KernelType           string     `json:"kernel_type"`
	Status               string     `json:"status"`
	PID                  *int       `json:"pid,omitempty"`
	WorkDir              string     `json:"work_dir"`
	ConfigPath           string     `json:"config_path"`
	ControllerAddress    string     `json:"controller_address"`
	ControllerSecret     string     `json:"-"`
	ActiveConfigChecksum string     `json:"active_config_checksum"`
	ActiveProfileCount   int        `json:"active_profile_count"`
	ActiveListenerCount  int        `json:"active_listener_count"`
	LastAction           string     `json:"last_action"`
	LastError            string     `json:"last_error"`
	LastStartedAt        *time.Time `json:"last_started_at,omitempty"`
	LastStoppedAt        *time.Time `json:"last_stopped_at,omitempty"`
	LastReloadedAt       *time.Time `json:"last_reloaded_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// KernelCapability describes the features supported by a kernel
type KernelCapability struct {
	KernelType          string   `json:"kernel_type"`
	SupportsReload      bool     `json:"supports_reload"`
	SupportsExternalAPI bool     `json:"supports_external_api"`
	SupportsHealthCheck bool     `json:"supports_health_check"`
	SupportedStrategies []string `json:"supported_strategies"`
}

// ProxyNodeFilter defines query parameters for node listing
type ProxyNodeFilter struct {
	Keyword        string
	SourceConfigID int
	Enabled        *bool
}
