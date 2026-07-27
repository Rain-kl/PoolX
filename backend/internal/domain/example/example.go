package example

import "time"

// Example 是脚手架垂直切片的示例资源。
type Example struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
