package relational

import (
	"github.com/Rain-kl/Foam/backend/internal/domain/admin"
	"github.com/Rain-kl/Foam/backend/internal/domain/example"
)

func toAdminDomain(value adminModel) admin.Admin {
	return admin.Admin{
		ID:           value.ID,
		Username:     value.Username,
		PasswordHash: value.PasswordHash,
		CreatedAt:    value.CreatedAt,
		UpdatedAt:    value.UpdatedAt,
	}
}

func toSessionDomain(value adminSessionModel) admin.Session {
	return admin.Session{
		ID:               value.ID,
		AdminID:          value.AdminID,
		RefreshTokenHash: value.RefreshTokenHash,
		ExpiresAt:        value.ExpiresAt,
		LastUsedAt:       value.LastUsedAt,
		CreatedAt:        value.CreatedAt,
	}
}

func toExampleDomain(value exampleModel) example.Example {
	return example.Example{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
	}
}
