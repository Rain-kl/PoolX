package model

import "time"

type PortProfileTemplate struct {
	ID                    int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                  string    `json:"name" gorm:"size:120;index;not null"`
	ListenHost            string    `json:"listen_host" gorm:"size:120;not null;default:127.0.0.1"`
	MixedPort             int       `json:"mixed_port" gorm:"not null"`
	SocksPort             int       `json:"socks_port" gorm:"not null;default:0"`
	HTTPPort              int       `json:"http_port" gorm:"not null;default:0"`
	StrategyType          string    `json:"strategy_type" gorm:"size:32;not null;default:select"`
	StrategyGroupName     string    `json:"strategy_group_name" gorm:"size:120;not null;default:POOLX"`
	TestURL               string    `json:"test_url" gorm:"size:255;not null"`
	TestIntervalSeconds   int       `json:"test_interval_seconds" gorm:"not null;default:300"`
	LoadBalanceStrategy   string    `json:"load_balance_strategy" gorm:"size:32;not null;default:consistent-hashing"`
	LoadBalanceLazy       bool      `json:"load_balance_lazy" gorm:"not null;default:false"`
	LoadBalanceDisableUDP bool      `json:"load_balance_disable_udp" gorm:"not null;default:false"`
	IncludeInRuntime      bool      `json:"include_in_runtime" gorm:"index;not null;default:true"`
	NodeIDsJSON           string    `json:"-" gorm:"type:text;not null"`
	CreatedAt             time.Time `json:"created_at" gorm:"index"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"index"`
}

func ListPortProfileTemplates() ([]*PortProfileTemplate, error) {
	var items []*PortProfileTemplate
	if err := DB.Order("id desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetPortProfileTemplateByID(id int) (*PortProfileTemplate, error) {
	item := &PortProfileTemplate{}
	if err := DB.First(item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return item, nil
}
