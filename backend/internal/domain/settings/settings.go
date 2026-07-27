package settings

// Config 表示可跨重启持久化、支持管理端热更新的运行参数。
// 仅保留脚手架通用字段；业务模块可在二次开发时扩展此结构。
type Config struct {
	App      AppConfig
	Frontend FrontendConfig
	Clash    ClashConfig
}

// AppConfig 定义产品展示相关运行参数。
type AppConfig struct {
	// DisplayName 管理端展示名称；空字符串表示使用默认 "Foam"。
	DisplayName string
}

// FrontendConfig 定义前端/公开地址相关运行参数。
type FrontendConfig struct {
	// PublicAPIBaseURL 覆盖配置文件中的公开 API 根地址；空表示回退到 config.yaml。
	PublicAPIBaseURL string
}

// ClashConfig 定义代理内核、运行控制及默认测速设置。
type ClashConfig struct {
	KernelType               string `json:"kernel_type"`
	MihomoBinaryPath         string `json:"mihomo_binary_path"`
	MihomoBinaryVersion      string `json:"mihomo_binary_version"`
	MihomoBinarySource       string `json:"mihomo_binary_source"`
	ClashExternalController  string `json:"clash_external_controller"`
	ClashMode                string `json:"clash_mode"`
	ClashSecret              string `json:"clash_secret"`
	ClashAllowLAN            bool   `json:"clash_allow_lan"`
	NodeTestDefaultURL       string `json:"node_test_default_url"`
	NodeTestDefaultTimeoutMS int    `json:"node_test_default_timeout_ms"`
}
