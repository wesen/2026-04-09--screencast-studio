package media

import (
	"context"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type PreviewRuntime interface {
	StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts PreviewOptions) (PreviewSession, error)
}

type PreviewSession interface {
	Wait() error
	Stop(ctx context.Context) error
	LatestFrame() ([]byte, error)
	TakeScreenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)
}

type RecordingRuntime interface {
	StartRecording(ctx context.Context, plan *dsl.CompiledPlan, opts RecordingOptions) (RecordingSession, error)
}

type RecordingSession interface {
	Wait() (*RecordingResult, error)
	Stop(ctx context.Context) error
	SetAudioGain(sourceID string, gain float64) error
	SetAudioCompressorEnabled(enabled bool) error
}

type PreviewOptions struct {
	OnFrame func([]byte)
	OnLog   func(stream, message string)
}

type ScreenshotOptions struct{}

type RecordingOptions struct {
	GracePeriod time.Duration
	MaxDuration time.Duration
	EventSink   func(RecordingEvent)
	Logger      func(string, ...any)
}

type RecordingEventType string

const (
	RecordingEventStateChanged   RecordingEventType = "state_changed"
	RecordingEventProcessStarted RecordingEventType = "process_started"
	RecordingEventProcessLog     RecordingEventType = "process_log"
	RecordingEventAudioLevel     RecordingEventType = "audio_level"
)

type RecordingState string

const (
	RecordingStateStarting RecordingState = "starting"
	RecordingStateRunning  RecordingState = "running"
	RecordingStateStopping RecordingState = "stopping"
	RecordingStateFinished RecordingState = "finished"
	RecordingStateFailed   RecordingState = "failed"
)

type RecordingEvent struct {
	Type         RecordingEventType `json:"type"`
	Timestamp    time.Time          `json:"timestamp"`
	State        RecordingState     `json:"state,omitempty"`
	Reason       string             `json:"reason,omitempty"`
	ProcessLabel string             `json:"process_label,omitempty"`
	OutputPath   string             `json:"output_path,omitempty"`
	Stream       string             `json:"stream,omitempty"`
	Message      string             `json:"message,omitempty"`
	DeviceID     string             `json:"device_id,omitempty"`
	LeftLevel    float64            `json:"left_level,omitempty"`
	RightLevel   float64            `json:"right_level,omitempty"`
	Available    bool               `json:"available,omitempty"`
}

type RecordingResult struct {
	State      RecordingState
	Reason     string
	Outputs    []dsl.PlannedOutput
	StartedAt  time.Time
	FinishedAt time.Time
}
