package model

import "time"

type RuntimeConfig struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	PortProfileID  int       `json:"port_profile_id" gorm:"uniqueIndex;not null"`
	KernelType     string    `json:"kernel_type" gorm:"size:32;not null"`
	Checksum       string    `json:"checksum" gorm:"size:64;not null"`
	RenderedConfig string    `json:"rendered_config" gorm:"type:text;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"index"`
}

func GetRuntimeConfigByPortProfileID(profileID int) (*RuntimeConfig, error) {
	item := &RuntimeConfig{}
	if err := DB.First(item, "port_profile_id = ?", profileID).Error; err != nil {
		return nil, err
	}
	return item, nil
}
