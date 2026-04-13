package app

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	ffmpegmedia "github.com/wesen/2026-04-09--screencast-studio/pkg/media/ffmpeg"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
)

type Application struct {
	recordingRuntime media.RecordingRuntime
}

type Option func(*Application)

func WithRecordingRuntime(runtime media.RecordingRuntime) Option {
	return func(a *Application) {
		a.recordingRuntime = runtime
	}
}

func New(opts ...Option) *Application {
	a := &Application{
		recordingRuntime: ffmpegmedia.NewRecordingRuntime(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}
	if a.recordingRuntime == nil {
		a.recordingRuntime = ffmpegmedia.NewRecordingRuntime()
	}
	return a
}

type CompileSummary struct {
	File      string
	SessionID string
	Outputs   []dsl.PlannedOutput
	Warnings  []string
}

type RecordOptions struct {
	GracePeriod time.Duration
	MaxDuration time.Duration
	EventSink   func(recording.RunEvent)
}

type RecordSummary struct {
	File       string
	SessionID  string
	State      string
	Reason     string
	Outputs    []dsl.PlannedOutput
	Warnings   []string
	StartedAt  time.Time
	FinishedAt time.Time
}

func (a *Application) DiscoveryList(ctx context.Context, kind string) ([]map[string]any, error) {
	rows := []map[string]any{}
	if kind == "all" || kind == "display" {
		displays, err := discovery.ListDisplays(ctx)
		if err != nil {
			return nil, err
		}
		for _, display := range displays {
			rows = append(rows, map[string]any{
				"kind":      "display",
				"id":        display.ID,
				"name":      display.Name,
				"primary":   display.Primary,
				"x":         display.X,
				"y":         display.Y,
				"width":     display.Width,
				"height":    display.Height,
				"connector": display.Connector,
			})
		}
	}
	if kind == "all" || kind == "window" {
		windows, err := discovery.ListWindows(ctx)
		if err != nil {
			return nil, err
		}
		for _, window := range windows {
			rows = append(rows, map[string]any{
				"kind":   "window",
				"id":     window.ID,
				"title":  window.Title,
				"x":      window.X,
				"y":      window.Y,
				"width":  window.Width,
				"height": window.Height,
			})
		}
	}
	if kind == "all" || kind == "camera" {
		cameras, err := discovery.ListCameras(ctx)
		if err != nil {
			return nil, err
		}
		for _, camera := range cameras {
			rows = append(rows, map[string]any{
				"kind":      "camera",
				"id":        camera.ID,
				"label":     camera.Label,
				"device":    camera.Device,
				"card_name": camera.CardName,
			})
		}
	}
	if kind == "all" || kind == "audio" {
		audioInputs, err := discovery.ListAudioInputs(ctx)
		if err != nil {
			return nil, err
		}
		for _, input := range audioInputs {
			rows = append(rows, map[string]any{
				"kind":        "audio",
				"id":          input.ID,
				"name":        input.Name,
				"driver":      input.Driver,
				"sample_spec": input.SampleSpec,
				"state":       input.State,
			})
		}
	}

	if kind != "all" && kind != "display" && kind != "window" && kind != "camera" && kind != "audio" {
		return nil, fmt.Errorf("unsupported discovery kind %q", kind)
	}

	return rows, nil
}

func (a *Application) DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error) {
	return discovery.SnapshotAll(ctx)
}

func (a *Application) NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error) {
	_ = ctx
	cfg, err := dsl.ParseAndNormalize(body)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (a *Application) CompileFile(ctx context.Context, file string) (*CompileSummary, error) {
	_ = ctx
	plan, err := a.compileFileAt(file, time.Now())
	if err != nil {
		return nil, err
	}

	return &CompileSummary{
		File:      file,
		SessionID: plan.SessionID,
		Outputs:   append([]dsl.PlannedOutput(nil), plan.Outputs...),
		Warnings:  append([]string(nil), plan.Warnings...),
	}, nil
}

func (a *Application) CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error) {
	return a.compileDSLAt(ctx, body, time.Now())
}

func (a *Application) RecordFile(ctx context.Context, file string, options RecordOptions) (*RecordSummary, error) {
	plan, err := a.compileFileAt(file, time.Now())
	if err != nil {
		return nil, err
	}

	summary, err := a.RecordPlan(ctx, plan, options)
	if err != nil {
		return nil, err
	}
	summary.File = file
	return summary, nil
}

func (a *Application) RecordPlan(ctx context.Context, plan *dsl.CompiledPlan, options RecordOptions) (*RecordSummary, error) {
	runtime := a.recordingRuntime
	if runtime == nil {
		runtime = ffmpegmedia.NewRecordingRuntime()
	}

	session, err := runtime.StartRecording(ctx, plan, media.RecordingOptions{
		GracePeriod: options.GracePeriod,
		MaxDuration: options.MaxDuration,
		EventSink: func(event media.RecordingEvent) {
			if options.EventSink == nil {
				return
			}
			options.EventSink(recording.RunEvent{
				Type:         recording.RunEventType(event.Type),
				Timestamp:    event.Timestamp,
				State:        recording.SessionState(event.State),
				Reason:       event.Reason,
				ProcessLabel: event.ProcessLabel,
				OutputPath:   event.OutputPath,
				Stream:       event.Stream,
				Message:      event.Message,
			})
		},
		Logger: func(format string, args ...any) {
			log.Info().Msgf(format, args...)
		},
	})
	summary := &RecordSummary{
		SessionID: plan.SessionID,
		Outputs:   append([]dsl.PlannedOutput(nil), plan.Outputs...),
		Warnings:  append([]string(nil), plan.Warnings...),
	}
	if err != nil {
		return summary, err
	}

	result, err := session.Wait()
	if result != nil {
		summary.State = string(result.State)
		summary.Reason = result.Reason
		summary.StartedAt = result.StartedAt
		summary.FinishedAt = result.FinishedAt
	}
	if err != nil {
		return summary, err
	}
	return summary, nil
}

func (a *Application) compileFileAt(file string, now time.Time) (*dsl.CompiledPlan, error) {
	body, err := dsl.LoadFile(file)
	if err != nil {
		return nil, err
	}

	return a.compileDSLAt(context.Background(), body, now)
}

func (a *Application) compileDSLAt(ctx context.Context, body []byte, now time.Time) (*dsl.CompiledPlan, error) {
	cfg, err := a.NormalizeDSL(ctx, body)
	if err != nil {
		return nil, err
	}

	plan, err := dsl.BuildPlan(cfg, now)
	if err != nil {
		return nil, err
	}
	if len(plan.Outputs) == 0 {
		return nil, fmt.Errorf("compiled plan produced no outputs")
	}
	return plan, nil
}
