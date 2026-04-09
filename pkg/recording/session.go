package recording

import (
	"fmt"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type SessionState string

const (
	StateStarting SessionState = "starting"
	StateRunning  SessionState = "running"
	StateStopping SessionState = "stopping"
	StateFinished SessionState = "finished"
	StateFailed   SessionState = "failed"
)

type sessionEventType string

const (
	eventWorkerExited    sessionEventType = "worker_exited"
	eventCancelRequested sessionEventType = "cancel_requested"
	eventHardTimeout     sessionEventType = "hard_timeout"
	eventStopCompleted   sessionEventType = "stop_completed"
)

type sessionEvent struct {
	Type    sessionEventType
	Process *ManagedProcess
	Err     error
	Reason  string
}

type Session struct {
	plan    *dsl.CompiledPlan
	options RunOptions

	state       SessionState
	reason      string
	startedAt   time.Time
	finishedAt  time.Time
	processes   []*ManagedProcess
	active      map[*ManagedProcess]struct{}
	finalTarget SessionState
	stopStarted bool
}

func newSession(plan *dsl.CompiledPlan, options RunOptions) *Session {
	return &Session{
		plan:        plan,
		options:     options,
		state:       StateStarting,
		active:      map[*ManagedProcess]struct{}{},
		finalTarget: StateFinished,
	}
}

func (s *Session) transition(next SessionState, reason string) error {
	if !isAllowedTransition(s.state, next) {
		return fmt.Errorf("invalid session transition %s -> %s", s.state, next)
	}
	s.state = next
	if reason != "" {
		s.reason = reason
	}
	switch next {
	case StateStarting:
		if s.startedAt.IsZero() {
			s.startedAt = time.Now()
		}
	case StateFinished, StateFailed:
		s.finishedAt = time.Now()
	}
	return nil
}

func isAllowedTransition(from, to SessionState) bool {
	switch from {
	case StateStarting:
		return to == StateRunning || to == StateStopping || to == StateFailed
	case StateRunning:
		return to == StateStopping || to == StateFinished || to == StateFailed
	case StateStopping:
		return to == StateFinished || to == StateFailed
	case StateFinished, StateFailed:
		return to == from
	default:
		return false
	}
}

func (s *Session) beginStopping(reason string, finalTarget SessionState) error {
	if s.state == StateStopping || s.state == StateFinished || s.state == StateFailed {
		return nil
	}
	s.finalTarget = finalTarget
	return s.transition(StateStopping, reason)
}

func (s *Session) markProcessStarted(proc *ManagedProcess) {
	s.processes = append(s.processes, proc)
	s.active[proc] = struct{}{}
}

func (s *Session) markProcessExited(proc *ManagedProcess) {
	delete(s.active, proc)
}

func (s *Session) remainingProcesses() int {
	return len(s.active)
}
