package httpserver

import (
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	adminauthapp "github.com/Rain-kl/Foam/backend/internal/application/adminauth"
	"github.com/Rain-kl/Foam/backend/internal/transport/http/middleware"
	"github.com/Rain-kl/Foam/backend/zashboard"
	"github.com/gin-gonic/gin"
)

// registerFrontend 在构建产物存在时托管静态文件，并为前端路由提供 SPA 回退。
func registerFrontend(router *gin.Engine, staticPath string, adminAuth *adminauthapp.Service) {
	registerZashboard(router, adminAuth)

	root, indexPath, ok := frontendRoot(staticPath)
	if !ok {
		return
	}
	files := http.FileServer(http.Dir(root))
	router.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) || isBackendPath(requestPath) {
			c.Status(http.StatusNotFound)
			return
		}
		if filePath, exists := frontendFile(root, requestPath); exists {
			if strings.HasPrefix(path.Clean(requestPath), "/assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Header("Cache-Control", "no-cache")
			}
			c.Request.URL.Path = "/" + filepath.ToSlash(filePath)
			files.ServeHTTP(c.Writer, c.Request)
			return
		}
		if path.Ext(path.Clean(requestPath)) != "" {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "no-cache")
		http.ServeFile(c.Writer, c.Request, indexPath)
	})
}

func registerZashboard(router *gin.Engine, adminAuth *adminauthapp.Service) {
	zFS, err := zashboard.FS()
	if err != nil {
		return
	}

	zGroup := router.Group("/zashboard")
	zGroup.Use(middleware.AdminAuth(adminAuth))
	zGroup.GET("", func(c *gin.Context) {
		// Keep query (e.g. token) when normalizing trailing slash.
		target := "/zashboard/"
		if raw := c.Request.URL.RawQuery; raw != "" {
			target += "?" + raw
		} else {
			target += "?" + zashboardDefaultQuery(c)
		}
		c.Redirect(http.StatusTemporaryRedirect, target)
	})
	zGroup.GET("/*any", func(c *gin.Context) {
		// Allow same-origin admin iframe embedding explicitly on static assets too.
		c.Header("X-Frame-Options", "SAMEORIGIN")
		// Bare /zashboard/ without backend params → inject proxy path so zashboard
		// skips setup and uses AdminAuth-protected Clash reverse proxy.
		anyPath := strings.Trim(c.Param("any"), "/")
		if anyPath == "" && c.Request.URL.RawQuery == "" {
			c.Redirect(http.StatusTemporaryRedirect, "/zashboard/?"+zashboardDefaultQuery(c))
			return
		}
		if servePrefixedFS(c, zFS, "/zashboard") {
			return
		}
		c.Status(http.StatusNotFound)
	})
}

// zashboardDefaultQuery builds URL params that point zashboard at the same-origin
// Clash API reverse proxy (/api/zashboard/clash/*). Auth is AdminAuth (cookie/JWT);
// the proxy injects the mihomo secret — no second login for the core API.
func zashboardDefaultQuery(c *gin.Context) string {
	hostname := c.Request.Host
	port := ""
	if h, p, err := net.SplitHostPort(c.Request.Host); err == nil {
		hostname = h
		port = p
	}
	if port == "" {
		proto := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
		switch {
		case proto == "https", c.Request.TLS != nil:
			port = "443"
		default:
			port = "80"
		}
	}
	q := url.Values{}
	q.Set("hostname", hostname)
	q.Set("port", port)
	q.Set("secondaryPath", "/api/zashboard/clash")
	return q.Encode()
}

func servePrefixedFS(c *gin.Context, embedFS fs.FS, prefix string) bool {
	requestPath := strings.TrimPrefix(c.Request.URL.Path, prefix)
	requestPath = strings.TrimPrefix(requestPath, "/")

	candidates := []string{"index.html"}
	if requestPath != "" {
		candidates = append([]string{requestPath}, candidates...)
	}

	for _, candidate := range candidates {
		content, err := fs.ReadFile(embedFS, candidate)
		if err == nil {
			contentType := "text/html; charset=utf-8"
			if ext := path.Ext(candidate); ext != "" && candidate != "index.html" {
				contentType = mime.TypeByExtension(ext)
				if contentType == "" {
					contentType = "application/octet-stream"
				}
			}
			c.Data(http.StatusOK, contentType, content)
			return true
		}
	}
	return false
}

func frontendRoot(staticPath string) (string, string, bool) {
	staticPath = strings.TrimSpace(staticPath)
	if staticPath == "" {
		return "", "", false
	}
	root, err := filepath.Abs(staticPath)
	if err != nil {
		return "", "", false
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "", "", false
	}
	indexPath := filepath.Join(root, "index.html")
	indexInfo, err := os.Stat(indexPath)
	if err != nil || !indexInfo.Mode().IsRegular() {
		return "", "", false
	}
	return filepath.Clean(root), indexPath, true
}

func frontendFile(root, requestPath string) (string, bool) {
	cleanPath := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if cleanPath == "" || cleanPath == "." {
		return "", false
	}
	fullPath := filepath.Join(root, filepath.FromSlash(cleanPath))
	relative, err := filepath.Rel(root, fullPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	info, err := os.Stat(fullPath)
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return relative, true
}

func isBackendPath(value string) bool {
	cleanPath := path.Clean("/" + value)
	for _, prefix := range []string{"/api", "/v1", "/swagger", "/zashboard"} {
		if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+"/") {
			return true
		}
	}
	return cleanPath == "/healthz" || cleanPath == "/readyz"
}
