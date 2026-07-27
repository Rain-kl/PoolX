package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Rain-kl/Foam/backend/internal/application/adminauth"
	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

const (
	AdminKey         = "admin"
	AccessCookieName = "foam_access_token"
	AccessCookiePath = "/"
)

// AdminAuth 校验管理员 JWT 令牌 (支持 Bearer 头, URL token 参数, 以及 Cookie)。
func AdminAuth(service *adminauth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			if token := strings.TrimSpace(c.Query("token")); token != "" {
				raw = token
				ok = true
			} else if token := strings.TrimSpace(c.Query("access_token")); token != "" {
				raw = token
				ok = true
			} else if cookie, err := c.Cookie(AccessCookieName); err == nil && strings.TrimSpace(cookie) != "" {
				raw = strings.TrimSpace(cookie)
				ok = true
			}
		}

		if !ok {
			handleUnauthorized(c)
			return
		}

		if service == nil {
			response.Error(c, http.StatusServiceUnavailable, "authRuntimeUnavailable", "管理员认证服务暂不可用")
			c.Abort()
			return
		}

		value, err := service.AuthenticateAccess(c.Request.Context(), raw)
		if err != nil {
			if errors.Is(err, adminauth.ErrRuntimeUnavailable) {
				response.Error(c, http.StatusServiceUnavailable, "authRuntimeUnavailable", "管理员认证服务暂不可用")
				c.Abort()
				return
			}
			handleUnauthorized(c)
			return
		}

		c.Set(AdminKey, value)
		c.Next()
	}
}

func handleUnauthorized(c *gin.Context) {
	// Zashboard is embedded in an admin iframe. Never redirect it to the SPA login
	// page — that lands on the parent app origin and trips X-Frame-Options.
	if isZashboardPath(c.Request.URL.Path) {
		writeZashboardUnauthorized(c)
		return
	}

	if isBrowserNavigation(c.Request) {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	response.Error(c, http.StatusUnauthorized, "adminUnauthorized", "未登录")
	c.Abort()
}

func isZashboardPath(path string) bool {
	return path == "/zashboard" || strings.HasPrefix(path, "/zashboard/")
}

func isIFrameRequest(req *http.Request) bool {
	return strings.EqualFold(req.Header.Get("Sec-Fetch-Dest"), "iframe")
}

func writeZashboardUnauthorized(c *gin.Context) {
	c.Header("X-Frame-Options", "SAMEORIGIN")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><body style="background:#090d16;color:#e2e8f0;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;"><div style="text-align:center;padding:1.5rem;"><h2 style="margin:0 0 .5rem;">请先登录后台</h2><p style="margin:0;opacity:.8;">未获得管理员授权凭证，请登录系统后重试。</p></div></body></html>`))
	c.Abort()
}

func isBrowserNavigation(req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}
	// iframe / embed navigations use Sec-Fetch-Mode: navigate; do not treat them
	// as top-level browser navigations that should redirect to /login.
	if isIFrameRequest(req) {
		return false
	}
	dest := strings.ToLower(req.Header.Get("Sec-Fetch-Dest"))
	if dest == "iframe" || dest == "embed" || dest == "frame" {
		return false
	}
	accept := req.Header.Get("Accept")
	return strings.Contains(accept, "text/html") || req.Header.Get("Sec-Fetch-Mode") == "navigate"
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := parts[1]
	return token, token != ""
}
