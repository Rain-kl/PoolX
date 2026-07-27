package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	adminauthapp "github.com/Rain-kl/Foam/backend/internal/application/adminauth"
	clashapp "github.com/Rain-kl/Foam/backend/internal/application/clash"
	exampleapp "github.com/Rain-kl/Foam/backend/internal/application/example"
	settingsapp "github.com/Rain-kl/Foam/backend/internal/application/settings"
	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	adminauthhttp "github.com/Rain-kl/Foam/backend/internal/transport/http/adminauth"
	clashhttp "github.com/Rain-kl/Foam/backend/internal/transport/http/clash"
	examplehttp "github.com/Rain-kl/Foam/backend/internal/transport/http/example"
	"github.com/Rain-kl/Foam/backend/internal/transport/http/middleware"
	settingshttp "github.com/Rain-kl/Foam/backend/internal/transport/http/settings"
	systemhttp "github.com/Rain-kl/Foam/backend/internal/transport/http/system"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Logger         *slog.Logger
	RequestTimeout time.Duration
	MaxBodyBytes   int64
	SecureCookies  bool
	// PublicAPIBaseURL 返回当前生效的公开 API 根地址（可被运行设置覆盖）。
	PublicAPIBaseURL   func() string
	FrontendStaticPath string
	// Readiness 返回可观测的就绪状态。Ready 仅为旧调用方保留。
	Readiness func(context.Context) ReadinessSnapshot
	Ready     func(context.Context) bool
	AdminAuth *adminauthapp.Service
	Examples  *exampleapp.Service
	Settings  *settingsapp.Service
	Clash     *clashapp.Service
}

type ReadinessComponent struct {
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

// ReadinessSnapshot 表示公开就绪端点的稳定响应契约。
type ReadinessSnapshot struct {
	Ready      bool                          `json:"ready"`
	State      string                        `json:"state"`
	UpdatedAt  time.Time                     `json:"updated_at"`
	Components map[string]ReadinessComponent `json:"components,omitempty"`
}

// New 创建 HTTP 路由（路径与响应 envelope 对齐 Wavelet）：
//
//	GET  /api/health
//	POST /api/v1/user/login|refresh
//	GET  /api/v1/user/logout|/self
//	POST /api/v1/user/change-password
//	GET  /api/v1/user-info
//	GET  /api/v1/config/public
//	*    /api/v1/admin/*   (status, version, settings, examples)
//	GET  /healthz /readyz  (probe aliases)
func New(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	router := gin.New()
	router.Use(
		gin.Recovery(),
		response.ErrorHandlerMiddleware(),
		middleware.RequestID(),
		middleware.SecurityHeaders(),
		middleware.MaxBodyBytes(deps.MaxBodyBytes),
		middleware.Timeout(deps.RequestTimeout),
		middleware.AccessLog(deps.Logger),
	)

	// Probe aliases (Docker / k8s).
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.OKNil())
	})
	router.GET("/readyz", func(c *gin.Context) {
		const stateReady = "ready"
		if deps.Readiness != nil {
			snapshot := deps.Readiness(c.Request.Context())
			status := http.StatusServiceUnavailable
			if snapshot.Ready {
				status = http.StatusOK
			}
			c.JSON(status, snapshot)
			return
		}
		if deps.Ready != nil && deps.Ready(c.Request.Context()) {
			c.JSON(http.StatusOK, gin.H{"ready": true, "state": stateReady})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true, "state": stateReady})
	})

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.OKNil())
	})

	v1 := api.Group("/v1")
	authHandler := adminauthhttp.NewHandler(deps.AdminAuth, deps.SecureCookies)

	userPublic := v1.Group("/user")
	authHandler.RegisterPublic(userPublic)

	userAuth := v1.Group("/user")
	userAuth.Use(middleware.AdminAuth(deps.AdminAuth))
	authHandler.RegisterAuthenticated(userAuth)

	userInfo := v1.Group("")
	userInfo.Use(middleware.AdminAuth(deps.AdminAuth))
	authHandler.RegisterUserInfoAliases(userInfo)

	v1.GET("/config/public", func(c *gin.Context) {
		payload := gin.H{}
		if deps.Settings != nil {
			payload["display_name"] = deps.Settings.DisplayName()
			payload["public_api_base_url"] = deps.Settings.PublicAPIBaseURL()
		} else if deps.PublicAPIBaseURL != nil {
			payload["public_api_base_url"] = deps.PublicAPIBaseURL()
		}
		response.Success(c, http.StatusOK, payload)
	})

	admin := v1.Group("/admin")
	admin.Use(middleware.AdminAuth(deps.AdminAuth))

	publicAPIBaseURL := deps.PublicAPIBaseURL
	if publicAPIBaseURL == nil {
		publicAPIBaseURL = func() string { return "" }
	}
	systemhttp.NewHandler(publicAPIBaseURL).Register(admin)
	if deps.Settings != nil {
		settingshttp.NewHandler(deps.Settings).Register(admin)
	}
	if deps.Examples != nil {
		examplehttp.NewHandler(deps.Examples).Register(admin)
	} else {
		// Keep path reserved so unauthenticated probes get 401 instead of SPA/404.
		admin.Any("/examples", func(*gin.Context) {})
		admin.Any("/examples/*path", func(*gin.Context) {})
	}

	if deps.Clash != nil {
		clashHandler := clashhttp.NewHandler(deps.Clash)
		clashHandler.Register(admin)

		// Zashboard Clash API Reverse Proxy Endpoints (Protected by AdminAuth)
		zDashboardGroup := api.Group("/zashboard")
		zDashboardGroup.Use(middleware.AdminAuth(deps.AdminAuth))
		zDashboardGroup.Any("/clash/*path", clashHandler.ProxyZashboardClash)

		v1ZDashboardGroup := v1.Group("/zashboard")
		v1ZDashboardGroup.Use(middleware.AdminAuth(deps.AdminAuth))
		v1ZDashboardGroup.Any("/clash/*path", clashHandler.ProxyZashboardClash)
	}

	registerFrontend(router, deps.FrontendStaticPath, deps.AdminAuth)
	return router
}
