package dsl

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var slugCleanerRe = regexp.MustCompile(`[^a-z0-9]+`)

func defaultDisplay() string {
	if d := strings.TrimSpace(os.Getenv("DISPLAY")); d != "" {
		return d
	}
	return ":0.0"
}

func boolValue(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func safePathSegment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, string(filepath.Separator), "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	if s == "" {
		return "unnamed"
	}
	return s
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugCleanerRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "unnamed"
	}
	return s
}
