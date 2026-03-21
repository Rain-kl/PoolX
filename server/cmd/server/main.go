package main

import (
	"ginnexttemplate/internal/app"
	"io/fs"
	"os"
)

func main() {
	indexPage, err := fs.ReadFile(os.DirFS("."), "web/build/index.html")
	if err != nil {
		panic(err)
	}
	app.RunServer(os.DirFS("."), "web/build", indexPage)
}
