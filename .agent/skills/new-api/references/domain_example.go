// Package references holds illustrative snippets for the new-api skill.
// Real code lives under backend/internal/domain/<name>/.
//
// 本文件演示 domain 层：纯实体，禁止 import Gin / GORM / HTTP。
package references

import "time"

// Widget is a sample domain entity after cloning the example slice.
// Place at: backend/internal/domain/widget/widget.go  (package widget)
type Widget struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
