package web

import (
	"context"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type ApplicationService interface {
	DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error)
	NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error)
	CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error)
	RecordPlan(ctx context.Context, plan *dsl.CompiledPlan, options app.RecordOptions) (*app.RecordSummary, error)
	SetRecordingAudioGain(ctx context.Context, sessionID, sourceID string, gain float64) error
	SetRecordingCompressorEnabled(ctx context.Context, sessionID string, enabled bool) error
}

var _ ApplicationService = (*app.Application)(nil)
