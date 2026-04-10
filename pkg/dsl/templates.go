package dsl

import (
	sterrors "errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type renderVars struct {
	SessionID  string
	SourceID   string
	SourceName string
	SourceType string
	Ext        string
	Now        time.Time
}

func renderDestination(tmpl string, vars renderVars) (string, error) {
	if strings.TrimSpace(tmpl) == "" {
		return "", errors.New("empty destination template")
	}
	path := strings.NewReplacer(
		"{session_id}", safePathSegment(vars.SessionID),
		"{source_id}", safePathSegment(vars.SourceID),
		"{source_name}", safePathSegment(vars.SourceName),
		"{source_type}", safePathSegment(vars.SourceType),
		"{ext}", vars.Ext,
		"{date}", vars.Now.Format("2006-01-02"),
		"{time}", vars.Now.Format("15-04-05"),
		"{timestamp}", vars.Now.Format("20060102-150405"),
	).Replace(tmpl)
	path = strings.NewReplacer(
		"{date}", vars.Now.Format("2006-01-02"),
		"{time}", vars.Now.Format("15-04-05"),
		"{timestamp}", vars.Now.Format("20060102-150405"),
	).Replace(path)
	path = expandHome(path)
	if strings.TrimSpace(path) == "" {
		return "", errors.New("destination template rendered to empty path")
	}
	if strings.Contains(path, "{index}") {
		return renderIndexedDestination(path)
	}
	return filepath.Clean(path), nil
}

func renderIndexedDestination(pathTemplate string) (string, error) {
	for index := 1; ; index++ {
		candidate := filepath.Clean(strings.ReplaceAll(pathTemplate, "{index}", strconv.Itoa(index)))
		exists, err := pathExists(candidate)
		if err != nil {
			return "", errors.Wrapf(err, "stat destination path %q", candidate)
		}
		if !exists {
			return candidate, nil
		}
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if sterrors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func videoExtension(container string) string {
	switch strings.ToLower(strings.TrimSpace(container)) {
	case "mov":
		return "mov"
	case "mp4":
		return "mp4"
	case "mkv":
		return "mkv"
	case "avi":
		return "avi"
	case "webm":
		return "webm"
	default:
		return "mov"
	}
}

func audioExtension(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "", "pcm_s16le", "wav":
		return "wav"
	case "aac":
		return "m4a"
	case "opus":
		return "ogg"
	case "mp3":
		return "mp3"
	default:
		return "wav"
	}
}
