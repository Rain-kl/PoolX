package router

import (
	"ginnexttemplate/internal/middleware"
	"io/fs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetRouter(router *gin.Engine, assetFS fs.FS, buildDir string, indexPage []byte) {
	SetApiRouter(router)
	swaggerRoute := router.Group("/swagger")
	swaggerRoute.Use(middleware.AdminAuth())
	swaggerRoute.GET("/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DocExpansion("list"),
		ginSwagger.PersistAuthorization(true),
		ginSwagger.DefaultModelsExpandDepth(1),
	))
	setWebRouter(router, assetFS, buildDir, indexPage)
}
