// Illustrative application service for the new-api skill.
//
// Real location: backend/internal/application/widget/service.go (package widget)
// Module: github.com/Rain-kl/Foam/backend
//
// Rules:
//   - First argument is always context.Context
//   - Inject repository interfaces only (no *gin.Context, no GORM)
//   - Map repository.ErrNotFound → package-level ErrNotFound for transport
package references

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("widget not found")
	ErrInvalidInput = errors.New("widget input invalid")
)

type CreateInput struct {
	Name        string
	Description string
}

type UpdateInput struct {
	Name        string
	Description string
}

type ListResult struct {
	Items    []Widget
	Total    int64
	Page     int
	PageSize int
}

// Service orchestrates widget CRUD (application layer).
type Service struct {
	widgets WidgetRepository
}

func NewService(widgets WidgetRepository) *Service {
	return &Service{widgets: widgets}
}

func (s *Service) List(ctx context.Context, page, pageSize int, search string) (ListResult, error) {
	// Real code: page, pageSize = repository.NormalizePage(...)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	items, total, err := s.widgets.List(ctx, offset, pageSize, strings.TrimSpace(search))
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) Get(ctx context.Context, id string) (Widget, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Widget{}, ErrNotFound
	}
	return s.widgets.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Widget, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Widget{}, ErrInvalidInput
	}
	now := time.Now().UTC()
	value := Widget{
		ID:          uuid.NewString(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.widgets.Create(ctx, value)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Widget, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return Widget{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Widget{}, ErrInvalidInput
	}
	current.Name = name
	current.Description = strings.TrimSpace(input.Description)
	current.UpdatedAt = time.Now().UTC()
	return s.widgets.Update(ctx, current)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.widgets.Delete(ctx, id)
}
