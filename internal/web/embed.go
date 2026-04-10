//go:build embed

package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embeddedFS embed.FS

var publicFS, _ = fs.Sub(embeddedFS, "dist")
