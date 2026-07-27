package relational

import "time"

type adminModel struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:100;uniqueIndex;not null;check:chk_admins_username,length(trim(username)) BETWEEN 1 AND 100"`
	PasswordHash string    `gorm:"size:255;not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (adminModel) TableName() string { return "admins" }

// adminSessionModel maps admin_sessions. Schema is owned by goose SQL migrations;
// no physical foreign key (index-only relation, Wavelet convention).
type adminSessionModel struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement"`
	AdminID          uint64 `gorm:"not null;index"`
	RefreshTokenHash string `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt        time.Time
	LastUsedAt       *time.Time
	CreatedAt        time.Time `gorm:"not null"`
}

func (adminSessionModel) TableName() string { return "admin_sessions" }

type exampleModel struct {
	ID          string    `gorm:"primaryKey;size:36"`
	Name        string    `gorm:"size:160;not null;index;check:chk_examples_name,length(trim(name)) BETWEEN 1 AND 160"`
	Description string    `gorm:"size:1024;not null;default:''"`
	CreatedAt   time.Time `gorm:"not null;index"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (exampleModel) TableName() string { return "examples" }

type runtimeSettingsModel struct {
	Key       string    `gorm:"size:64;primaryKey;check:chk_runtime_settings_key,length(trim(key)) BETWEEN 1 AND 64"`
	ValueJSON string    `gorm:"type:text;not null;check:chk_runtime_settings_json_length,length(value_json) <= 1048576"`
	Revision  uint64    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (runtimeSettingsModel) TableName() string { return "runtime_settings" }
