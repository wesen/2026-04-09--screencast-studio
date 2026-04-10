package dsl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRenderDestination(t *testing.T) {
	path, err := renderDestination("./recordings/{session_id}/{source_name}.{ext}", renderVars{
		SessionID:  "demo",
		SourceName: "Main Display",
		Ext:        "mov",
		Now:        time.Date(2026, 4, 9, 15, 4, 5, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Equal(t, "recordings/demo/Main Display.mov", path)
}

func TestRenderDestinationDateAndIncrementalIndex(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "captures", "2026-04-09", "demo", "Main Display-2026-04-09-1.mov")
	require.NoError(t, os.MkdirAll(filepath.Dir(existing), 0o755))
	require.NoError(t, os.WriteFile(existing, []byte("existing"), 0o644))

	path, err := renderDestination(filepath.Join(root, "captures", "{date}", "{session_id}", "{source_name}-{date}-{index}.{ext}"), renderVars{
		SessionID:  "demo",
		SourceName: "Main Display",
		Ext:        "mov",
		Now:        time.Date(2026, 4, 9, 15, 4, 5, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "captures", "2026-04-09", "demo", "Main Display-2026-04-09-2.mov"), path)
}
