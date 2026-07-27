package zashboard

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var embedFS embed.FS

// FS 返回嵌入的 zashboard/dist 文件系统子集。
func FS() (fs.FS, error) {
	return fs.Sub(embedFS, "dist")
}
