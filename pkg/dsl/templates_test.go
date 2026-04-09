package dsl

import (
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
