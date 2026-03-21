package model

import "time"

type NodeTestResult struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	ProxyNodeID  int       `json:"proxy_node_id" gorm:"index;not null"`
	Status       string    `json:"status" gorm:"size:32;index;not null"`
	LatencyMS    *int      `json:"latency_ms,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty" gorm:"type:text"`
	TestURL      string    `json:"test_url,omitempty" gorm:"size:255"`
	DialAddress  string    `json:"dial_address" gorm:"size:255;not null"`
	StartedAt    time.Time `json:"started_at" gorm:"index"`
	FinishedAt   time.Time `json:"finished_at" gorm:"index"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}

func ListNodeTestResults(proxyNodeID int, limit int) ([]*NodeTestResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var items []*NodeTestResult
	if err := DB.Where("proxy_node_id = ?", proxyNodeID).
		Order("id desc").
		Limit(limit).
		Find(&items).
		Error; err != nil {
		return nil, err
	}
	return items, nil
}
