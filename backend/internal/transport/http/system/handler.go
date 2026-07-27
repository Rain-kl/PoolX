package system

import (
	"net/http"
	"strings"

	"github.com/Rain-kl/Foam/backend/internal/buildinfo"
	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	publicAPIBaseURL func() string
}

func NewHandler(publicAPIBaseURL func() string) *Handler {
	if publicAPIBaseURL == nil {
		publicAPIBaseURL = func() string { return "" }
	}
	return &Handler{publicAPIBaseURL: publicAPIBaseURL}
}

// Register mounts admin system status under /api/v1/admin.
func (h *Handler) Register(router *gin.RouterGroup) {
	router.GET("/status", h.status)
	router.GET("/version", h.version)
}

func (h *Handler) version(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"current_version": buildinfo.CurrentVersion(),
		"build_time":      buildinfo.CurrentBuildTime(),
		"is_canary":       buildinfo.IsCanary(),
	})
}

func (h *Handler) status(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"public_api_base_url": strings.TrimRight(strings.TrimSpace(h.publicAPIBaseURL()), "/"),
		"current_version":     buildinfo.CurrentVersion(),
		"build_time":          buildinfo.CurrentBuildTime(),
		"is_canary":           buildinfo.IsCanary(),
		"ok":                  true,
	})
}
