package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	settingsapp "github.com/Rain-kl/Foam/backend/internal/application/settings"
	settingsdomain "github.com/Rain-kl/Foam/backend/internal/domain/settings"
	"github.com/Rain-kl/Foam/backend/internal/infra/config"
	"github.com/Rain-kl/Foam/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type memoryRepo struct {
	value    settingsdomain.Config
	revision uint64
	updated  time.Time
	found    bool
}

func (m *memoryRepo) Get(context.Context) (settingsdomain.Config, time.Time, uint64, bool, error) {
	if !m.found {
		return settingsdomain.Config{}, time.Time{}, 0, false, nil
	}
	return m.value, m.updated, m.revision, true, nil
}

func (m *memoryRepo) Save(_ context.Context, value settingsdomain.Config, expectedRevision uint64) (time.Time, uint64, error) {
	if expectedRevision != m.revision {
		return time.Time{}, 0, repository.ErrConflict
	}
	m.value = value
	m.revision = expectedRevision + 1
	m.updated = time.Now().UTC()
	m.found = true
	return m.updated, m.revision, nil
}

func TestSettingsHTTPGetAndUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	base := config.Config{Frontend: config.FrontendConfig{PublicAPIBaseURL: "http://127.0.0.1:8000"}}
	svc := settingsapp.NewService(base, settingsdomain.Config{}, time.Time{}, 0, &memoryRepo{})
	router := gin.New()
	group := router.Group("/api/v1/admin")
	NewHandler(svc).Register(group)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d body=%s", getRec.Code, getRec.Body.String())
	}
	if !bytes.Contains(getRec.Body.Bytes(), []byte(`"error_msg"`)) {
		t.Fatalf("GET missing envelope: %s", getRec.Body.String())
	}

	body, _ := json.Marshal(map[string]any{
		"revision": 0,
		"config": map[string]any{
			"app":      map[string]any{"display_name": "Demo"},
			"frontend": map[string]any{"public_api_base_url": "https://api.demo.test"},
		},
	})
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(body))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d body=%s", putRec.Code, putRec.Body.String())
	}
	if svc.DisplayName() != "Demo" {
		t.Fatalf("DisplayName = %q", svc.DisplayName())
	}
	if svc.PublicAPIBaseURL() != "https://api.demo.test" {
		t.Fatalf("PublicAPIBaseURL = %q", svc.PublicAPIBaseURL())
	}
}
