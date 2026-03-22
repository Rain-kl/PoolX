package middleware

import (
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path

		switch {
		case strings.HasPrefix(requestPath, "/_next/static/"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		case isStaticPublicAsset(requestPath):
			c.Header("Cache-Control", "public, max-age=86400")
		default:
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

func isStaticPublicAsset(requestPath string) bool {
	ext := strings.ToLower(path.Ext(requestPath))
	switch ext {
	case ".ico", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".css", ".js":
		return true
	default:
		return false
	}
}
