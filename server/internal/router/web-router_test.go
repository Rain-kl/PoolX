package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ginnexttemplate/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestNormalizeStaticExportDataNavigationRewritesDocumentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(normalizeStaticExportDataNavigation())
	engine.GET("/*any", func(c *gin.Context) {
		c.String(http.StatusOK, c.Request.URL.Path)
	})

	req := httptest.NewRequest(http.MethodGet, "/website.txt", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Dest", "document")

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if body := recorder.Body.String(); body != "/website" {
		t.Fatalf("expected document request to be rewritten to /website, got %q", body)
	}
}

func TestNormalizeStaticExportDataNavigationKeepsDataRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(normalizeStaticExportDataNavigation())
	engine.GET("/*any", func(c *gin.Context) {
		c.String(http.StatusOK, c.Request.URL.Path)
	})

	req := httptest.NewRequest(http.MethodGet, "/website.txt", nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if body := recorder.Body.String(); body != "/website.txt" {
		t.Fatalf("expected data request to keep txt path, got %q", body)
	}
}

func TestCacheHeadersDisableExportedPageCaching(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.Cache())
	engine.GET("/website", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/website", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate" {
		t.Fatalf("unexpected cache-control for page: %q", got)
	}
}

func TestCacheHeadersKeepImmutableStaticAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.Cache())
	engine.GET("/_next/static/app.js", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/_next/static/app.js", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("unexpected cache-control for static asset: %q", got)
	}
}
