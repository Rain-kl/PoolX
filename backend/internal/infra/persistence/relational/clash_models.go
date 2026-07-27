package relational

import (
	"encoding/json"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/domain/clash"
)

type sourceConfigModel struct {
	ID             int        `gorm:"primaryKey;autoIncrement"`
	SourceType     string     `gorm:"size:32;not null;default:'upload'"`
	SourceURL      string     `gorm:"size:2048;not null;default:''"`
	ContentType    string     `gorm:"size:255;not null;default:''"`
	FetchedAt      *time.Time `gorm:""`
	Filename       string     `gorm:"size:255;not null;default:''"`
	ContentHash    string     `gorm:"size:64;not null;default:''"`
	RawContent     string     `gorm:"type:text;not null;default:''"`
	Status         string     `gorm:"size:32;not null;default:'parsed'"`
	TotalNodes     int        `gorm:"not null;default:0"`
	ValidNodes     int        `gorm:"not null;default:0"`
	InvalidNodes   int        `gorm:"not null;default:0"`
	DuplicateNodes int        `gorm:"not null;default:0"`
	ImportedNodes  int        `gorm:"not null;default:0"`
	UploadedBy     string     `gorm:"size:64;not null;default:''"`
	UploadedByID   int        `gorm:"not null;default:0"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
}

func (sourceConfigModel) TableName() string { return "source_configs" }

func toSourceConfigDomain(m *sourceConfigModel) *clash.SourceConfig {
	if m == nil {
		return nil
	}
	return &clash.SourceConfig{
		ID:             m.ID,
		SourceType:     m.SourceType,
		SourceURL:      m.SourceURL,
		ContentType:    m.ContentType,
		FetchedAt:      m.FetchedAt,
		Filename:       m.Filename,
		ContentHash:    m.ContentHash,
		RawContent:     m.RawContent,
		Status:         m.Status,
		TotalNodes:     m.TotalNodes,
		ValidNodes:     m.ValidNodes,
		InvalidNodes:   m.InvalidNodes,
		DuplicateNodes: m.DuplicateNodes,
		ImportedNodes:  m.ImportedNodes,
		UploadedBy:     m.UploadedBy,
		UploadedByID:   m.UploadedByID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func toSourceConfigModel(d *clash.SourceConfig) *sourceConfigModel {
	if d == nil {
		return nil
	}
	return &sourceConfigModel{
		ID:             d.ID,
		SourceType:     d.SourceType,
		SourceURL:      d.SourceURL,
		ContentType:    d.ContentType,
		FetchedAt:      d.FetchedAt,
		Filename:       d.Filename,
		ContentHash:    d.ContentHash,
		RawContent:     d.RawContent,
		Status:         d.Status,
		TotalNodes:     d.TotalNodes,
		ValidNodes:     d.ValidNodes,
		InvalidNodes:   d.InvalidNodes,
		DuplicateNodes: d.DuplicateNodes,
		ImportedNodes:  d.ImportedNodes,
		UploadedBy:     d.UploadedBy,
		UploadedByID:   d.UploadedByID,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

type proxyNodeModel struct {
	ID               int        `gorm:"primaryKey;autoIncrement"`
	SourceConfigID   int        `gorm:"not null;default:0"`
	SourceConfigName string     `gorm:"size:255;not null;default:''"`
	Name             string     `gorm:"size:255;not null;default:''"`
	Type             string     `gorm:"size:64;not null;default:''"`
	Server           string     `gorm:"size:255;not null;default:''"`
	Port             int        `gorm:"not null;default:0"`
	Tags             string     `gorm:"size:255;not null;default:''"`
	Fingerprint      string     `gorm:"size:64;not null;default:''"`
	MetadataJSON     string     `gorm:"type:text;not null;default:''"`
	Enabled          bool       `gorm:"not null;default:true"`
	LastTestStatus   string     `gorm:"size:32;not null;default:'unknown'"`
	LastLatencyMS    *int       `gorm:""`
	LastTestError    string     `gorm:"type:text;not null;default:''"`
	LastTestedAt     *time.Time `gorm:""`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
}

func (proxyNodeModel) TableName() string { return "proxy_nodes" }

func toProxyNodeDomain(m *proxyNodeModel) *clash.ProxyNode {
	if m == nil {
		return nil
	}
	return &clash.ProxyNode{
		ID:               m.ID,
		SourceConfigID:   m.SourceConfigID,
		SourceConfigName: m.SourceConfigName,
		Name:             m.Name,
		Type:             m.Type,
		Server:           m.Server,
		Port:             m.Port,
		Tags:             m.Tags,
		Fingerprint:      m.Fingerprint,
		MetadataJSON:     m.MetadataJSON,
		Enabled:          m.Enabled,
		LastTestStatus:   m.LastTestStatus,
		LastLatencyMS:    m.LastLatencyMS,
		LastTestError:    m.LastTestError,
		LastTestedAt:     m.LastTestedAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func toProxyNodeModel(d *clash.ProxyNode) *proxyNodeModel {
	if d == nil {
		return nil
	}
	return &proxyNodeModel{
		ID:               d.ID,
		SourceConfigID:   d.SourceConfigID,
		SourceConfigName: d.SourceConfigName,
		Name:             d.Name,
		Type:             d.Type,
		Server:           d.Server,
		Port:             d.Port,
		Tags:             d.Tags,
		Fingerprint:      d.Fingerprint,
		MetadataJSON:     d.MetadataJSON,
		Enabled:          d.Enabled,
		LastTestStatus:   d.LastTestStatus,
		LastLatencyMS:    d.LastLatencyMS,
		LastTestError:    d.LastTestError,
		LastTestedAt:     d.LastTestedAt,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

type nodeTestResultModel struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	NodeID       int       `gorm:"not null"`
	TestType     string    `gorm:"size:32;not null;default:'delay'"`
	Success      bool      `gorm:"not null;default:false"`
	LatencyMS    int       `gorm:"not null;default:0"`
	ErrorMessage string    `gorm:"type:text;not null;default:''"`
	TestedAt     time.Time `gorm:"not null"`
}

func (nodeTestResultModel) TableName() string { return "node_test_results" }

func toNodeTestResultDomain(m *nodeTestResultModel) *clash.NodeTestResult {
	if m == nil {
		return nil
	}
	return &clash.NodeTestResult{
		ID:           m.ID,
		NodeID:       m.NodeID,
		TestType:     m.TestType,
		Success:      m.Success,
		LatencyMS:    m.LatencyMS,
		ErrorMessage: m.ErrorMessage,
		TestedAt:     m.TestedAt,
	}
}

type portProfileModel struct {
	ID                int       `gorm:"primaryKey;autoIncrement"`
	Name              string    `gorm:"size:120;not null;default:''"`
	ListenHost        string    `gorm:"size:120;not null;default:'0.0.0.0'"`
	MixedPort         int       `gorm:"not null;default:0"`
	SocksPort         int       `gorm:"not null;default:0"`
	HTTPPort          int       `gorm:"not null;default:0"`
	ProxySettingsJSON string    `gorm:"type:text;not null;default:''"`
	IncludeInRuntime  bool      `gorm:"not null;default:true"`
	KernelType        string    `gorm:"size:32;not null;default:'mihomo'"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

func (portProfileModel) TableName() string { return "port_profiles" }

func toPortProfileDomain(m *portProfileModel) *clash.PortProfile {
	if m == nil {
		return nil
	}
	settings := clash.DefaultPortProfileProxySettings()
	if m.ProxySettingsJSON != "" {
		_ = json.Unmarshal([]byte(m.ProxySettingsJSON), &settings)
	}
	return &clash.PortProfile{
		ID:               m.ID,
		Name:             m.Name,
		ListenHost:       m.ListenHost,
		MixedPort:        m.MixedPort,
		SocksPort:        m.SocksPort,
		HTTPPort:         m.HTTPPort,
		ProxySettings:    settings,
		IncludeInRuntime: m.IncludeInRuntime,
		KernelType:       m.KernelType,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func toPortProfileModel(d *clash.PortProfile) *portProfileModel {
	if d == nil {
		return nil
	}
	settingsJSON, _ := json.Marshal(d.ProxySettings)
	return &portProfileModel{
		ID:                d.ID,
		Name:              d.Name,
		ListenHost:        d.ListenHost,
		MixedPort:         d.MixedPort,
		SocksPort:         d.SocksPort,
		HTTPPort:          d.HTTPPort,
		ProxySettingsJSON: string(settingsJSON),
		IncludeInRuntime:  d.IncludeInRuntime,
		KernelType:        d.KernelType,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

type portProfileTemplateModel struct {
	ID                int       `gorm:"primaryKey;autoIncrement"`
	Name              string    `gorm:"size:120;not null;default:''"`
	Description       string    `gorm:"size:255;not null;default:''"`
	MixedPort         int       `gorm:"not null;default:0"`
	ProxySettingsJSON string    `gorm:"type:text;not null;default:''"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

func (portProfileTemplateModel) TableName() string { return "port_profile_templates" }

func toPortProfileTemplateDomain(m *portProfileTemplateModel) *clash.PortProfileTemplate {
	if m == nil {
		return nil
	}
	settings := clash.DefaultPortProfileProxySettings()
	if m.ProxySettingsJSON != "" {
		_ = json.Unmarshal([]byte(m.ProxySettingsJSON), &settings)
	}
	return &clash.PortProfileTemplate{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		MixedPort:     m.MixedPort,
		ProxySettings: settings,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toPortProfileTemplateModel(d *clash.PortProfileTemplate) *portProfileTemplateModel {
	if d == nil {
		return nil
	}
	settingsJSON, _ := json.Marshal(d.ProxySettings)
	return &portProfileTemplateModel{
		ID:                d.ID,
		Name:              d.Name,
		Description:       d.Description,
		MixedPort:         d.MixedPort,
		ProxySettingsJSON: string(settingsJSON),
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

type portProfileNodeModel struct {
	ID              int       `gorm:"primaryKey;autoIncrement"`
	PortProfileID   int       `gorm:"not null"`
	ProxyNodeID     int       `gorm:"not null;default:0"`
	NodeFingerprint string    `gorm:"size:255;not null;default:''"`
	SortOrder       int       `gorm:"not null;default:0"`
	CreatedAt       time.Time `gorm:"not null"`
}

func (portProfileNodeModel) TableName() string { return "port_profile_nodes" }

type runtimeConfigModel struct {
	ID             int       `gorm:"primaryKey;autoIncrement"`
	PortProfileID  int       `gorm:"not null"`
	KernelType     string    `gorm:"size:32;not null;default:'mihomo'"`
	Checksum       string    `gorm:"size:64;not null;default:''"`
	RenderedConfig string    `gorm:"type:text;not null;default:''"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (runtimeConfigModel) TableName() string { return "runtime_configs" }

func toRuntimeConfigDomain(m *runtimeConfigModel) *clash.RuntimeConfig {
	if m == nil {
		return nil
	}
	return &clash.RuntimeConfig{
		ID:             m.ID,
		PortProfileID:  m.PortProfileID,
		KernelType:     m.KernelType,
		Checksum:       m.Checksum,
		RenderedConfig: m.RenderedConfig,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

type kernelInstanceModel struct {
	ID                   int        `gorm:"primaryKey;autoIncrement"`
	KernelType           string     `gorm:"size:32;not null;default:'mihomo'"`
	Status               string     `gorm:"size:32;not null;default:'stopped'"`
	PID                  *int       `gorm:""`
	WorkDir              string     `gorm:"size:255;not null;default:''"`
	ConfigPath           string     `gorm:"size:255;not null;default:''"`
	ControllerAddress    string     `gorm:"size:255;not null;default:''"`
	ControllerSecret     string     `gorm:"size:255;not null;default:''"`
	ActiveConfigChecksum string     `gorm:"size:64;not null;default:''"`
	ActiveProfileCount   int        `gorm:"not null;default:0"`
	ActiveListenerCount  int        `gorm:"not null;default:0"`
	LastAction           string     `gorm:"size:32;not null;default:''"`
	LastError            string     `gorm:"type:text;not null;default:''"`
	LastStartedAt        *time.Time `gorm:""`
	LastStoppedAt        *time.Time `gorm:""`
	LastReloadedAt       *time.Time `gorm:""`
	CreatedAt            time.Time  `gorm:"not null"`
	UpdatedAt            time.Time  `gorm:"not null"`
}

func (kernelInstanceModel) TableName() string { return "kernel_instances" }

func toKernelInstanceDomain(m *kernelInstanceModel) *clash.KernelInstance {
	if m == nil {
		return nil
	}
	return &clash.KernelInstance{
		ID:                   m.ID,
		KernelType:           m.KernelType,
		Status:               m.Status,
		PID:                  m.PID,
		WorkDir:              m.WorkDir,
		ConfigPath:           m.ConfigPath,
		ControllerAddress:    m.ControllerAddress,
		ControllerSecret:     m.ControllerSecret,
		ActiveConfigChecksum: m.ActiveConfigChecksum,
		ActiveProfileCount:   m.ActiveProfileCount,
		ActiveListenerCount:  m.ActiveListenerCount,
		LastAction:           m.LastAction,
		LastError:            m.LastError,
		LastStartedAt:        m.LastStartedAt,
		LastStoppedAt:        m.LastStoppedAt,
		LastReloadedAt:       m.LastReloadedAt,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

func toKernelInstanceModel(d *clash.KernelInstance) *kernelInstanceModel {
	if d == nil {
		return nil
	}
	return &kernelInstanceModel{
		ID:                   d.ID,
		KernelType:           d.KernelType,
		Status:               d.Status,
		PID:                  d.PID,
		WorkDir:              d.WorkDir,
		ConfigPath:           d.ConfigPath,
		ControllerAddress:    d.ControllerAddress,
		ControllerSecret:     d.ControllerSecret,
		ActiveConfigChecksum: d.ActiveConfigChecksum,
		ActiveProfileCount:   d.ActiveProfileCount,
		ActiveListenerCount:  d.ActiveListenerCount,
		LastAction:           d.LastAction,
		LastError:            d.LastError,
		LastStartedAt:        d.LastStartedAt,
		LastStoppedAt:        d.LastStoppedAt,
		LastReloadedAt:       d.LastReloadedAt,
		CreatedAt:            d.CreatedAt,
		UpdatedAt:            d.UpdatedAt,
	}
}
