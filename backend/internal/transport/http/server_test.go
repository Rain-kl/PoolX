package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testDependencies() Dependencies {
	return Dependencies{RequestTimeout: time.Second, MaxBodyBytes: 1024, Logger: slog.Default()}
}

func TestReadinessEndpointReturnsStructuredState(t *testing.T) {
	deps := testDependencies()
	deps.Readiness = func(context.Context) ReadinessSnapshot {
		return ReadinessSnapshot{
			Ready: true, State: "ready", UpdatedAt: time.Now().UTC(),
			Components: map[string]ReadinessComponent{
				"database": {State: "ready"},
			},
		}
	}
	router := New(deps)
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body ReadinessSnapshot
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Ready || body.State != "ready" || body.Components["database"].State != "ready" {
		t.Fatalf("body = %#v", body)
	}
}

func TestReadinessEndpointReturns503WhileNotReady(t *testing.T) {
	deps := testDependencies()
	deps.Readiness = func(context.Context) ReadinessSnapshot {
		return ReadinessSnapshot{Ready: false, State: "starting", UpdatedAt: time.Now().UTC()}
	}
	router := New(deps)
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), `"state":"starting"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSystemEndpointsRequireAdminAuthentication(t *testing.T) {
	deps := testDependencies()
	deps.PublicAPIBaseURL = func() string { return "https://api.example.com" }
	router := New(deps)
	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/admin/status"},
		{method: http.MethodGet, path: "/api/v1/admin/version"},
		{method: http.MethodGet, path: "/api/v1/admin/examples"},
		{method: http.MethodGet, path: "/api/v1/user/self"},
	} {
		request := httptest.NewRequest(route.method, route.path, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want 401 body=%s", route.method, route.path, recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), `"error_msg"`) {
			t.Fatalf("%s %s missing error_msg envelope: %s", route.method, route.path, recorder.Body.String())
		}
	}
}

func TestPublicHealthEndpoints(t *testing.T) {
	router := New(testDependencies())
	for _, path := range []string{"/healthz", "/api/health", "/api/v1/config/public"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", path, recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), `"error_msg"`) && path != "/readyz" {
			// healthz and /api/health use Wavelet envelope
			if path == "/healthz" || path == "/api/health" {
				if !strings.Contains(recorder.Body.String(), `"error_msg"`) {
					t.Fatalf("%s missing envelope: %s", path, recorder.Body.String())
				}
			}
		}
	}
}

func TestFrontendSPAFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>ok</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	deps := testDependencies()
	deps.FrontendStaticPath = dir
	router := New(deps)
	request := httptest.NewRequest(http.MethodGet, "/example/page/dashboard", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "ok") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
