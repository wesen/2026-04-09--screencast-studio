package web

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
)

var ErrRecordingAlreadyActive = errors.New("recording already active")

type recordingSessionState struct {
	Active     bool
	SessionID  string
	State      string
	Reason     string
	StartedAt  time.Time
	FinishedAt time.Time
	Warnings   []string
	Outputs    []dsl.PlannedOutput
	Logs       []processLogEntry
}

type processLogEntry struct {
	Timestamp    time.Time
	ProcessLabel string
	Stream       string
	Message      string
}

type managedRecording struct {
	cancel context.CancelFunc
	done   chan struct{}
	state  recordingSessionState
}

type RecordingManager struct {
	app       ApplicationService
	publish   func(ServerEvent)
	parentCtx context.Context

	mu      sync.RWMutex
	current *managedRecording
}

func NewRecordingManager(parentCtx context.Context, application ApplicationService, publish func(ServerEvent)) *RecordingManager {
	if publish == nil {
		publish = func(ServerEvent) {}
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	return &RecordingManager{
		app:       application,
		publish:   publish,
		parentCtx: parentCtx,
	}
}

func (m *RecordingManager) parentContext() context.Context {
	if m.parentCtx == nil {
		return context.Background()
	}
	return m.parentCtx
}

func (m *RecordingManager) Current() recordingSessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return recordingSessionState{}
	}
	return cloneRecordingState(m.current.state)
}

func (m *RecordingManager) Start(dslBody []byte, gracePeriod, maxDuration time.Duration) (recordingSessionState, error) {
	plan, err := m.app.CompileDSL(context.Background(), dslBody)
	if err != nil {
		log.Error().
			Str("event", "recording.session.compile.error").
			Err(err).
			Msg("failed to compile recording plan")
		return recordingSessionState{}, err
	}
	if plan != nil && len(plan.AudioJobs) > 0 {
		audioJobCount := len(plan.AudioJobs)
		audioOutputCount := 0
		filteredOutputs := make([]dsl.PlannedOutput, 0, len(plan.Outputs))
		for _, output := range plan.Outputs {
			if output.Kind == "audio" {
				audioOutputCount++
				continue
			}
			filteredOutputs = append(filteredOutputs, output)
		}
		plan.AudioJobs = nil
		plan.Outputs = filteredOutputs
		log.Warn().
			Str("event", "debug.recording.audio_jobs.disabled").
			Str("session_id", plan.SessionID).
			Int("audio_job_count", audioJobCount).
			Int("audio_output_count", audioOutputCount).
			Int("remaining_video_jobs", len(plan.VideoJobs)).
			Int("remaining_outputs", len(plan.Outputs)).
			Msg("temporarily disabling audio jobs for elimination testing")
	}

	m.mu.Lock()
	if m.current != nil && m.current.state.Active {
		current := cloneRecordingState(m.current.state)
		m.mu.Unlock()
		log.Warn().
			Str("event", "recording.session.start.rejected").
			Str("session_id", current.SessionID).
			Str("state", current.State).
			Msg("recording start rejected because another session is active")
		return current, ErrRecordingAlreadyActive
	}

	runCtx, cancel := context.WithCancel(m.parentContext())
	current := &managedRecording{
		cancel: cancel,
		done:   make(chan struct{}),
		state: recordingSessionState{
			Active:    true,
			SessionID: plan.SessionID,
			State:     string(recording.StateStarting),
			StartedAt: time.Now(),
			Warnings:  append([]string(nil), plan.Warnings...),
			Outputs:   append([]dsl.PlannedOutput(nil), plan.Outputs...),
			Logs:      []processLogEntry{},
		},
	}
	m.current = current
	snapshot := cloneRecordingState(current.state)
	m.mu.Unlock()

	log.Info().
		Str("event", "recording.session.start").
		Str("session_id", plan.SessionID).
		Int("output_count", len(plan.Outputs)).
		Int("warning_count", len(plan.Warnings)).
		Dur("grace_period", gracePeriod).
		Dur("max_duration", maxDuration).
		Msg("recording session starting")

	m.publishState(snapshot)

	group, groupCtx := errgroup.WithContext(runCtx)
	group.Go(func() error {
		defer close(current.done)
		log.Info().
			Str("event", "recording.session.run.begin").
			Str("session_id", plan.SessionID).
			Msg("recording session worker running")
		summary, recordErr := m.app.RecordPlan(groupCtx, plan, app.RecordOptions{
			GracePeriod: gracePeriod,
			MaxDuration: maxDuration,
			EventSink: func(event recording.RunEvent) {
				m.applyRunEvent(plan.SessionID, event)
			},
		})
		_, _ = m.finish(plan.SessionID, summary, recordErr)
		return nil
	})

	return snapshot, nil
}

func (m *RecordingManager) Stop() recordingSessionState {
	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	if current == nil || !current.state.Active {
		snapshot := m.Current()
		log.Info().
			Str("event", "recording.session.stop.noop").
			Str("session_id", snapshot.SessionID).
			Str("state", snapshot.State).
			Msg("recording stop requested but no active session exists")
		return snapshot
	}

	log.Info().
		Str("event", "recording.session.stop.requested").
		Str("session_id", current.state.SessionID).
		Str("state", current.state.State).
		Msg("recording stop requested")
	current.cancel()
	return m.Current()
}

func (m *RecordingManager) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	m.mu.RLock()
	current := m.current
	var snapshot recordingSessionState
	if current != nil {
		snapshot = cloneRecordingState(current.state)
	}
	m.mu.RUnlock()

	if current == nil {
		log.Info().
			Str("event", "recording.shutdown.noop").
			Msg("recording manager shutdown requested with no session")
		return nil
	}

	log.Info().
		Str("event", "recording.shutdown.begin").
		Str("session_id", snapshot.SessionID).
		Str("state", snapshot.State).
		Bool("active", snapshot.Active).
		Msg("recording manager shutdown starting")

	if snapshot.Active {
		current.cancel()
		log.Info().
			Str("event", "recording.shutdown.cancel").
			Str("session_id", snapshot.SessionID).
			Msg("recording manager requested session cancellation")
	}

	if current.done == nil {
		log.Info().
			Str("event", "recording.shutdown.done").
			Str("session_id", snapshot.SessionID).
			Msg("recording manager shutdown finished without wait")
		return nil
	}

	select {
	case <-current.done:
		log.Info().
			Str("event", "recording.shutdown.done").
			Str("session_id", snapshot.SessionID).
			Msg("recording manager shutdown finished")
		return nil
	case <-ctx.Done():
		log.Error().
			Str("event", "recording.shutdown.timeout").
			Str("session_id", snapshot.SessionID).
			Err(ctx.Err()).
			Msg("recording manager shutdown timed out")
		return ctx.Err()
	}
}

func (m *RecordingManager) applyRunEvent(sessionID string, event recording.RunEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == nil || m.current.state.SessionID != sessionID {
		return
	}

	switch event.Type {
	case recording.RunEventStateChanged:
		m.current.state.State = string(event.State)
		m.current.state.Reason = event.Reason
		if event.State == recording.StateFinished || event.State == recording.StateFailed {
			m.current.state.Active = false
			m.current.state.FinishedAt = event.Timestamp
		}
	case recording.RunEventProcessLog:
		m.current.state.Logs = append(m.current.state.Logs, processLogEntry{
			Timestamp:    event.Timestamp,
			ProcessLabel: event.ProcessLabel,
			Stream:       event.Stream,
			Message:      event.Message,
		})
		if len(m.current.state.Logs) > 200 {
			m.current.state.Logs = append([]processLogEntry(nil), m.current.state.Logs[len(m.current.state.Logs)-200:]...)
		}
	case recording.RunEventProcessStarted:
		log.Info().
			Str("event", "recording.process.started").
			Str("session_id", sessionID).
			Str("process_label", event.ProcessLabel).
			Str("output_path", event.OutputPath).
			Msg("recording subprocess started event received")
	case recording.RunEventAudioLevel:
		m.publish(ServerEvent{
			Type:      "telemetry.audio_meter",
			Timestamp: event.Timestamp,
			Payload: &studiov1.AudioMeterEvent{
				DeviceId:   event.DeviceID,
				LeftLevel:  event.LeftLevel,
				RightLevel: event.RightLevel,
				Available:  event.Available,
			},
		})
		return
	}

	snapshot := cloneRecordingState(m.current.state)
	if event.Type == recording.RunEventStateChanged {
		log.Info().
			Str("event", "recording.session.state").
			Str("session_id", sessionID).
			Str("state", m.current.state.State).
			Str("reason", m.current.state.Reason).
			Bool("active", m.current.state.Active).
			Msg("recording session state updated")
	}
	if event.Type == recording.RunEventProcessLog {
		m.publish(ServerEvent{
			Type:      "session.log",
			Timestamp: event.Timestamp,
			Payload: &studiov1.ProcessLog{
				Timestamp:    formatTimestamp(event.Timestamp),
				ProcessLabel: event.ProcessLabel,
				Stream:       event.Stream,
				Message:      event.Message,
			},
		})
		return
	}
	m.publishState(snapshot)
}

func (m *RecordingManager) finish(sessionID string, summary *app.RecordSummary, recordErr error) (recordingSessionState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == nil || m.current.state.SessionID != sessionID {
		return recordingSessionState{}, false
	}

	if summary != nil {
		if summary.State != "" {
			m.current.state.State = summary.State
		}
		m.current.state.Reason = summary.Reason
		m.current.state.Outputs = append([]dsl.PlannedOutput(nil), summary.Outputs...)
		m.current.state.Warnings = append([]string(nil), summary.Warnings...)
		m.current.state.StartedAt = summary.StartedAt
		m.current.state.FinishedAt = summary.FinishedAt
	}
	if recordErr != nil && m.current.state.Reason == "" {
		m.current.state.Reason = recordErr.Error()
	}
	m.current.state.Active = false

	snapshot := cloneRecordingState(m.current.state)
	if recordErr != nil {
		log.Error().
			Str("event", "recording.session.finish").
			Str("session_id", sessionID).
			Str("state", snapshot.State).
			Str("reason", snapshot.Reason).
			Err(recordErr).
			Msg("recording session finished with error")
	} else {
		log.Info().
			Str("event", "recording.session.finish").
			Str("session_id", sessionID).
			Str("state", snapshot.State).
			Str("reason", snapshot.Reason).
			Time("started_at", snapshot.StartedAt).
			Time("finished_at", snapshot.FinishedAt).
			Msg("recording session finished")
	}
	m.publishState(snapshot)
	return snapshot, true
}

func (m *RecordingManager) publishState(state recordingSessionState) {
	m.publish(ServerEvent{
		Type:      "session.state",
		Timestamp: time.Now(),
		Payload:   mapRecordingSession(state),
	})
}

func cloneRecordingState(state recordingSessionState) recordingSessionState {
	clone := state
	clone.Warnings = append([]string(nil), state.Warnings...)
	clone.Outputs = append([]dsl.PlannedOutput(nil), state.Outputs...)
	clone.Logs = append([]processLogEntry(nil), state.Logs...)
	return clone
}
