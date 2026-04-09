package recording

import "testing"

func TestSessionTransitionMatrix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		from SessionState
		to   SessionState
		ok   bool
	}{
		{name: "starting to running", from: StateStarting, to: StateRunning, ok: true},
		{name: "starting to stopping", from: StateStarting, to: StateStopping, ok: true},
		{name: "starting to failed", from: StateStarting, to: StateFailed, ok: true},
		{name: "starting to finished", from: StateStarting, to: StateFinished, ok: false},
		{name: "running to stopping", from: StateRunning, to: StateStopping, ok: true},
		{name: "running to finished", from: StateRunning, to: StateFinished, ok: true},
		{name: "running to failed", from: StateRunning, to: StateFailed, ok: true},
		{name: "running to starting", from: StateRunning, to: StateStarting, ok: false},
		{name: "stopping to finished", from: StateStopping, to: StateFinished, ok: true},
		{name: "stopping to failed", from: StateStopping, to: StateFailed, ok: true},
		{name: "stopping to running", from: StateStopping, to: StateRunning, ok: false},
		{name: "finished to finished", from: StateFinished, to: StateFinished, ok: true},
		{name: "finished to failed", from: StateFinished, to: StateFailed, ok: false},
		{name: "failed to failed", from: StateFailed, to: StateFailed, ok: true},
		{name: "failed to finished", from: StateFailed, to: StateFinished, ok: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isAllowedTransition(tc.from, tc.to); got != tc.ok {
				t.Fatalf("isAllowedTransition(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.ok)
			}
		})
	}
}

func TestSessionTransitionSetsReasonAndFinishTime(t *testing.T) {
	t.Parallel()

	s := newSession(nil, RunOptions{})
	if err := s.transition(StateRunning, "workers started"); err != nil {
		t.Fatalf("transition to running: %v", err)
	}
	if s.state != StateRunning {
		t.Fatalf("state = %s, want %s", s.state, StateRunning)
	}
	if s.reason != "workers started" {
		t.Fatalf("reason = %q, want %q", s.reason, "workers started")
	}

	if err := s.transition(StateFinished, "done"); err != nil {
		t.Fatalf("transition to finished: %v", err)
	}
	if s.state != StateFinished {
		t.Fatalf("state = %s, want %s", s.state, StateFinished)
	}
	if s.reason != "done" {
		t.Fatalf("reason = %q, want %q", s.reason, "done")
	}
	if s.finishedAt.IsZero() {
		t.Fatal("finishedAt was not set")
	}
}

func TestSessionBeginStoppingPreservesTerminalState(t *testing.T) {
	t.Parallel()

	s := newSession(nil, RunOptions{})
	if err := s.transition(StateRunning, "running"); err != nil {
		t.Fatalf("transition to running: %v", err)
	}
	if err := s.beginStopping("interrupt", StateFinished); err != nil {
		t.Fatalf("beginStopping: %v", err)
	}
	if s.state != StateStopping {
		t.Fatalf("state = %s, want %s", s.state, StateStopping)
	}
	if s.finalTarget != StateFinished {
		t.Fatalf("finalTarget = %s, want %s", s.finalTarget, StateFinished)
	}

	if err := s.transition(StateFinished, "done"); err != nil {
		t.Fatalf("transition to finished: %v", err)
	}
	if err := s.beginStopping("ignored", StateFailed); err != nil {
		t.Fatalf("beginStopping on terminal state: %v", err)
	}
	if s.state != StateFinished {
		t.Fatalf("terminal state changed to %s", s.state)
	}
}
