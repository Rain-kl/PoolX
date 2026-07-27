package settings

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	settingsdomain "github.com/Rain-kl/Foam/backend/internal/domain/settings"
	"github.com/Rain-kl/Foam/backend/internal/infra/clash/kernel"
	"github.com/Rain-kl/Foam/backend/internal/infra/config"
	"github.com/Rain-kl/Foam/backend/internal/repository"
)

var (
	ErrInvalidInput = errors.New("运行设置参数无效")
	ErrConflict     = errors.New("运行设置已被其他会话更新")
)

// EditableConfig 是管理接口可写入的字段。
type EditableConfig struct {
	App      AppConfigInput      `json:"app"`
	Frontend FrontendConfigInput `json:"frontend"`
	Clash    ClashConfigInput    `json:"clash"`
}

type AppConfigInput struct {
	DisplayName string `json:"display_name"`
}

type FrontendConfigInput struct {
	PublicAPIBaseURL string `json:"public_api_base_url"`
}

type ClashConfigInput struct {
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

// Snapshot 是管理端读取的完整设置视图。
type Snapshot struct {
	Config    EditableConfig `json:"config"`
	Revision  uint64         `json:"revision"`
	UpdatedAt time.Time      `json:"updated_at"`
	// FilePublicAPIBaseURL 来自 config.yaml 的基线值（未覆盖时生效）。
	FilePublicAPIBaseURL string `json:"file_public_api_base_url"`
}

// Service 管理运行设置的内存镜像与持久化。
type Service struct {
	mu         sync.RWMutex
	repository repository.RuntimeSettingsRepository
	fileBase   config.Config
	current    settingsdomain.Config
	updatedAt  time.Time
	revision   uint64
}

func NewService(fileBase config.Config, current settingsdomain.Config, updatedAt time.Time, revision uint64, repo repository.RuntimeSettingsRepository) *Service {
	return &Service{
		repository: repo,
		fileBase:   fileBase,
		current:    current,
		updatedAt:  updatedAt,
		revision:   revision,
	}
}

// LoadPersisted 从仓储加载并与文件配置合并为初始 domain 配置。
func LoadPersisted(ctx context.Context, _ config.Config, repo repository.RuntimeSettingsRepository) (settingsdomain.Config, time.Time, uint64, error) {
	value, updatedAt, revision, found, err := repo.Get(ctx)
	if err != nil {
		return settingsdomain.Config{}, time.Time{}, 0, err
	}
	if !found {
		return settingsdomain.Config{}, time.Time{}, 0, nil
	}
	return value, updatedAt, revision, nil
}

func (s *Service) Get() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}

// PublicAPIBaseURL 返回生效的公开 API 根地址（运行覆盖优先）。
func (s *Service) PublicAPIBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if value := strings.TrimRight(strings.TrimSpace(s.current.Frontend.PublicAPIBaseURL), "/"); value != "" {
		return value
	}
	return s.fileBase.Frontend.EffectivePublicAPIBaseURL()
}

// DisplayName 返回生效的产品展示名。
func (s *Service) DisplayName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if value := strings.TrimSpace(s.current.App.DisplayName); value != "" {
		return value
	}
	return "Foam"
}

func (s *Service) Update(ctx context.Context, expectedRevision uint64, input EditableConfig) (Snapshot, error) {
	normalized, err := normalizeEditable(input)
	if err != nil {
		return Snapshot{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if expectedRevision != s.revision {
		return Snapshot{}, ErrConflict
	}

	next := settingsdomain.Config{
		App: settingsdomain.AppConfig{
			DisplayName: normalized.App.DisplayName,
		},
		Frontend: settingsdomain.FrontendConfig{
			PublicAPIBaseURL: normalized.Frontend.PublicAPIBaseURL,
		},
		Clash: settingsdomain.ClashConfig{
			KernelType:               normalized.Clash.KernelType,
			MihomoBinaryPath:         normalized.Clash.MihomoBinaryPath,
			MihomoBinaryVersion:      normalized.Clash.MihomoBinaryVersion,
			MihomoBinarySource:       normalized.Clash.MihomoBinarySource,
			ClashExternalController:  normalized.Clash.ClashExternalController,
			ClashMode:                normalized.Clash.ClashMode,
			ClashSecret:              normalized.Clash.ClashSecret,
			ClashAllowLAN:            normalized.Clash.ClashAllowLAN,
			NodeTestDefaultURL:       normalized.Clash.NodeTestDefaultURL,
			NodeTestDefaultTimeoutMS: normalized.Clash.NodeTestDefaultTimeoutMS,
		},
	}

	updatedAt, revision, err := s.repository.Save(ctx, next, expectedRevision)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return Snapshot{}, ErrConflict
		}
		return Snapshot{}, err
	}
	s.current = next
	s.updatedAt = updatedAt
	s.revision = revision
	return s.snapshotLocked(), nil
}

func (s *Service) snapshotLocked() Snapshot {
	clashConfig := s.current.Clash
	if clashConfig.KernelType == "" {
		clashConfig.KernelType = "mihomo"
	}
	if clashConfig.MihomoBinaryPath == "" {
		clashConfig.MihomoBinaryPath = kernel.DefaultMihomoBinaryPath()
	}
	if clashConfig.ClashExternalController == "" {
		clashConfig.ClashExternalController = "127.0.0.1:9090"
	}
	if clashConfig.ClashMode == "" {
		clashConfig.ClashMode = "rule"
	}
	if clashConfig.NodeTestDefaultURL == "" {
		clashConfig.NodeTestDefaultURL = "https://cp.cloudflare.com/generate_204"
	}
	if clashConfig.NodeTestDefaultTimeoutMS <= 0 {
		clashConfig.NodeTestDefaultTimeoutMS = 5000
	}

	return Snapshot{
		Config: EditableConfig{
			App: AppConfigInput{
				DisplayName: s.current.App.DisplayName,
			},
			Frontend: FrontendConfigInput{
				PublicAPIBaseURL: s.current.Frontend.PublicAPIBaseURL,
			},
			Clash: ClashConfigInput{
				KernelType:               clashConfig.KernelType,
				MihomoBinaryPath:         clashConfig.MihomoBinaryPath,
				MihomoBinaryVersion:      clashConfig.MihomoBinaryVersion,
				MihomoBinarySource:       clashConfig.MihomoBinarySource,
				ClashExternalController:  clashConfig.ClashExternalController,
				ClashMode:                clashConfig.ClashMode,
				ClashSecret:              clashConfig.ClashSecret,
				ClashAllowLAN:            clashConfig.ClashAllowLAN,
				NodeTestDefaultURL:       clashConfig.NodeTestDefaultURL,
				NodeTestDefaultTimeoutMS: clashConfig.NodeTestDefaultTimeoutMS,
			},
		},
		Revision:             s.revision,
		UpdatedAt:            s.updatedAt,
		FilePublicAPIBaseURL: s.fileBase.Frontend.EffectivePublicAPIBaseURL(),
	}
}

func normalizeEditable(input EditableConfig) (EditableConfig, error) {
	displayName := strings.TrimSpace(input.App.DisplayName)
	if utf8.RuneCountInString(displayName) > 64 {
		return EditableConfig{}, ErrInvalidInput
	}

	publicBase := strings.TrimSpace(input.Frontend.PublicAPIBaseURL)
	if publicBase != "" {
		publicBase = strings.TrimRight(publicBase, "/")
		parsed, err := url.Parse(publicBase)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return EditableConfig{}, ErrInvalidInput
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return EditableConfig{}, ErrInvalidInput
		}
	}

	clashIn := input.Clash
	if clashIn.KernelType == "" {
		clashIn.KernelType = "mihomo"
	}
	if clashIn.ClashExternalController == "" {
		clashIn.ClashExternalController = "127.0.0.1:9090"
	}
	if clashIn.ClashMode == "" {
		clashIn.ClashMode = "rule"
	}
	if clashIn.NodeTestDefaultURL == "" {
		clashIn.NodeTestDefaultURL = "https://cp.cloudflare.com/generate_204"
	}
	if clashIn.NodeTestDefaultTimeoutMS <= 0 {
		clashIn.NodeTestDefaultTimeoutMS = 5000
	}

	return EditableConfig{
		App:      AppConfigInput{DisplayName: displayName},
		Frontend: FrontendConfigInput{PublicAPIBaseURL: publicBase},
		Clash:    clashIn,
	}, nil
}
