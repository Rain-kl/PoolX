package model

import "time"

const (
	currentDatabaseSchemaVersion = 5
	databaseSchemaVersionRowID   = 1
)

type DatabaseSchemaVersion struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Version   int       `json:"version" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DatabaseSchemaVersion) TableName() string {
	return "database_schema_versions"
}
