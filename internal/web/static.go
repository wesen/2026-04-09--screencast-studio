package web

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

func serveSPAFromFS(w http.ResponseWriter, r *http.Request, root fs.FS) bool {
	if root == nil {
		return false
	}

	requestPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if requestPath == "." || requestPath == "" {
		requestPath = "index.html"
	}

	if ok := serveFSAsset(w, r, root, requestPath); ok {
		return true
	}

	if requestPath != "index.html" {
		return serveFSAsset(w, r, root, "index.html")
	}

	return false
}

func serveFSAsset(w http.ResponseWriter, r *http.Request, root fs.FS, name string) bool {
	data, err := fs.ReadFile(root, name)
	if err != nil {
		return false
	}

	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	return true
}
