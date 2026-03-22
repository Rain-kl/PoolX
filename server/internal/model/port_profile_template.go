package model

import "time"

type PortProfileTemplate struct {
	ID                int                      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string                   `json:"name" gorm:"size:120;index;not null"`
	ListenHost        string                   `json:"listen_host" gorm:"size:120;not null;default:127.0.0.1"`
	MixedPort         int                      `json:"mixed_port" gorm:"not null"`
	SocksPort         int                      `json:"socks_port" gorm:"not null;default:0"`
	HTTPPort          int                      `json:"http_port" gorm:"not null;default:0"`
	ProxySettingsJSON string                   `json:"-" gorm:"type:text;not null"`
	ProxySettings     PortProfileProxySettings `json:"proxy_settings" gorm:"-"`
	IncludeInRuntime  bool                     `json:"include_in_runtime" gorm:"index;not null;default:true"`
	NodeIDsJSON       string                   `json:"-" gorm:"type:text;not null"`
	CreatedAt         time.Time                `json:"created_at" gorm:"index"`
	UpdatedAt         time.Time                `json:"updated_at" gorm:"index"`
}

func (template *PortProfileTemplate) HydrateProxySettings() error {
	settings, err := ParsePortProfileProxySettings(template.ProxySettingsJSON)
	if err != nil {
		return err
	}
	template.ProxySettings = settings
	return nil
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
