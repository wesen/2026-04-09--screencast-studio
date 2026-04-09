package dsl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseAndNormalizeExample(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "jank-prototype", "examples", "example.yaml"))
	require.NoError(t, err)

	cfg, err := ParseAndNormalize(body)
	require.NoError(t, err)

	require.Equal(t, "demo", cfg.SessionID)
	require.Len(t, cfg.VideoSources, 4)
	require.Len(t, cfg.AudioSources, 1)
	require.Equal(t, "audio_mix", cfg.AudioMixTemplate)
}

func TestBuildPlanProducesOutputs(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "jank-prototype", "examples", "example.yaml"))
	require.NoError(t, err)

	cfg, err := ParseAndNormalize(body)
	require.NoError(t, err)

	plan, err := BuildPlan(cfg, time.Date(2026, 4, 9, 14, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	require.Len(t, plan.Outputs, 5)
	require.Equal(t, "demo", plan.SessionID)
	require.Contains(t, plan.Outputs[0].Path, "recordings/demo")
}

func TestParseAndNormalizeRejectsMissingTemplates(t *testing.T) {
	_, err := ParseAndNormalize([]byte("schema: recorder.config/v1\nvideo_sources: []\naudio_sources: []\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "destination_templates")
}
