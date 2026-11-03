package web

import (
	"embed"
	"io/fs"
)

//go:embed all:out
var out embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(out, "out")
}
