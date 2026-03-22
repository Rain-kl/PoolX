package model

import "time"

const (
	KernelInstanceStatusStopped   = "stopped"
	KernelInstanceStatusStarting  = "starting"
	KernelInstanceStatusRunning   = "running"
	KernelInstanceStatusStopping  = "stopping"
	KernelInstanceStatusError     = "error"
	KernelInstanceStatusReloading = "reloading"
)

type KernelInstance struct {
	ID                   int        `json:"id" gorm:"primaryKey;autoIncrement"`
	KernelType           string     `json:"kernel_type" gorm:"size:32;uniqueIndex;not null"`
	Status               string     `json:"status" gorm:"size:32;index;not null;default:stopped"`
	PID                  *int       `json:"pid,omitempty"`
	WorkDir              string     `json:"work_dir" gorm:"size:255;not null"`
	ConfigPath           string     `json:"config_path" gorm:"size:255;not null"`
	ControllerAddress    string     `json:"controller_address" gorm:"size:255;not null"`
	ControllerSecret     string     `json:"-" gorm:"size:255;not null"`
	ActiveConfigChecksum string     `json:"active_config_checksum" gorm:"size:64;not null;default:''"`
	ActiveProfileCount   int        `json:"active_profile_count" gorm:"not null;default:0"`
	ActiveListenerCount  int        `json:"active_listener_count" gorm:"not null;default:0"`
	LastAction           string     `json:"last_action" gorm:"size:32;not null;default:''"`
	LastError            string     `json:"last_error" gorm:"type:text"`
	LastStartedAt        *time.Time `json:"last_started_at,omitempty"`
	LastStoppedAt        *time.Time `json:"last_stopped_at,omitempty"`
	LastReloadedAt       *time.Time `json:"last_reloaded_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at" gorm:"index"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"index"`
}

func GetKernelInstanceByType(kernelType string) (*KernelInstance, error) {
	item := &KernelInstance{}
	if err := DB.First(item, "kernel_type = ?", kernelType).Error; err != nil {
		return nil, err
	}
	return item, nil
}
