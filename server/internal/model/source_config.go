package model

import "time"

const (
	SourceConfigStatusParsed              = "parsed"
	SourceConfigStatusImported            = "imported"
	SourceConfigSourceTypeUpload          = "upload"
	SourceConfigSourceTypeSubscriptionURL = "subscription_url"
)

type SourceConfig struct {
	ID             int        `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceType     string     `json:"source_type" gorm:"size:32;index;not null;default:upload"`
	SourceURL      string     `json:"source_url" gorm:"size:2048"`
	ContentType    string     `json:"content_type" gorm:"size:255"`
	FetchedAt      *time.Time `json:"fetched_at" gorm:"index"`
	Filename       string     `json:"filename" gorm:"size:255;index;not null"`
	ContentHash    string     `json:"content_hash" gorm:"size:64;index;not null"`
	RawContent     string     `json:"-" gorm:"type:text;not null"`
	Status         string     `json:"status" gorm:"size:32;index;not null;default:parsed"`
	TotalNodes     int        `json:"total_nodes" gorm:"not null;default:0"`
	ValidNodes     int        `json:"valid_nodes" gorm:"not null;default:0"`
	InvalidNodes   int        `json:"invalid_nodes" gorm:"not null;default:0"`
	DuplicateNodes int        `json:"duplicate_nodes" gorm:"not null;default:0"`
	ImportedNodes  int        `json:"imported_nodes" gorm:"not null;default:0"`
	UploadedBy     string     `json:"uploaded_by" gorm:"size:64;index"`
	UploadedByID   int        `json:"uploaded_by_id" gorm:"index"`
	CreatedAt      time.Time  `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"index"`
}

func GetSourceConfigByID(id int) (*SourceConfig, error) {
	item := &SourceConfig{}
	if err := DB.First(item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return item, nil
}
