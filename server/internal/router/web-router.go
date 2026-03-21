package router

import (
	"ginnexttemplate/internal/handler"
	"ginnexttemplate/internal/middleware"
	"ginnexttemplate/internal/pkg/utils/embedfs"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	pathpkg "path"
	"strings"
)

func setWebRouter(router *gin.Engine, assetFS fs.FS, buildDir string, indexPage []byte) {
	exportedBuildFS, err := fs.Sub(assetFS, buildDir)
	if err != nil {
		panic(err)
	}

	router.Use(middleware.GlobalWebRateLimit())
	fileDownloadRoute := router.Group("/")
	fileDownloadRoute.GET("/upload/:file", middleware.DownloadRateLimit(), controller.DownloadFile)
	router.Use(normalizeStaticExportDataNavigation())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", embedfs.EmbedFolder(assetFS, buildDir)))
	router.NoRoute(func(c *gin.Context) {
		if serveExportedPage(c, exportedBuildFS) {
			return
		}

		if isStaticAssetRequest(c.Request.URL.Path) {
			c.Status(http.StatusNotFound)
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPage)
	})
}

func serveExportedPage(c *gin.Context, buildFS fs.FS) bool {
	requestPath := strings.Trim(c.Request.URL.Path, "/")

	candidates := []string{"index.html"}
	if requestPath != "" {
		candidates = []string{
			requestPath + ".html",
			pathpkg.Join(requestPath, "index.html"),
		}
	}

	for _, candidate := range candidates {
		content, err := fs.ReadFile(buildFS, candidate)
		if err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", content)
			return true
		}
	}

	return false
}

func normalizeStaticExportDataNavigation() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if strings.HasSuffix(requestPath, ".txt") && isDocumentNavigationRequest(c.Request) {
			normalizedPath := strings.TrimSuffix(requestPath, ".txt")
			if normalizedPath == "" {
				normalizedPath = "/"
			}
			c.Request.URL.Path = normalizedPath
		}

		c.Next()
	}
}

func isDocumentNavigationRequest(request *http.Request) bool {
	if request.Header.Get("Sec-Fetch-Mode") == "navigate" || request.Header.Get("Sec-Fetch-Dest") == "document" {
		return true
	}

	return strings.Contains(request.Header.Get("Accept"), "text/html")
}

func isStaticAssetRequest(requestPath string) bool {
	return strings.HasPrefix(requestPath, "/_next/") || pathpkg.Ext(requestPath) != ""
}
