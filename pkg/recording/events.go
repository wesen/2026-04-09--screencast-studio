package recording

import "time"

type RunEventType string

const (
	RunEventStateChanged   RunEventType = "state_changed"
	RunEventProcessStarted RunEventType = "process_started"
	RunEventProcessLog     RunEventType = "process_log"
)

type RunEvent struct {
	Type         RunEventType `json:"type"`
	Timestamp    time.Time    `json:"timestamp"`
	State        SessionState `json:"state,omitempty"`
	Reason       string       `json:"reason,omitempty"`
	ProcessLabel string       `json:"process_label,omitempty"`
	OutputPath   string       `json:"output_path,omitempty"`
	Stream       string       `json:"stream,omitempty"`
	Message      string       `json:"message,omitempty"`
}

func emitRunEvent(sink func(RunEvent), event RunEvent) {
	if sink == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	sink(event)
}
