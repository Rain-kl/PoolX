package main

import (
	"embed"
	_ "poolx/docs"
	"poolx/internal/app"
)

//go:embed all:web/build
var buildFS embed.FS

//go:embed web/build/index.html
var indexPage []byte

// @title PoolX Server API
// @version 3.0
// @description PoolX Server API documentation.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Admin API can use Bearer Token.
func main() {
	app.RunServer(buildFS, "web/build", indexPage)
}
