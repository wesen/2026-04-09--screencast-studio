package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/pkg/errors"
)

const defaultPNPMVersion = "10.15.0"

type packageJSON struct {
	PackageManager string `json:"packageManager"`
}

func main() {
	ctx := context.Background()
	if err := buildAndExportFrontend(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildAndExportFrontend(ctx context.Context) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	uiDir := filepath.Join(repoRoot, "ui")
	outDir := filepath.Join(repoRoot, "internal", "web", "dist")
	if err := os.RemoveAll(outDir); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "remove old web dist")
	}

	pnpmVersion, err := getPnpmVersionFromPackageJSON(uiDir)
	if err != nil {
		pnpmVersion = defaultPNPMVersion
	}

	if err := buildWithDagger(ctx, uiDir, outDir, pnpmVersion); err != nil {
		if err := buildLocal(uiDir, outDir); err != nil {
			return errors.Wrapf(err, "local web build after dagger failure (%v)", err)
		}
	}

	return nil
}

func buildWithDagger(ctx context.Context, uiDir, outDir, pnpmVersion string) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return errors.Wrap(err, "connect dagger")
	}
	defer func() { _ = client.Close() }()

	source := client.Host().Directory(uiDir, dagger.HostDirectoryOpts{
		Exclude: []string{
			"dist",
			"node_modules",
			"storybook-static",
			"tsconfig.tsbuildinfo",
		},
	})

	pnpmStore := client.CacheVolume("screencast-studio-ui-pnpm-store")
	pathValue := "/pnpm:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	container := client.Container().
		From("node:22-bookworm").
		WithEnvVariable("PNPM_HOME", "/pnpm").
		WithEnvVariable("PATH", pathValue).
		WithMountedCache("/pnpm/store", pnpmStore).
		WithDirectory("/src/ui", source).
		WithWorkdir("/src/ui").
		WithExec([]string{"sh", "-lc", "corepack enable && corepack prepare pnpm@" + pnpmVersion + " --activate"}).
		WithExec([]string{"pnpm", "--version"}).
		WithExec([]string{"pnpm", "install", "--prefer-offline", "--frozen-lockfile"}).
		WithExec([]string{"pnpm", "run", "build"})

	if _, err := container.Directory("/src/ui/dist").Export(ctx, outDir); err != nil {
		return errors.Wrap(err, "export built frontend")
	}

	return nil
}

func buildLocal(uiDir, outDir string) error {
	if _, err := exec.LookPath("pnpm"); err != nil {
		return errors.Wrap(err, "pnpm not found in PATH")
	}

	for _, args := range [][]string{
		{"install", "--prefer-offline", "--frozen-lockfile"},
		{"run", "build"},
	} {
		cmd := exec.Command("pnpm", args...)
		cmd.Dir = uiDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "pnpm %s", strings.Join(args, " "))
		}
	}

	return copyDir(filepath.Join(uiDir, "dist"), outDir)
}

func getPnpmVersionFromPackageJSON(uiDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(uiDir, "package.json"))
	if err != nil {
		return "", errors.Wrap(err, "read package.json")
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", errors.Wrap(err, "decode package.json")
	}

	value := strings.TrimSpace(pkg.PackageManager)
	if !strings.HasPrefix(value, "pnpm@") {
		return "", errors.New("packageManager field missing pnpm version")
	}

	version := strings.TrimSpace(strings.TrimPrefix(value, "pnpm@"))
	if version == "" {
		return "", errors.New("empty pnpm version in packageManager")
	}

	return version, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "get working directory")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}
