package model

import "time"

const (
	PortProfileStrategySelect      = "select"
	PortProfileStrategyURLTest     = "url-test"
	PortProfileStrategyFallback    = "fallback"
	PortProfileStrategyLoadBalance = "load-balance"
)

type PortProfile struct {
	ID                  int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                string    `json:"name" gorm:"size:120;index;not null"`
	ListenHost          string    `json:"listen_host" gorm:"size:120;not null;default:127.0.0.1"`
	MixedPort           int       `json:"mixed_port" gorm:"not null"`
	SocksPort           int       `json:"socks_port" gorm:"not null;default:0"`
	HTTPPort            int       `json:"http_port" gorm:"not null;default:0"`
	StrategyType        string    `json:"strategy_type" gorm:"size:32;not null;default:select"`
	StrategyGroupName   string    `json:"strategy_group_name" gorm:"size:120;not null;default:POOLX"`
	TestURL             string    `json:"test_url" gorm:"size:255;not null"`
	TestIntervalSeconds int       `json:"test_interval_seconds" gorm:"not null;default:300"`
	Enabled             bool      `json:"enabled" gorm:"index;not null;default:true"`
	KernelType          string    `json:"kernel_type" gorm:"size:32;not null;default:mihomo"`
	CreatedAt           time.Time `json:"created_at" gorm:"index"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"index"`
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
