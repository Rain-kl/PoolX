package repository

import (
	"context"

	"github.com/Rain-kl/Foam/backend/internal/domain/example"
)

// ExampleRepository 定义示例资源的持久化能力。
type ExampleRepository interface {
	List(ctx context.Context, query PageQuery) ([]example.Example, int64, error)
	GetByID(ctx context.Context, id string) (example.Example, error)
	Create(ctx context.Context, value example.Example) (example.Example, error)
	Update(ctx context.Context, value example.Example) (example.Example, error)
	Delete(ctx context.Context, id string) error
}
