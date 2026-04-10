package web

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
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
	app       ApplicationService
	publish   func(ServerEvent)
	parentCtx context.Context
	limit     int
	runner    PreviewRunner

	mu          sync.RWMutex
	byID        map[string]*managedPreview
	bySignature map[string]*managedPreview
}

func NewPreviewManager(parentCtx context.Context, application ApplicationService, publish func(ServerEvent), limit int, runner PreviewRunner) *PreviewManager {
	if publish == nil {
		publish = func(ServerEvent) {}
	}
	if limit <= 0 {
		limit = 4
	}
	if runner == nil {
		runner = FFmpegPreviewRunner{}
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	return &PreviewManager{
		app:         application,
		publish:     publish,
		parentCtx:   parentCtx,
		limit:       limit,
		runner:      runner,
		byID:        map[string]*managedPreview{},
		bySignature: map[string]*managedPreview{},
	}
}

func (m *PreviewManager) parentContext() context.Context {
	if m.parentCtx == nil {
		return context.Background()
	}
	return m.parentCtx
}

func (m *PreviewManager) Ensure(ctx context.Context, dslBody []byte, sourceID string) (previewSnapshot, error) {
	log.Info().
		Str("event", "preview.ensure.requested").
		Str("source_id", sourceID).
		Int("dsl_bytes", len(dslBody)).
		Msg("preview ensure requested")

	cfg, err := m.app.NormalizeDSL(ctx, dslBody)
	if err != nil {
		log.Error().
			Str("event", "preview.ensure.normalize.error").
			Str("source_id", sourceID).
			Err(err).
			Msg("failed to normalize preview dsl")
		return previewSnapshot{}, err
	}

	source, err := findPreviewSource(cfg, sourceID)
	if err != nil {
		log.Error().
			Str("event", "preview.ensure.source.error").
			Str("source_id", sourceID).
			Err(err).
			Msg("failed to resolve preview source")
		return previewSnapshot{}, err
	}
	signature := computePreviewSignature(source)

	m.mu.Lock()
	if existing, ok := m.bySignature[signature]; ok && existing.state != "failed" {
		existing.leases++
		snapshot := snapshotPreview(existing)
		m.mu.Unlock()
		log.Info().
			Str("event", "preview.ensure.reused").
			Str("preview_id", snapshot.ID).
			Str("source_id", snapshot.SourceID).
			Int("leases", snapshot.Leases).
			Msg("reused existing preview")
		m.publishPreviewState(snapshot)
		return snapshot, nil
	}
	if len(m.byID) >= m.limit {
		m.mu.Unlock()
		log.Warn().
			Str("event", "preview.ensure.limit_exceeded").
			Str("source_id", sourceID).
			Int("limit", m.limit).
			Msg("preview limit exceeded")
		return previewSnapshot{}, ErrPreviewLimitExceeded
	}

	previewCtx, cancel := context.WithCancel(m.parentContext())
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

	log.Info().
		Str("event", "preview.ensure.created").
		Str("preview_id", preview.id).
		Str("source_id", source.ID).
		Str("source_name", source.Name).
		Str("source_type", source.Type).
		Msg("created preview worker")

	m.publishPreviewState(snapshot)

	group, groupCtx := errgroup.WithContext(previewCtx)
	group.Go(func() error {
		defer close(preview.done)
		log.Info().
			Str("event", "preview.run.begin").
			Str("preview_id", preview.id).
			Str("source_id", source.ID).
			Msg("preview runner starting")
		err := m.runner.Run(groupCtx, source, func(frame []byte) {
			m.storePreviewFrame(preview.id, frame)
		}, func(stream, line string) {
			logTimestamp := time.Now()
			m.publish(ServerEvent{
				Type:      "preview.log",
				Timestamp: logTimestamp,
				Payload: &studiov1.ProcessLog{
					Timestamp:    formatTimestamp(logTimestamp),
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
		log.Warn().
			Str("event", "preview.release.not_found").
			Str("preview_id", previewID).
			Msg("preview release requested for unknown preview")
		return previewSnapshot{}, ErrPreviewNotFound
	}

	if preview.leases > 0 {
		preview.leases--
	}
	if preview.leases == 0 {
		preview.state = "stopping"
		preview.reason = "released"
		log.Info().
			Str("event", "preview.release.cancel").
			Str("preview_id", previewID).
			Str("source_id", preview.source.ID).
			Msg("preview release triggered cancellation")
		preview.cancel()
	}

	snapshot := snapshotPreview(preview)
	log.Info().
		Str("event", "preview.release.done").
		Str("preview_id", snapshot.ID).
		Str("state", snapshot.State).
		Int("leases", snapshot.Leases).
		Msg("preview release handled")
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
		log.Warn().
			Str("event", "preview.finish.missing").
			Str("preview_id", previewID).
			Msg("preview finished after manager entry was removed")
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
	if runErr != nil {
		log.Error().
			Str("event", "preview.finish").
			Str("preview_id", previewID).
			Str("source_id", preview.source.ID).
			Str("state", snapshot.State).
			Str("reason", snapshot.Reason).
			Err(runErr).
			Msg("preview finished with error")
	} else {
		log.Info().
			Str("event", "preview.finish").
			Str("preview_id", previewID).
			Str("source_id", preview.source.ID).
			Str("state", snapshot.State).
			Str("reason", snapshot.Reason).
			Msg("preview finished")
	}
	m.publishPreviewState(snapshot)

	if preview.leases == 0 || preview.state == "finished" {
		log.Info().
			Str("event", "preview.cleanup").
			Str("preview_id", previewID).
			Str("signature", preview.signature).
			Msg("removing preview from manager maps")
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
