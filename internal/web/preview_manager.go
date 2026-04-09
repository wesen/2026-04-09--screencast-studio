package web

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

var (
	ErrPreviewNotFound       = errors.New("preview not found")
	ErrPreviewLimitExceeded  = errors.New("preview limit exceeded")
	ErrPreviewSourceNotFound = errors.New("preview source not found")
)

type previewSnapshot struct {
	ID          string
	SourceID    string
	Name        string
	SourceType  string
	State       string
	Reason      string
	Leases      int
	HasFrame    bool
	LastFrameAt time.Time
}

type managedPreview struct {
	id        string
	signature string
	source    dsl.EffectiveVideoSource
	cancel    context.CancelFunc
	done      chan struct{}

	state       string
	reason      string
	leases      int
	latestFrame []byte
	lastFrameAt time.Time
	frameSeq    uint64
}

type PreviewManager struct {
	app     ApplicationService
	publish func(ServerEvent)
	limit   int
	runner  PreviewRunner

	mu          sync.RWMutex
	byID        map[string]*managedPreview
	bySignature map[string]*managedPreview
}

func NewPreviewManager(application ApplicationService, publish func(ServerEvent), limit int, runner PreviewRunner) *PreviewManager {
	if publish == nil {
		publish = func(ServerEvent) {}
	}
	if limit <= 0 {
		limit = 4
	}
	if runner == nil {
		runner = FFmpegPreviewRunner{}
	}
	return &PreviewManager{
		app:         application,
		publish:     publish,
		limit:       limit,
		runner:      runner,
		byID:        map[string]*managedPreview{},
		bySignature: map[string]*managedPreview{},
	}
}

func (m *PreviewManager) Ensure(ctx context.Context, dslBody []byte, sourceID string) (previewSnapshot, error) {
	cfg, err := m.app.NormalizeDSL(ctx, dslBody)
	if err != nil {
		return previewSnapshot{}, err
	}

	source, err := findPreviewSource(cfg, sourceID)
	if err != nil {
		return previewSnapshot{}, err
	}
	signature := computePreviewSignature(source)

	m.mu.Lock()
	if existing, ok := m.bySignature[signature]; ok && existing.state != "failed" {
		existing.leases++
		snapshot := snapshotPreview(existing)
		m.mu.Unlock()
		m.publishPreviewState(snapshot)
		return snapshot, nil
	}
	if len(m.byID) >= m.limit {
		m.mu.Unlock()
		return previewSnapshot{}, ErrPreviewLimitExceeded
	}

	previewCtx, cancel := context.WithCancel(context.Background())
	preview := &managedPreview{
		id:        "preview-" + signature[:12],
		signature: signature,
		source:    source,
		cancel:    cancel,
		done:      make(chan struct{}),
		state:     "starting",
		leases:    1,
	}
	m.byID[preview.id] = preview
	m.bySignature[signature] = preview
	snapshot := snapshotPreview(preview)
	m.mu.Unlock()

	m.publishPreviewState(snapshot)

	group, groupCtx := errgroup.WithContext(previewCtx)
	group.Go(func() error {
		defer close(preview.done)
		err := m.runner.Run(groupCtx, source, func(frame []byte) {
			m.storePreviewFrame(preview.id, frame)
		}, func(stream, line string) {
			m.publish(ServerEvent{
				Type:      "preview.log",
				Timestamp: time.Now(),
				Payload: apiProcessLog{
					Timestamp:    formatTimestamp(time.Now()),
					ProcessLabel: preview.id,
					Stream:       stream,
					Message:      line,
				},
			})
		})
		m.finishPreview(preview.id, err)
		return nil
	})

	return snapshot, nil
}

func (m *PreviewManager) Release(previewID string) (previewSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	preview, ok := m.byID[previewID]
	if !ok {
		return previewSnapshot{}, ErrPreviewNotFound
	}

	if preview.leases > 0 {
		preview.leases--
	}
	if preview.leases == 0 {
		preview.state = "stopping"
		preview.reason = "released"
		preview.cancel()
	}

	snapshot := snapshotPreview(preview)
	m.publishPreviewState(snapshot)
	return snapshot, nil
}

func (m *PreviewManager) List() []previewSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	previews := make([]previewSnapshot, 0, len(m.byID))
	for _, preview := range m.byID {
		previews = append(previews, snapshotPreview(preview))
	}
	return previews
}

func (m *PreviewManager) Lookup(previewID string) (*managedPreview, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	preview, ok := m.byID[previewID]
	return preview, ok
}

func (m *PreviewManager) Snapshot(previewID string) (previewSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	preview, ok := m.byID[previewID]
	if !ok {
		return previewSnapshot{}, false
	}
	return snapshotPreview(preview), true
}

func (m *PreviewManager) LatestFrame(previewID string) ([]byte, uint64, previewSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	preview, ok := m.byID[previewID]
	if !ok {
		return nil, 0, previewSnapshot{}, false
	}
	frame := append([]byte(nil), preview.latestFrame...)
	return frame, preview.frameSeq, snapshotPreview(preview), true
}

func (m *PreviewManager) storePreviewFrame(previewID string, frame []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	preview, ok := m.byID[previewID]
	if !ok {
		return
	}
	preview.latestFrame = append([]byte(nil), frame...)
	preview.lastFrameAt = time.Now()
	preview.frameSeq++
	if preview.state == "starting" {
		preview.state = "running"
		preview.reason = ""
	}
	m.publishPreviewState(snapshotPreview(preview))
}

func (m *PreviewManager) finishPreview(previewID string, runErr error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	preview, ok := m.byID[previewID]
	if !ok {
		return
	}

	if runErr != nil {
		preview.state = "failed"
		preview.reason = runErr.Error()
	} else {
		preview.state = "finished"
		if preview.reason == "" {
			preview.reason = "preview finished"
		}
	}
	snapshot := snapshotPreview(preview)
	m.publishPreviewState(snapshot)

	if preview.leases == 0 || preview.state == "finished" {
		delete(m.byID, previewID)
		delete(m.bySignature, preview.signature)
	}
}

func (m *PreviewManager) publishPreviewState(snapshot previewSnapshot) {
	m.publish(ServerEvent{
		Type:      "preview.state",
		Timestamp: time.Now(),
		Payload:   mapPreviewResponse(snapshot),
	})
}

func snapshotPreview(preview *managedPreview) previewSnapshot {
	return previewSnapshot{
		ID:          preview.id,
		SourceID:    preview.source.ID,
		Name:        preview.source.Name,
		SourceType:  preview.source.Type,
		State:       preview.state,
		Reason:      preview.reason,
		Leases:      preview.leases,
		HasFrame:    len(preview.latestFrame) > 0,
		LastFrameAt: preview.lastFrameAt,
	}
}

func findPreviewSource(cfg *dsl.EffectiveConfig, sourceID string) (dsl.EffectiveVideoSource, error) {
	if cfg == nil {
		return dsl.EffectiveVideoSource{}, ErrPreviewSourceNotFound
	}
	for _, source := range cfg.VideoSources {
		if source.ID == sourceID && source.Enabled {
			return source, nil
		}
	}
	return dsl.EffectiveVideoSource{}, ErrPreviewSourceNotFound
}
