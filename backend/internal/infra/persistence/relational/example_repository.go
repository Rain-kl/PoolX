package relational

import (
	"context"
	"strings"

	"github.com/Rain-kl/Foam/backend/internal/domain/example"
	"github.com/Rain-kl/Foam/backend/internal/repository"
)

type ExampleRepository struct{ db *Database }

func NewExampleRepository(db *Database) *ExampleRepository {
	return &ExampleRepository{db: db}
}

func (r *ExampleRepository) List(ctx context.Context, query repository.PageQuery) ([]example.Example, int64, error) {
	tx := r.db.db.WithContext(ctx).Model(&exampleModel{})
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		tx = tx.Where("name LIKE ? OR description LIKE ?", like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, mapError(err)
	}
	order := "created_at DESC, id DESC"
	if query.Sort.Field == "name" {
		if query.Sort.Direction == repository.SortAscending {
			order = "name ASC, id ASC"
		} else {
			order = "name DESC, id DESC"
		}
	} else if query.Sort.Direction == repository.SortAscending {
		order = "created_at ASC, id ASC"
	}
	var rows []exampleModel
	if err := tx.Order(order).Offset(query.Offset).Limit(query.Limit).Find(&rows).Error; err != nil {
		return nil, 0, mapError(err)
	}
	items := make([]example.Example, 0, len(rows))
	for _, row := range rows {
		items = append(items, toExampleDomain(row))
	}
	return items, total, nil
}

func (r *ExampleRepository) GetByID(ctx context.Context, id string) (example.Example, error) {
	var row exampleModel
	if err := r.db.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return example.Example{}, mapError(err)
	}
	return toExampleDomain(row), nil
}

func (r *ExampleRepository) Create(ctx context.Context, value example.Example) (example.Example, error) {
	row := exampleModel{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
		CreatedAt:   value.CreatedAt.UTC(),
		UpdatedAt:   value.UpdatedAt.UTC(),
	}
	if err := r.db.db.WithContext(ctx).Create(&row).Error; err != nil {
		return example.Example{}, mapError(err)
	}
	return toExampleDomain(row), nil
}

func (r *ExampleRepository) Update(ctx context.Context, value example.Example) (example.Example, error) {
	result := r.db.db.WithContext(ctx).Model(&exampleModel{}).Where("id = ?", value.ID).Updates(map[string]any{
		"name":        value.Name,
		"description": value.Description,
		"updated_at":  value.UpdatedAt.UTC(),
	})
	if result.Error != nil {
		return example.Example{}, mapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return example.Example{}, repository.ErrNotFound
	}
	return r.GetByID(ctx, value.ID)
}

func (r *ExampleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.db.WithContext(ctx).Where("id = ?", id).Delete(&exampleModel{})
	if result.Error != nil {
		return mapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
