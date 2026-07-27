package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBearerTokenAcceptsCaseInsensitiveSchemeAndWhitespace(t *testing.T) {
	token, ok := bearerToken("  bearer\tsecret-token  ")
	if !ok || token != "secret-token" {
		t.Fatalf("token = %q, ok = %v", token, ok)
	}
	for _, value := range []string{"", "Bearer", "Basic token", "Bearer token extra"} {
		if _, ok := bearerToken(value); ok {
			t.Fatalf("header %q unexpectedly accepted", value)
		}
	}
}

func TestAdminAuthNilServiceReturnsServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", AdminAuth(nil), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ErrorMsg string `json:"error_msg"`
		Data     any    `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ErrorMsg == "" {
		t.Fatalf("expected error_msg, body = %s", rec.Body.String())
	}
}

func TestAdminAuthZashboardUnauthorizedDoesNotRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/zashboard/*any", AdminAuth(nil), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, tc := range []struct {
		name   string
		path   string
		header map[string]string
	}{
		{
			name: "html navigation",
			path: "/zashboard/",
			header: map[string]string{
				"Accept":         "text/html",
				"Sec-Fetch-Mode": "navigate",
			},
		},
		{
			name: "iframe navigation",
			path: "/zashboard/",
			header: map[string]string{
				"Accept":         "text/html",
				"Sec-Fetch-Dest": "iframe",
				"Sec-Fetch-Mode": "navigate",
			},
		},
		{
			name: "asset under zashboard",
			path: "/zashboard/assets/index.js",
			header: map[string]string{
				"Accept":         "text/html",
				"Sec-Fetch-Mode": "navigate",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			for key, value := range tc.header {
				req.Header.Set(key, value)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusFound || rec.Header().Get("Location") != "" {
				t.Fatalf("unexpected redirect status=%d location=%q body=%s", rec.Code, rec.Header().Get("Location"), rec.Body.String())
			}
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 body=%s", rec.Code, rec.Body.String())
			}
			if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
				t.Fatalf("X-Frame-Options = %q, want SAMEORIGIN", got)
			}
			if !strings.Contains(rec.Body.String(), "请先登录后台") {
				t.Fatalf("body missing unauthorized html: %s", rec.Body.String())
			}
		})
	}
}

func TestIsBrowserNavigationSkipsIFrame(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/self", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Dest", "iframe")
	if isBrowserNavigation(req) {
		t.Fatal("iframe navigation should not be treated as top-level browser navigation")
	}
}
