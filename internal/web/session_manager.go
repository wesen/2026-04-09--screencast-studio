package web

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

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
	state  recordingSessionState
}

type RecordingManager struct {
	app     ApplicationService
	publish func(ServerEvent)

	mu      sync.RWMutex
	current *managedRecording
}

func NewRecordingManager(application ApplicationService, publish func(ServerEvent)) *RecordingManager {
	if publish == nil {
		publish = func(ServerEvent) {}
	}
	return &RecordingManager{
		app:     application,
		publish: publish,
	}
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
		return recordingSessionState{}, err
	}

	m.mu.Lock()
	if m.current != nil && m.current.state.Active {
		current := cloneRecordingState(m.current.state)
		m.mu.Unlock()
		return current, ErrRecordingAlreadyActive
	}

	runCtx, cancel := context.WithCancel(context.Background())
	current := &managedRecording{
		cancel: cancel,
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

	m.publishState(snapshot)

	group, groupCtx := errgroup.WithContext(runCtx)
	group.Go(func() error {
		summary, recordErr := m.app.RecordPlan(groupCtx, plan, app.RecordOptions{
			GracePeriod: gracePeriod,
			MaxDuration: maxDuration,
			EventSink: func(event recording.RunEvent) {
				m.applyRunEvent(plan.SessionID, event)
			},
		})
		m.finish(plan.SessionID, summary, recordErr)
		return nil
	})

	return snapshot, nil
}

func (m *RecordingManager) Stop() recordingSessionState {
	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	if current == nil || !current.state.Active {
		return m.Current()
	}

	log.Info().Str("session_id", current.state.SessionID).Msg("recording stop requested")
	current.cancel()
	return m.Current()
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
		// Process-start events are currently only logged through the event hub.
	}

	snapshot := cloneRecordingState(m.current.state)
	if event.Type == recording.RunEventProcessLog {
		m.publish(ServerEvent{
			Type:      "session.log",
			Timestamp: event.Timestamp,
			Payload: apiProcessLog{
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

func (m *RecordingManager) finish(sessionID string, summary *app.RecordSummary, recordErr error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == nil || m.current.state.SessionID != sessionID {
		return
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
	m.publishState(snapshot)
}

func (m *RecordingManager) publishState(state recordingSessionState) {
	m.publish(ServerEvent{
		Type:      "session.state",
		Timestamp: time.Now(),
		Payload:   mapRecordingSessionResponse(state),
	})
}

func cloneRecordingState(state recordingSessionState) recordingSessionState {
	clone := state
	clone.Warnings = append([]string(nil), state.Warnings...)
	clone.Outputs = append([]dsl.PlannedOutput(nil), state.Outputs...)
	clone.Logs = append([]processLogEntry(nil), state.Logs...)
	return clone
}
