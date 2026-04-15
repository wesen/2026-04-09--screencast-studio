package web

import (
	"sync"
	"time"
)

type ServerEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload,omitempty"`
}

type EventHub struct {
	mu          sync.RWMutex
	nextID      int
	subscribers map[int]chan ServerEvent
}

func NewEventHub() *EventHub {
	return &EventHub{
		subscribers: map[int]chan ServerEvent{},
	}
}

func (h *EventHub) Publish(event ServerEvent) {
	start := time.Now()
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	labels := eventMetricLabels(event.Type)
	eventHubEventsPublished.Inc(labels)
	defer func() {
		eventHubPublishNanoseconds.Add(labels, uint64(time.Since(start)))
	}()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.subscribers {
		select {
		case ch <- event:
			eventHubEventsDelivered.Inc(labels)
		default:
			eventHubEventsDropped.Inc(labels)
		}
	}
}

func (h *EventHub) Subscribe(buffer int) (<-chan ServerEvent, func()) {
	if buffer <= 0 {
		buffer = 32
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.nextID
	h.nextID++

	ch := make(chan ServerEvent, buffer)
	h.subscribers[id] = ch
	eventHubSubscribers.Set(nil, int64(len(h.subscribers)))

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if existing, ok := h.subscribers[id]; ok {
			delete(h.subscribers, id)
			eventHubSubscribers.Set(nil, int64(len(h.subscribers)))
			close(existing)
		}
	}

	return ch, unsubscribe
}
