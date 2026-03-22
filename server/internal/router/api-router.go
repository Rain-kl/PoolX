package router

import (
	controller "poolx/internal/handler"
	"poolx/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/about", controller.GetAbout)
		apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), controller.GitHubOAuth)
		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), controller.WeChatAuth)
		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.WeChatBind)
		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)

		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
			userRoute.GET("/logout", controller.Logout)

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth(), middleware.NoTokenAuth())
			{
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.POST("/self/update", controller.UpdateSelf)
				selfRoute.POST("/self/delete", controller.DeleteSelf)
				selfRoute.GET("/token", controller.GenerateToken)
			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
			{
				adminRoute.GET("/", controller.GetAllUsers)
				adminRoute.GET("/search", controller.SearchUsers)
				adminRoute.GET("/:id", controller.GetUser)
				adminRoute.POST("/", controller.CreateUser)
				adminRoute.POST("/manage", controller.ManageUser)
				adminRoute.POST("/update", controller.UpdateUser)
				adminRoute.POST("/:id/delete", controller.DeleteUser)
			}
		}
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth(), middleware.NoTokenAuth())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.POST("/update", controller.UpdateOption)
			optionRoute.POST("/geoip/preview", controller.PreviewGeoIP)
		}
		updateRoute := apiRouter.Group("/update")
		updateRoute.Use(middleware.RootAuth(), middleware.NoTokenAuth())
		{
			updateRoute.GET("/latest-release", controller.GetLatestRelease)
			updateRoute.GET("/logs/ws", controller.StreamServerUpgradeLogs)
			updateRoute.POST("/manual-upload", controller.UploadManualServerBinary)
			updateRoute.POST("/manual-upgrade", controller.ConfirmManualServerUpgrade)
			updateRoute.POST("/upgrade", controller.UpgradeServer)
		}
		kernelRoute := apiRouter.Group("/kernel")
		kernelRoute.Use(middleware.RootAuth(), middleware.NoTokenAuth())
		{
			mihomoRoute := kernelRoute.Group("/mihomo")
			{
				mihomoRoute.POST("/inspect", controller.InspectMihomoBinary)
				mihomoRoute.POST("/upload", controller.UploadMihomoBinary)
				mihomoRoute.POST("/download", controller.DownloadMihomoBinary)
			}
		}
		fileRoute := apiRouter.Group("/file")
		fileRoute.Use(middleware.AdminAuth())
		{
			fileRoute.GET("/", controller.GetAllFiles)
			fileRoute.GET("/search", controller.SearchFiles)
			fileRoute.POST("/", middleware.UploadRateLimit(), controller.UploadFile)
			fileRoute.POST("/:id/delete", controller.DeleteFile)
		}
		sourceConfigRoute := apiRouter.Group("/source-configs")
		sourceConfigRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			sourceConfigRoute.POST("/parse", middleware.UploadRateLimit(), controller.ParseSourceConfig)
			sourceConfigRoute.POST("/test", controller.TestSourceConfigNodes)
			sourceConfigRoute.POST("/import", controller.ImportSourceConfig)
		}
		proxyNodeRoute := apiRouter.Group("/proxy-nodes")
		proxyNodeRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			proxyNodeRoute.GET("", controller.GetProxyNodes)
			proxyNodeRoute.GET("/options", controller.GetProxyNodeOptions)
			proxyNodeRoute.POST("/delete", controller.DeleteProxyNodes)
			proxyNodeRoute.POST("/tags", controller.UpdateProxyNodeTags)
			proxyNodeRoute.POST("/test", controller.TestProxyNodes)
			proxyNodeRoute.POST("/:id/status", controller.UpdateProxyNodeStatus)
			proxyNodeRoute.POST("/:id/delete", controller.DeleteProxyNode)
		}
		apiRouter.GET("/capabilities", middleware.AdminAuth(), middleware.NoTokenAuth(), controller.GetKernelCapability)
		portProfileRoute := apiRouter.Group("/port-profiles")
		portProfileRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			portProfileRoute.GET("", controller.GetPortProfiles)
			portProfileRoute.POST("", controller.CreatePortProfile)
			portProfileRoute.POST("/preview", controller.PreviewPortProfile)
			portProfileRoute.GET("/:id", controller.GetPortProfile)
			portProfileRoute.POST("/:id", controller.UpdatePortProfile)
			portProfileRoute.GET("/:id/preview", controller.PreviewSavedPortProfile)
			portProfileRoute.POST("/:id/runtime/save", controller.SaveRuntimeConfig)
			portProfileRoute.POST("/:id/delete", controller.DeletePortProfile)
		}
		templateRoute := apiRouter.Group("/port-profile-templates")
		templateRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			templateRoute.GET("", controller.GetPortProfileTemplates)
			templateRoute.POST("", controller.SavePortProfileTemplate)
			templateRoute.POST("/:id/delete", controller.DeletePortProfileTemplate)
		}
		runtimeRoute := apiRouter.Group("/runtime")
		runtimeRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			runtimeRoute.GET("/status", controller.GetRuntimeStatus)
			runtimeRoute.GET("/logs", controller.GetRuntimeLogs)
			runtimeRoute.POST("/start", controller.StartRuntime)
			runtimeRoute.POST("/stop", controller.StopRuntime)
			runtimeRoute.POST("/reload", controller.ReloadRuntime)
		}
		zashboardRoute := apiRouter.Group("/zashboard")
		zashboardRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			zashboardRoute.Any("/clash/*path", controller.ProxyZashboardClash)
		}
		logRoute := apiRouter.Group("/log")
		logRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
		{
			logRoute.GET("/", controller.GetAppLogs)
			logRoute.POST("/", controller.PushAppLog)
		}
	}
}
