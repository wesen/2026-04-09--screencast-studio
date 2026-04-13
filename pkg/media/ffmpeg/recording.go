package ffmpeg

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
)

type RecordingRuntime struct{}

func NewRecordingRuntime() *RecordingRuntime {
	return &RecordingRuntime{}
}

func (RecordingRuntime) StartRecording(ctx context.Context, plan *dsl.CompiledPlan, opts media.RecordingOptions) (media.RecordingSession, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	session := &recordingSession{
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go func() {
		defer close(session.done)
		result, err := recording.Run(runCtx, plan, recording.RunOptions{
			GracePeriod: opts.GracePeriod,
			MaxDuration: opts.MaxDuration,
			Logger:      opts.Logger,
			EventSink: func(event recording.RunEvent) {
				if opts.EventSink == nil {
					return
				}
				opts.EventSink(media.RecordingEvent{
					Type:         media.RecordingEventType(event.Type),
					Timestamp:    event.Timestamp,
					State:        media.RecordingState(event.State),
					Reason:       event.Reason,
					ProcessLabel: event.ProcessLabel,
					OutputPath:   event.OutputPath,
					Stream:       event.Stream,
					Message:      event.Message,
				})
			},
		})
		session.setResult(result, err)
	}()
	return session, nil
}

type recordingSession struct {
	cancel context.CancelFunc
	done   chan struct{}

	mu     sync.RWMutex
	result *media.RecordingResult
	err    error
}

func (s *recordingSession) Wait() (*media.RecordingResult, error) {
	if s == nil {
		return nil, nil
	}
	if s.done != nil {
		<-s.done
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.result == nil {
		return nil, s.err
	}
	result := *s.result
	result.Outputs = append([]dsl.PlannedOutput(nil), s.result.Outputs...)
	return &result, s.err
}

func (s *recordingSession) Stop(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.cancel != nil {
		s.cancel()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if s.done == nil {
		return nil
	}
	select {
	case <-s.done:
		_, err := s.Wait()
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *recordingSession) SetAudioGain(sourceID string, gain float64) error {
	_ = sourceID
	_ = gain
	return errors.New("ffmpeg recording runtime does not support live audio gain control")
}

func (s *recordingSession) SetAudioCompressorEnabled(enabled bool) error {
	_ = enabled
	return errors.New("ffmpeg recording runtime does not support live audio compressor control")
}

func (s *recordingSession) setResult(result *recording.RunResult, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
	if result == nil {
		s.result = nil
		return
	}
	s.result = &media.RecordingResult{
		State:      media.RecordingState(result.State),
		Reason:     result.Reason,
		Outputs:    append([]dsl.PlannedOutput(nil), result.Outputs...),
		StartedAt:  result.StartedAt,
		FinishedAt: result.FinishedAt,
	}
}
