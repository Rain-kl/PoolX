package main

import (
	"io/fs"
	"os"
	"poolx/internal/app"
)

func main() {
	indexPage, err := fs.ReadFile(os.DirFS("."), "web/build/index.html")
	if err != nil {
		panic(err)
	}
	app.RunServer(os.DirFS("."), "web/build", indexPage)
}
