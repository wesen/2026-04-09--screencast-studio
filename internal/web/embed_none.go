//go:build !embed

package web

import (
	"io/fs"
	"os"
	"path/filepath"
)

var publicFS fs.FS = diskPublicFS()

func diskPublicFS() fs.FS {
	repoRoot, err := findRepoRootFromCWD()
	if err != nil {
		return nil
	}

	distDir := filepath.Join(repoRoot, "internal", "web", "dist")
	if _, err := os.Stat(filepath.Join(distDir, "index.html")); err != nil {
		return nil
	}

	return os.DirFS(distDir)
}

func findRepoRootFromCWD() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
