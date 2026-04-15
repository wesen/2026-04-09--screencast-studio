package web

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstreamermedia "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
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
	session   media.PreviewSession

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
	runtime   media.PreviewRuntime

	mu          sync.RWMutex
	byID        map[string]*managedPreview
	bySignature map[string]*managedPreview
}

func NewPreviewManager(parentCtx context.Context, application ApplicationService, publish func(ServerEvent), limit int, runtime media.PreviewRuntime) *PreviewManager {
	if publish == nil {
		publish = func(ServerEvent) {}
	}
	if limit <= 0 {
		limit = 4
	}
	if runtime == nil {
		runtime = gstreamermedia.NewPreviewRuntime()
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	return &PreviewManager{
		app:         application,
		publish:     publish,
		parentCtx:   parentCtx,
		limit:       limit,
		runtime:     runtime,
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
	m.pruneTerminalPreviewsLocked()
	if existing, ok := m.bySignature[signature]; ok {
		if previewReusable(existing) {
			existing.leases++
			snapshot := snapshotPreview(existing)
			m.mu.Unlock()
			previewEnsures.Inc(previewMetricLabelsWithResult(snapshot.SourceType, "reused"))
			log.Info().
				Str("event", "preview.ensure.reused").
				Str("preview_id", snapshot.ID).
				Str("source_id", snapshot.SourceID).
				Int("leases", snapshot.Leases).
				Msg("reused existing preview")
			m.publishPreviewState(snapshot)
			return snapshot, nil
		}
		if existing.leases == 0 {
			delete(m.bySignature, signature)
		}
	}
	if m.activePreviewCountLocked() >= m.limit {
		m.mu.Unlock()
		previewEnsures.Inc(previewMetricLabelsWithResult(source.Type, "limit_exceeded"))
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

	previewEnsures.Inc(previewMetricLabelsWithResult(source.Type, "created"))
	log.Info().
		Str("event", "preview.ensure.created").
		Str("preview_id", preview.id).
		Str("source_id", source.ID).
		Str("source_name", source.Name).
		Str("source_type", source.Type).
		Msg("created preview worker")

	m.publishPreviewState(snapshot)

	session, err := m.runtime.StartPreview(previewCtx, source, media.PreviewOptions{
		OnFrame: func(frame []byte) {
			m.storePreviewFrame(preview.id, frame)
		},
		OnLog: func(stream, line string) {
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
		},
	})
	if err != nil {
		m.mu.Lock()
		delete(m.byID, preview.id)
		delete(m.bySignature, preview.signature)
		m.mu.Unlock()
		cancel()
		return previewSnapshot{}, err
	}

	m.mu.Lock()
	preview.session = session
	m.mu.Unlock()

	go func() {
		defer close(preview.done)
		log.Info().
			Str("event", "preview.run.begin").
			Str("preview_id", preview.id).
			Str("source_id", source.ID).
			Msg("preview session waiting")
		err := session.Wait()
		m.finishPreview(preview.id, err)
	}()

	return snapshot, nil
}

func (m *PreviewManager) Release(previewID string) (previewSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	preview, ok := m.byID[previewID]
	if !ok {
		previewReleases.Inc(previewMetricLabelsWithResult("unknown", "not_found"))
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
		delete(m.bySignature, preview.signature)
		log.Info().
			Str("event", "preview.release.cancel").
			Str("preview_id", previewID).
			Str("source_id", preview.source.ID).
			Msg("preview release triggered cancellation")
		preview.cancel()
	}

	snapshot := snapshotPreview(preview)
	previewReleases.Inc(previewMetricLabelsWithResult(snapshot.SourceType, "released"))
	log.Info().
		Str("event", "preview.release.done").
		Str("preview_id", snapshot.ID).
		Str("state", snapshot.State).
		Int("leases", snapshot.Leases).
		Msg("preview release handled")
	m.publishPreviewState(snapshot)
	return snapshot, nil
}

func (m *PreviewManager) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	type shutdownTarget struct {
		id       string
		sourceID string
		cancel   context.CancelFunc
		done     chan struct{}
	}

	m.mu.Lock()
	targets := make([]shutdownTarget, 0, len(m.byID))
	snapshots := make([]previewSnapshot, 0, len(m.byID))
	for _, preview := range m.byID {
		if preview.state != "finished" && preview.state != "failed" {
			preview.state = "stopping"
			if preview.reason == "" {
				preview.reason = "shutdown requested"
			}
		}
		targets = append(targets, shutdownTarget{
			id:       preview.id,
			sourceID: preview.source.ID,
			cancel:   preview.cancel,
			done:     preview.done,
		})
		snapshots = append(snapshots, snapshotPreview(preview))
	}
	m.mu.Unlock()

	if len(targets) == 0 {
		log.Info().
			Str("event", "preview.shutdown.noop").
			Msg("preview manager shutdown requested with no active previews")
		return nil
	}

	log.Info().
		Str("event", "preview.shutdown.begin").
		Int("preview_count", len(targets)).
		Msg("preview manager shutdown starting")

	for _, snapshot := range snapshots {
		m.publishPreviewState(snapshot)
	}
	for _, target := range targets {
		log.Info().
			Str("event", "preview.shutdown.cancel").
			Str("preview_id", target.id).
			Str("source_id", target.sourceID).
			Msg("preview manager requested preview cancellation")
		if target.cancel != nil {
			target.cancel()
		}
	}

	for i, target := range targets {
		if target.done == nil {
			continue
		}
		select {
		case <-target.done:
			log.Info().
				Str("event", "preview.shutdown.wait.done").
				Str("preview_id", target.id).
				Str("source_id", target.sourceID).
				Int("remaining", len(targets)-i-1).
				Msg("preview shutdown wait completed")
		case <-ctx.Done():
			pending := []string{target.id}
			for _, remaining := range targets[i+1:] {
				pending = append(pending, remaining.id)
			}
			log.Error().
				Str("event", "preview.shutdown.timeout").
				Strs("pending_previews", pending).
				Err(ctx.Err()).
				Msg("preview manager shutdown timed out")
			return fmt.Errorf("preview shutdown timed out waiting for %v: %w", pending, ctx.Err())
		}
	}

	log.Info().
		Str("event", "preview.shutdown.done").
		Int("preview_count", len(targets)).
		Msg("preview manager shutdown finished")
	return nil
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
	copyStart := time.Now()
	frame := append([]byte(nil), preview.latestFrame...)
	previewLatestFrameCopyNanoseconds.Add(previewMetricLabels(preview.source.Type), uint64(time.Since(copyStart)))
	return frame, preview.frameSeq, snapshotPreview(preview), true
}

func (m *PreviewManager) TakeScreenshot(ctx context.Context, previewID string) ([]byte, previewSnapshot, error) {
	m.mu.RLock()
	preview, ok := m.byID[previewID]
	if !ok {
		m.mu.RUnlock()
		return nil, previewSnapshot{}, ErrPreviewNotFound
	}
	if preview.session == nil {
		snapshot := snapshotPreview(preview)
		m.mu.RUnlock()
		return nil, snapshot, errors.New("preview session not available")
	}
	session := preview.session
	snapshot := snapshotPreview(preview)
	m.mu.RUnlock()

	shot, err := session.TakeScreenshot(ctx, media.ScreenshotOptions{})
	if err != nil {
		return nil, snapshot, err
	}
	return shot, snapshot, nil
}

func (m *PreviewManager) storePreviewFrame(previewID string, frame []byte) {
	storeStart := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	preview, ok := m.byID[previewID]
	if !ok {
		return
	}
	labels := previewMetricLabels(preview.source.Type)
	preview.latestFrame = append([]byte(nil), frame...)
	preview.lastFrameAt = time.Now()
	preview.frameSeq++
	previewFrameUpdates.Inc(labels)
	if preview.state == "starting" {
		preview.state = "running"
		preview.reason = ""
	}
	snapshot := snapshotPreview(preview)
	publishStart := time.Now()
	m.publishPreviewState(snapshot)
	previewStatePublishNanoseconds.Add(labels, uint64(time.Since(publishStart)))
	previewFrameStoreNanoseconds.Add(labels, uint64(time.Since(storeStart)))
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

func (m *PreviewManager) pruneTerminalPreviewsLocked() {
	for previewID, preview := range m.byID {
		if preview == nil {
			delete(m.byID, previewID)
			continue
		}
		if preview.state != "finished" && preview.state != "failed" {
			continue
		}
		delete(m.byID, previewID)
		if current := m.bySignature[preview.signature]; current == preview {
			delete(m.bySignature, preview.signature)
		}
	}
}

func (m *PreviewManager) activePreviewCountLocked() int {
	count := 0
	for _, preview := range m.byID {
		if previewCountsAgainstLimit(preview) {
			count++
		}
	}
	return count
}

func previewReusable(preview *managedPreview) bool {
	if preview == nil {
		return false
	}
	return preview.state == "starting" || preview.state == "running"
}

func previewCountsAgainstLimit(preview *managedPreview) bool {
	if preview == nil {
		return false
	}
	if preview.state == "finished" || preview.state == "failed" {
		return false
	}
	if preview.state == "stopping" && preview.leases == 0 {
		return false
	}
	return true
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
