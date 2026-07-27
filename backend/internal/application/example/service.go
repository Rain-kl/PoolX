package example

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/domain/example"
	"github.com/Rain-kl/Foam/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("示例资源不存在")
	ErrInvalidInput = errors.New("示例资源参数无效")
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
	Items    []example.Example
	Total    int64
	Page     int
	PageSize int
}

// Service 编排示例资源 CRUD。
type Service struct {
	examples repository.ExampleRepository
}

func NewService(examples repository.ExampleRepository) *Service {
	return &Service{examples: examples}
}

func (s *Service) List(ctx context.Context, page, pageSize int, search string) (ListResult, error) {
	page, pageSize = repository.NormalizePage(page, pageSize, repository.DefaultPageSize)
	items, total, err := s.examples.List(ctx, repository.PageQuery{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
		Search: strings.TrimSpace(search),
		Sort:   repository.SortQuery{Field: "created_at", Direction: repository.SortDescending},
	})
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) Get(ctx context.Context, id string) (example.Example, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return example.Example{}, ErrNotFound
	}
	value, err := s.examples.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return example.Example{}, ErrNotFound
		}
		return example.Example{}, err
	}
	return value, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (example.Example, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return example.Example{}, ErrInvalidInput
	}
	now := time.Now().UTC()
	value := example.Example{
		ID:          uuid.NewString(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.examples.Create(ctx, value)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (example.Example, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return example.Example{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return example.Example{}, ErrInvalidInput
	}
	current.Name = name
	current.Description = strings.TrimSpace(input.Description)
	current.UpdatedAt = time.Now().UTC()
	updated, err := s.examples.Update(ctx, current)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return example.Example{}, ErrNotFound
		}
		return example.Example{}, err
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	if err := s.examples.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
