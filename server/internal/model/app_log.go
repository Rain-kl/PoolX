package model

import "time"

const (
	AppLogClassificationSystem   = "system"
	AppLogClassificationBusiness = "business"
	AppLogClassificationSecurity = "security"

	AppLogLevelDebug = "debug"
	AppLogLevelInfo  = "info"
	AppLogLevelWarn  = "warn"
	AppLogLevelError = "error"
)

type AppLog struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Classification string    `json:"classification" gorm:"size:32;index;not null"`
	Level          string    `json:"level" gorm:"size:16;index;not null;default:info"`
	Message        string    `json:"message" gorm:"type:text;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"index"`
}

func GetRecentAppLogs(limit int, classification string) ([]*AppLog, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	query := DB.Order("id desc").Limit(limit)
	if classification != "" {
		query = query.Where("classification = ?", classification)
	}

	var logs []*AppLog
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	for left, right := 0, len(logs)-1; left < right; left, right = left+1, right-1 {
		logs[left], logs[right] = logs[right], logs[left]
	}

	return logs, nil
}

func GetAppLogsAfterID(afterID int, limit int, classification string) ([]*AppLog, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	query := DB.Order("id asc").Limit(limit).Where("id > ?", afterID)
	if classification != "" {
		query = query.Where("classification = ?", classification)
	}

	var logs []*AppLog
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func CreateAppLog(logEntry *AppLog) error {
	return DB.Create(logEntry).Error
}
