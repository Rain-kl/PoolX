// Illustrative repository interface for the new-api skill.
//
// Real location: backend/internal/repository/widget.go (package repository)
// Implementation: backend/internal/infra/persistence/relational/widget_repository.go
//
// Reuse shared helpers from backend/internal/repository:
//
//	PageQuery, NormalizePage, ErrNotFound, ErrConflict
package references

import "context"

// WidgetRepository defines persistence for the widget resource.
// In production code, method signatures use domain/widget and repository.PageQuery.
type WidgetRepository interface {
	List(ctx context.Context, offset, limit int, search string) (items []Widget, total int64, err error)
	GetByID(ctx context.Context, id string) (Widget, error)
	Create(ctx context.Context, value Widget) (Widget, error)
	Update(ctx context.Context, value Widget) (Widget, error)
	Delete(ctx context.Context, id string) error
}

// Relational checklist when adding a real implementation:
//  1. models.go        — widgetModel + TableName()
//  2. schema.go        — append to schemaModels and schemaIndexes
//  3. mapping.go       — toWidgetDomain
//  4. widget_repository.go — GORM CRUD; map driver errors via mapError
