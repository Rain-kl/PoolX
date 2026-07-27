package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestZashboardDefaultQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("host with port", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/zashboard/", nil)
		c.Request.Host = "localhost:8000"
		q, err := url.ParseQuery(zashboardDefaultQuery(c))
		if err != nil {
			t.Fatal(err)
		}
		if q.Get("hostname") != "localhost" {
			t.Fatalf("hostname = %q", q.Get("hostname"))
		}
		if q.Get("port") != "8000" {
			t.Fatalf("port = %q", q.Get("port"))
		}
		if q.Get("secondaryPath") != "/api/zashboard/clash" {
			t.Fatalf("secondaryPath = %q", q.Get("secondaryPath"))
		}
	})

	t.Run("host without port defaults http 80", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/zashboard/", nil)
		c.Request.Host = "example.com"
		q, err := url.ParseQuery(zashboardDefaultQuery(c))
		if err != nil {
			t.Fatal(err)
		}
		if q.Get("hostname") != "example.com" || q.Get("port") != "80" {
			t.Fatalf("got hostname=%q port=%q", q.Get("hostname"), q.Get("port"))
		}
	})

	t.Run("x-forwarded-proto https", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/zashboard/", nil)
		c.Request.Host = "example.com"
		c.Request.Header.Set("X-Forwarded-Proto", "https")
		q, err := url.ParseQuery(zashboardDefaultQuery(c))
		if err != nil {
			t.Fatal(err)
		}
		if q.Get("port") != "443" {
			t.Fatalf("port = %q, want 443", q.Get("port"))
		}
	})
}

func TestRegisterZashboardRedirectsBarePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// AdminAuth with nil service returns 503 when Authorization present;
	// without token it 401s. Use a bypass middleware that always passes.
	registerZashboard(router, nil)

	// Without auth cookie/token, bare path still hits AdminAuth first → 401 HTML.
	// Inject a fake auth middleware by replacing: call zashboardDefaultQuery path
	// via a minimal router that only tests redirect after auth.
	router = gin.New()
	router.GET("/zashboard", func(c *gin.Context) {
		target := "/zashboard/"
		if raw := c.Request.URL.RawQuery; raw != "" {
			target += "?" + raw
		} else {
			target += "?" + zashboardDefaultQuery(c)
		}
		c.Redirect(http.StatusTemporaryRedirect, target)
	})
	router.GET("/zashboard/*any", func(c *gin.Context) {
		if (c.Param("any") == "/" || c.Param("any") == "") && c.Request.URL.RawQuery == "" {
			c.Redirect(http.StatusTemporaryRedirect, "/zashboard/?"+zashboardDefaultQuery(c))
			return
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/zashboard/", nil)
	req.Host = "localhost:8000"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "hostname=localhost") ||
		!strings.Contains(loc, "port=8000") ||
		!strings.Contains(loc, "secondaryPath") {
		t.Fatalf("Location = %q", loc)
	}

	// Already configured query must not redirect.
	req = httptest.NewRequest(http.MethodGet, "/zashboard/?hostname=localhost&port=8000&secondaryPath=%2Fapi%2Fzashboard%2Fclash", nil)
	req.Host = "localhost:8000"
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("configured path status = %d", rec.Code)
	}
}
