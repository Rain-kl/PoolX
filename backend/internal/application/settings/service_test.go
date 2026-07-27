package settings

import (
	"context"
	"testing"
	"time"

	settingsdomain "github.com/Rain-kl/Foam/backend/internal/domain/settings"
	"github.com/Rain-kl/Foam/backend/internal/infra/config"
	"github.com/Rain-kl/Foam/backend/internal/repository"
)

type memorySettingsRepo struct {
	value    settingsdomain.Config
	revision uint64
	updated  time.Time
	found    bool
}

func (m *memorySettingsRepo) Get(context.Context) (settingsdomain.Config, time.Time, uint64, bool, error) {
	if !m.found {
		return settingsdomain.Config{}, time.Time{}, 0, false, nil
	}
	return m.value, m.updated, m.revision, true, nil
}

func (m *memorySettingsRepo) Save(_ context.Context, value settingsdomain.Config, expectedRevision uint64) (time.Time, uint64, error) {
	if expectedRevision != m.revision {
		return time.Time{}, 0, repository.ErrConflict
	}
	m.value = value
	m.revision = expectedRevision + 1
	m.updated = time.Now().UTC()
	m.found = true
	return m.updated, m.revision, nil
}

func TestServiceUpdateAndPublicAPIBaseURL(t *testing.T) {
	t.Parallel()
	base := config.Config{
		Frontend: config.FrontendConfig{PublicAPIBaseURL: "http://127.0.0.1:8000"},
	}
	repo := &memorySettingsRepo{}
	svc := NewService(base, settingsdomain.Config{}, time.Time{}, 0, repo)

	if got := svc.PublicAPIBaseURL(); got != "http://127.0.0.1:8000" {
		t.Fatalf("PublicAPIBaseURL() = %q", got)
	}
	if got := svc.DisplayName(); got != "Foam" {
		t.Fatalf("DisplayName() = %q", got)
	}

	snapshot, err := svc.Update(context.Background(), 0, EditableConfig{
		App:      AppConfigInput{DisplayName: "Acme Console"},
		Frontend: FrontendConfigInput{PublicAPIBaseURL: "https://api.example.com/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Revision != 1 {
		t.Fatalf("revision = %d", snapshot.Revision)
	}
	if svc.PublicAPIBaseURL() != "https://api.example.com" {
		t.Fatalf("PublicAPIBaseURL after update = %q", svc.PublicAPIBaseURL())
	}
	if svc.DisplayName() != "Acme Console" {
		t.Fatalf("DisplayName after update = %q", svc.DisplayName())
	}

	if _, err := svc.Update(context.Background(), 0, EditableConfig{}); !errorsIsConflict(err) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestNormalizeRejectsInvalidURL(t *testing.T) {
	t.Parallel()
	base := config.Config{Frontend: config.FrontendConfig{PublicAPIBaseURL: "http://127.0.0.1:8000"}}
	svc := NewService(base, settingsdomain.Config{}, time.Time{}, 0, &memorySettingsRepo{})
	if _, err := svc.Update(context.Background(), 0, EditableConfig{
		Frontend: FrontendConfigInput{PublicAPIBaseURL: "not-a-url"},
	}); err != ErrInvalidInput {
		t.Fatalf("err = %v", err)
	}
}

func errorsIsConflict(err error) bool {
	return err == ErrConflict
}
