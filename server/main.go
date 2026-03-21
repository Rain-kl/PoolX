package main

import (
	"embed"
	_ "ginnexttemplate/docs"
	"ginnexttemplate/internal/app"
)

//go:embed all:web/build
var buildFS embed.FS

//go:embed web/build/index.html
var indexPage []byte

// @title GinNextTemplate Server API
// @version 3.0
// @description GinNextTemplate Server API documentation.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Admin API can use Bearer Token.
func main() {
	app.RunServer(buildFS, "web/build", indexPage)
}
