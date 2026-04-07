package router_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"poolx/internal/pkg/sourcefetch"
	"poolx/internal/service"
)

func TestParseSourceConfigURLRoute(t *testing.T) {
	originalFetcher := service.SourceConfigURLFetcherForTest()
	t.Cleanup(func() {
		service.SetSourceConfigURLFetcherForTest(originalFetcher)
	})

	service.SetSourceConfigURLFetcherForTest(func(context.Context, string) (*sourcefetch.FetchResult, error) {
		return &sourcefetch.FetchResult{
			DisplayName: "clash.yaml",
			Content: []byte(`
proxies:
  - name: hk-1
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: secret
`),
			ContentType: "text/yaml",
			FetchedAt:   time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC),
		}, nil
	})

	engine, cookies := loginRootAndBuildEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/source-configs/parse-url", bytes.NewBufferString(`{"url":"https://sub.example.com/clash.yaml"}`))
	req.Header.Set("Content-Type", "application/json")
	for _, cookieValue := range cookies {
		req.AddCookie(cookieValue)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"source_url":"https://sub.example.com/clash.yaml"`) {
		t.Fatalf("expected source_url in response body, got %s", recorder.Body.String())
	}
}

func TestParseSourceConfigURLRouteRejectsMissingURL(t *testing.T) {
	engine, cookies := loginRootAndBuildEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/source-configs/parse-url", bytes.NewBufferString(`{"url":""}`))
	req.Header.Set("Content-Type", "application/json")
	for _, cookieValue := range cookies {
		req.AddCookie(cookieValue)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}

func TestParseSourceConfigURLRouteRejectsInvalidURL(t *testing.T) {
	engine, cookies := loginRootAndBuildEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/source-configs/parse-url", bytes.NewBufferString(`{"url":"ftp://example.com/sub.yaml"}`))
	req.Header.Set("Content-Type", "application/json")
	for _, cookieValue := range cookies {
		req.AddCookie(cookieValue)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}

func TestParseSourceConfigURLRouteRejectsMalformedHostURL(t *testing.T) {
	engine, cookies := loginRootAndBuildEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/source-configs/parse-url", bytes.NewBufferString(`{"url":"http://:80/path"}`))
	req.Header.Set("Content-Type", "application/json")
	for _, cookieValue := range cookies {
		req.AddCookie(cookieValue)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}
