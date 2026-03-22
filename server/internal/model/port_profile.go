package model

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	PortProfileStrategySelect      = "select"
	PortProfileStrategyURLTest     = "url-test"
	PortProfileStrategyFallback    = "fallback"
	PortProfileStrategyLoadBalance = "load-balance"
)

type PortProfile struct {
	ID                int                      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string                   `json:"name" gorm:"size:120;index;not null"`
	ListenHost        string                   `json:"listen_host" gorm:"size:120;not null;default:127.0.0.1"`
	MixedPort         int                      `json:"mixed_port" gorm:"not null"`
	SocksPort         int                      `json:"socks_port" gorm:"not null;default:0"`
	HTTPPort          int                      `json:"http_port" gorm:"not null;default:0"`
	ProxySettingsJSON string                   `json:"-" gorm:"type:text;not null"`
	ProxySettings     PortProfileProxySettings `json:"proxy_settings" gorm:"-"`
	IncludeInRuntime  bool                     `json:"include_in_runtime" gorm:"index;not null;default:true"`
	KernelType        string                   `json:"kernel_type" gorm:"size:32;not null;default:mihomo"`
	CreatedAt         time.Time                `json:"created_at" gorm:"index"`
	UpdatedAt         time.Time                `json:"updated_at" gorm:"index"`
}

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

func ParsePortProfileProxySettings(raw string) (PortProfileProxySettings, error) {
	settings := DefaultPortProfileProxySettings()
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return settings, nil
	}
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return DefaultPortProfileProxySettings(), err
	}
	return settings, nil
}

func EncodePortProfileProxySettings(settings PortProfileProxySettings) (string, error) {
	payload, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func (profile *PortProfile) HydrateProxySettings() error {
	settings, err := ParsePortProfileProxySettings(profile.ProxySettingsJSON)
	if err != nil {
		return err
	}
	profile.ProxySettings = settings
	return nil
}

type PortProfileNode struct {
	ID            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	PortProfileID int       `json:"port_profile_id" gorm:"index;not null"`
	ProxyNodeID   int       `json:"proxy_node_id" gorm:"index;not null"`
	SortOrder     int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt     time.Time `json:"created_at" gorm:"index"`
}

type PortProfileWithNodes struct {
	Profile PortProfile    `json:"profile"`
	NodeIDs []int          `json:"node_ids"`
	Nodes   []*ProxyNode   `json:"nodes,omitempty"`
	Runtime *RuntimeConfig `json:"runtime,omitempty"`
}

func ListPortProfiles() ([]*PortProfile, error) {
	var items []*PortProfile
	if err := DB.Order("id desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetPortProfileByID(id int) (*PortProfile, error) {
	item := &PortProfile{}
	if err := DB.First(item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func ListPortProfileNodes(profileID int) ([]*PortProfileNode, error) {
	var items []*PortProfileNode
	if err := DB.Where("port_profile_id = ?", profileID).Order("sort_order asc, id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
