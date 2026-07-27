package adminauth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/domain/admin"
	"github.com/Rain-kl/Foam/backend/internal/infra/persistence/relational"
	"github.com/Rain-kl/Foam/backend/internal/infra/security"
	"github.com/Rain-kl/Foam/backend/internal/repository"
)

func TestRefreshTokenRotationAndLogout(t *testing.T) {
	database, err := relational.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.InitializeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	service := NewService(relational.NewAdminRepository(database), relational.NewAdminSessionRepository(database), security.NewTokenService("12345678901234567890123456789012"), time.Minute, time.Hour)
	ctx := context.Background()
	if err := service.Bootstrap(ctx, "admin", "password123"); err != nil {
		t.Fatal(err)
	}
	_, tokens, err := service.Login(ctx, "admin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := service.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Refresh(ctx, tokens.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("旧 refresh token 仍可使用: %v", err)
	}
	if err := service.Logout(ctx, rotated.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AuthenticateAccess(ctx, rotated.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("注销后的 access token 仍可使用: %v", err)
	}
	if _, err := service.Refresh(ctx, rotated.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("注销后的 refresh token 仍可使用: %v", err)
	}
}

func TestChangePasswordRevokesAllSessions(t *testing.T) {
	database, err := relational.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.InitializeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	service := NewService(relational.NewAdminRepository(database), relational.NewAdminSessionRepository(database), security.NewTokenService("12345678901234567890123456789012"), time.Minute, time.Hour)
	ctx := context.Background()
	if err := service.Bootstrap(ctx, "admin", "password123"); err != nil {
		t.Fatal(err)
	}
	adminValue, tokens, err := service.Login(ctx, "admin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ChangePassword(ctx, adminValue.ID, "password123", "password456"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AuthenticateAccess(ctx, tokens.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("修改密码后的 access token 仍可使用: %v", err)
	}
	if _, err := service.Refresh(ctx, tokens.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("修改密码后的 refresh token 仍可使用: %v", err)
	}
	if _, _, err := service.Login(ctx, "admin", "password123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("旧密码仍可登录: %v", err)
	}
	if _, _, err := service.Login(ctx, "admin", "password456"); err != nil {
		t.Fatalf("新密码无法登录: %v", err)
	}
}

func TestLoginDistinguishesPersistenceFailure(t *testing.T) {
	service := NewService(failingAdminRepository{}, nil, security.NewTokenService("12345678901234567890123456789012"), time.Minute, time.Hour)
	if _, _, err := service.Login(context.Background(), "admin", "password123"); !errors.Is(err, ErrRuntimeUnavailable) {
		t.Fatalf("login persistence error = %v", err)
	}
}

type failingAdminRepository struct{}

func (failingAdminRepository) Count(context.Context) (int64, error) {
	return 0, errors.New("db down")
}
func (failingAdminRepository) Create(context.Context, admin.Admin) (admin.Admin, error) {
	return admin.Admin{}, errors.New("db down")
}
func (failingAdminRepository) GetByUsername(context.Context, string) (admin.Admin, error) {
	return admin.Admin{}, errors.New("db down")
}
func (failingAdminRepository) GetByID(context.Context, uint64) (admin.Admin, error) {
	return admin.Admin{}, errors.New("db down")
}
func (failingAdminRepository) UpdatePasswordAndRevokeSessions(context.Context, uint64, string) error {
	return errors.New("db down")
}

var _ repository.AdminRepository = failingAdminRepository{}
