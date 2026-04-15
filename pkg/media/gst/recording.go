package gst

import (
	"context"
	stderrors "errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

var (
	errRecordingStopRequested = stderrors.New("recording stop requested")
	errRecordingMaxDuration   = stderrors.New("max duration reached")
	audioLevelParseFailures   = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_audio_level_parse_failures_total",
		"Total audio level message parse failures in the GStreamer recording runtime.",
		"reason",
		"rms_type",
	)
)

type RecordingRuntime struct {
	mu       sync.RWMutex
	sessions map[string]*recordingSession
}

func NewRecordingRuntime() *RecordingRuntime {
	return &RecordingRuntime{sessions: map[string]*recordingSession{}}
}

func (r *RecordingRuntime) StartRecording(ctx context.Context, plan *dsl.CompiledPlan, opts media.RecordingOptions) (media.RecordingSession, error) {
	if err := initGStreamer(); err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("compiled plan is required")
	}
	if len(plan.VideoJobs) == 0 && len(plan.AudioJobs) == 0 {
		return nil, errors.New("compiled plan has no executable jobs")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	runCtx, cancel := context.WithCancelCause(ctx)
	if opts.MaxDuration > 0 {
		go func() {
			timer := time.NewTimer(opts.MaxDuration)
			defer timer.Stop()
			select {
			case <-timer.C:
				cancel(fmt.Errorf("%w after %s", errRecordingMaxDuration, opts.MaxDuration))
			case <-runCtx.Done():
			}
		}()
	}
	eventCh := make(chan media.RecordingEvent, 128)
	workers := make([]*recordingWorker, 0, len(plan.VideoJobs)+len(plan.AudioJobs))
	for _, job := range plan.VideoJobs {
		worker, err := startVideoRecordingWorker(runCtx, job, eventCh)
		if err != nil {
			cancel(err)
			for _, started := range workers {
				_ = started.stop(250 * time.Millisecond)
				started.cleanup()
			}
			return nil, err
		}
		workers = append(workers, worker)
	}
	for _, job := range plan.AudioJobs {
		worker, err := startAudioRecordingWorker(runCtx, job, eventCh)
		if err != nil {
			cancel(err)
			for _, started := range workers {
				_ = started.stop(250 * time.Millisecond)
				started.cleanup()
			}
			return nil, err
		}
		workers = append(workers, worker)
	}

	session := &recordingSession{
		runtime:   r,
		sessionID: plan.SessionID,
		cancel:    cancel,
		done:      make(chan struct{}),
	}
	for _, worker := range workers {
		if worker.audio != nil {
			session.audioControls = append(session.audioControls, worker.audio)
		}
	}
	r.registerSession(session)
	go session.run(runCtx, plan, opts, workers, eventCh)
	return session, nil
}

type recordingSession struct {
	runtime       *RecordingRuntime
	sessionID     string
	cancel        context.CancelCauseFunc
	done          chan struct{}
	audioControls []*audioControls

	mu     sync.RWMutex
	result *media.RecordingResult
	err    error
}

func (r *RecordingRuntime) registerSession(session *recordingSession) {
	if r == nil || session == nil || session.sessionID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sessions == nil {
		r.sessions = map[string]*recordingSession{}
	}
	r.sessions[session.sessionID] = session
}

func (r *RecordingRuntime) unregisterSession(sessionID string) {
	if r == nil || sessionID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
}

func (r *RecordingRuntime) SetAudioGain(sessionID, sourceID string, gain float64) error {
	session := r.lookupSession(sessionID)
	if session == nil {
		return errors.New("recording session not found")
	}
	return session.SetAudioGain(sourceID, gain)
}

func (r *RecordingRuntime) SetAudioCompressorEnabled(sessionID string, enabled bool) error {
	session := r.lookupSession(sessionID)
	if session == nil {
		return errors.New("recording session not found")
	}
	return session.SetAudioCompressorEnabled(enabled)
}

func (r *RecordingRuntime) lookupSession(sessionID string) *recordingSession {
	if r == nil || sessionID == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sessions[sessionID]
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
		s.cancel(errRecordingStopRequested)
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

func (s *recordingSession) run(ctx context.Context, plan *dsl.CompiledPlan, opts media.RecordingOptions, workers []*recordingWorker, eventCh <-chan media.RecordingEvent) {
	defer close(s.done)
	defer func() {
		if s.runtime != nil {
			s.runtime.unregisterSession(s.sessionID)
		}
	}()
	startedAt := time.Now()
	emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, Timestamp: startedAt, State: media.RecordingStateStarting, Reason: "gstreamer recording session created"})
	for _, worker := range workers {
		emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventProcessStarted, ProcessLabel: worker.label, OutputPath: worker.outputPath})
		emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventProcessLog, ProcessLabel: worker.label, OutputPath: worker.outputPath, Stream: "system", Message: "gstreamer pipeline started"})
	}
	emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, State: media.RecordingStateRunning, Reason: "all gstreamer recording pipelines started"})

	result := &media.RecordingResult{
		State:     media.RecordingStateFinished,
		Reason:    "all gstreamer recording pipelines stopped cleanly",
		Outputs:   append([]dsl.PlannedOutput(nil), plan.Outputs...),
		StartedAt: startedAt,
	}

	type workerResult struct {
		worker *recordingWorker
		err    error
	}
	results := make(chan workerResult, len(workers))
	for _, worker := range workers {
		worker := worker
		go func() {
			results <- workerResult{worker: worker, err: worker.wait(ctx)}
		}()
	}

	remaining := len(workers)
	stoppingEmitted := false
	for remaining > 0 {
		select {
		case event := <-eventCh:
			emitRecordingEvent(opts.EventSink, event)
		case wr := <-results:
			remaining--
			if wr.err != nil && result.State != media.RecordingStateFailed {
				reason := fmt.Sprintf("%s failed: %v", wr.worker.label, wr.err)
				emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventProcessLog, ProcessLabel: wr.worker.label, OutputPath: wr.worker.outputPath, Stream: "stderr", Message: reason})
				if !stoppingEmitted {
					emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, State: media.RecordingStateStopping, Reason: reason})
					stoppingEmitted = true
				}
				result.State = media.RecordingStateFailed
				result.Reason = reason
				if s.cancel != nil {
					s.cancel(errors.New(reason))
				}
			}
		case <-ctx.Done():
			if !stoppingEmitted {
				reason := recordingContextReason(ctx)
				emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, State: media.RecordingStateStopping, Reason: reason})
				for _, worker := range workers {
					emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventProcessLog, ProcessLabel: worker.label, OutputPath: worker.outputPath, Stream: "system", Message: "stopping gstreamer pipeline via EOS"})
				}
				result.State = media.RecordingStateFinished
				result.Reason = reason
				stoppingEmitted = true
			}
		}
	}

	result.FinishedAt = time.Now()
	if result.State == media.RecordingStateFailed {
		emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, Timestamp: result.FinishedAt, State: media.RecordingStateFailed, Reason: result.Reason})
		s.setResult(result, errors.New(result.Reason))
		return
	}
	emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, Timestamp: result.FinishedAt, State: media.RecordingStateFinished, Reason: result.Reason})
	s.setResult(result, nil)
}

func (s *recordingSession) SetAudioGain(sourceID string, gain float64) error {
	if gain <= 0 {
		return errors.New("audio gain must be > 0")
	}
	s.mu.RLock()
	controls := append([]*audioControls(nil), s.audioControls...)
	s.mu.RUnlock()
	if len(controls) == 0 {
		return errors.New("recording session has no audio controls")
	}
	updated := 0
	for _, control := range controls {
		for id, element := range control.sourceVolumes {
			if sourceID != "" && sourceID != id {
				continue
			}
			element.Set("volume", gain)
			updated++
		}
	}
	if updated == 0 {
		return fmt.Errorf("audio source %q not found", sourceID)
	}
	return nil
}

func (s *recordingSession) SetAudioCompressorEnabled(enabled bool) error {
	s.mu.RLock()
	controls := append([]*audioControls(nil), s.audioControls...)
	s.mu.RUnlock()
	if len(controls) == 0 {
		return errors.New("recording session has no audio compressor")
	}
	updated := 0
	for _, control := range controls {
		if control.compressor == nil {
			continue
		}
		if enabled {
			control.compressor.Set("ratio", 4.0)
			control.compressor.Set("threshold", 0.5)
		} else {
			control.compressor.Set("ratio", 1.0)
			control.compressor.Set("threshold", 1.0)
		}
		updated++
	}
	if updated == 0 {
		return errors.New("recording session has no audio compressor")
	}
	return nil
}

func (s *recordingSession) setResult(result *media.RecordingResult, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result = result
	s.err = err
}

type audioControls struct {
	sourceVolumes map[string]*gst.Element
	compressor    *gst.Element
}

type recordingWorker struct {
	label      string
	outputPath string
	pipeline   *gst.Pipeline
	watch      *busWatch
	resultCh   chan error
	events     chan media.RecordingEvent
	audio      *audioControls
	stopFn     func(time.Duration) error
	cleanupFn  func()
}

func startVideoRecordingWorker(ctx context.Context, job dsl.VideoJob, eventCh chan media.RecordingEvent) (*recordingWorker, error) {
	resolvedSource, err := resolveWindowSource(ctx, job.Source)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	pipeline, err := buildVideoRecordingPipeline(resolvedSource, job.OutputPath)
	if err != nil {
		return nil, err
	}
	bus := pipeline.GetPipelineBus()
	resultCh := make(chan error, 1)
	watch, err := startBusWatch(bus, func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			gstErr := msg.ParseError()
			err := fmt.Errorf("video recording pipeline error: %w", gstErr)
			select {
			case resultCh <- err:
			default:
			}
			return false
		case gst.MessageEOS:
			select {
			case resultCh <- nil:
			default:
			}
			return false
		default:
			return true
		}
	})
	if err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, err
	}
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		watch.Stop()
		pipeline.BlockSetState(gst.StateNull)
		return nil, fmt.Errorf("start video recording pipeline: %w", err)
	}
	log.Info().
		Str("event", "recording.gst.start").
		Str("process_label", resolvedSource.Name).
		Str("output_path", job.OutputPath).
		Str("source_type", resolvedSource.Type).
		Str("recording_mode", "direct-pipeline").
		Msg("started gstreamer recording pipeline")
	return &recordingWorker{
		label:      resolvedSource.Name,
		outputPath: job.OutputPath,
		pipeline:   pipeline,
		watch:      watch,
		resultCh:   resultCh,
		events:     eventCh,
	}, nil
}

func (w *recordingWorker) wait(ctx context.Context) error {
	defer w.cleanup()
	select {
	case err := <-w.resultCh:
		return err
	case <-ctx.Done():
		return w.stop(10 * time.Second)
	}
}

func (w *recordingWorker) stop(timeout time.Duration) error {
	if w == nil {
		return nil
	}
	if w.stopFn != nil {
		return w.stopFn(timeout)
	}
	if w.pipeline == nil {
		return nil
	}
	sendEOS(w.pipeline)
	select {
	case err := <-w.resultCh:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for recording EOS")
	}
}

func (w *recordingWorker) cleanup() {
	if w == nil {
		return
	}
	if w.cleanupFn != nil {
		w.cleanupFn()
		return
	}
	if w.watch != nil {
		w.watch.Stop()
	}
	if w.pipeline != nil {
		w.pipeline.BlockSetState(gst.StateNull)
	}
}

func startAudioRecordingWorker(ctx context.Context, job dsl.AudioMixJob, eventCh chan media.RecordingEvent) (*recordingWorker, error) {
	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	pipeline, audioControl, err := buildAudioRecordingPipeline(job)
	if err != nil {
		return nil, err
	}
	bus := pipeline.GetPipelineBus()
	resultCh := make(chan error, 1)
	watch, err := startBusWatch(bus, func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			gstErr := msg.ParseError()
			err := fmt.Errorf("audio recording pipeline error: %w", gstErr)
			select {
			case resultCh <- err:
			default:
			}
			return false
		case gst.MessageEOS:
			select {
			case resultCh <- nil:
			default:
			}
			return false
		case gst.MessageElement:
			if event, ok := audioLevelEventFromMessage(job, msg); ok {
				select {
				case eventCh <- event:
				default:
				}
			}
			return true
		default:
			return true
		}
	})
	if err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, err
	}
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		watch.Stop()
		pipeline.BlockSetState(gst.StateNull)
		return nil, fmt.Errorf("start audio recording pipeline: %w", err)
	}
	log.Info().
		Str("event", "recording.gst.start").
		Str("process_label", job.Name).
		Str("output_path", job.OutputPath).
		Int("source_count", len(job.Sources)).
		Msg("started gstreamer audio recording pipeline")
	return &recordingWorker{label: job.Name, outputPath: job.OutputPath, pipeline: pipeline, watch: watch, resultCh: resultCh, events: eventCh, audio: audioControl}, nil
}

func buildVideoRecordingPipeline(source dsl.EffectiveVideoSource, outputPath string) (*gst.Pipeline, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	elements := []*gst.Element{}
	sourceElements, err := buildSourceElements(source)
	if err != nil {
		return nil, err
	}
	elements = append(elements, sourceElements...)

	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, fmt.Errorf("create videoconvert: %w", err)
	}
	elements = append(elements, videoconvert)

	if source.Type == "camera" && boolValue(source.Capture.Mirror, false) {
		videoflip, err := gst.NewElement("videoflip")
		if err != nil {
			return nil, fmt.Errorf("create videoflip: %w", err)
		}
		videoflip.Set("method", "horizontal-flip")
		elements = append(elements, videoflip)
	}

	fps := source.Capture.FPS
	if fps <= 0 {
		fps = 24
	}
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, fmt.Errorf("create videorate: %w", err)
	}
	fpsCaps, err := newCapsFilter(fmt.Sprintf("video/x-raw,format=I420,framerate=%d/1,pixel-aspect-ratio=1/1", fps))
	if err != nil {
		return nil, err
	}
	x264enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, fmt.Errorf("create x264enc: %w", err)
	}
	x264enc.Set("bitrate", qualityToBitrate(source.Output.Quality))
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)

	mux, err := newVideoMuxer(source.Output.Container)
	if err != nil {
		return nil, err
	}
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, fmt.Errorf("create filesink: %w", err)
	}
	filesink.Set("location", outputPath)

	elements = append(elements, videorate, fpsCaps, x264enc, mux, filesink)
	pipeline.AddMany(elements...)
	if err := linkElements(elements...); err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, err
	}
	return pipeline, nil
}

func buildAudioRecordingPipeline(job dsl.AudioMixJob) (*gst.Pipeline, *audioControls, error) {
	if len(job.Sources) == 0 {
		return nil, nil, errors.New("audio mix job has no sources")
	}
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, nil, fmt.Errorf("create pipeline: %w", err)
	}
	controls := &audioControls{sourceVolumes: map[string]*gst.Element{}}

	mixer, err := gst.NewElement("audiomixer")
	if err != nil {
		return nil, nil, fmt.Errorf("create audiomixer: %w", err)
	}
	pipeline.Add(mixer)

	for i, src := range job.Sources {
		branch, volume, err := buildAudioSourceBranch(src, job.Output)
		if err != nil {
			return nil, nil, err
		}
		controls.sourceVolumes[src.ID] = volume
		pipeline.AddMany(branch...)
		if err := linkElements(branch...); err != nil {
			pipeline.BlockSetState(gst.StateNull)
			return nil, nil, err
		}
		srcPad := volume.GetStaticPad("src")
		if srcPad == nil {
			return nil, nil, fmt.Errorf("audio branch %d missing src pad", i)
		}
		sinkPad := mixer.GetRequestPad("sink_%u")
		if sinkPad == nil {
			return nil, nil, fmt.Errorf("audio mixer request pad %d failed", i)
		}
		if ret := srcPad.Link(sinkPad); ret != gst.PadLinkOK {
			return nil, nil, fmt.Errorf("link audio branch %d to mixer: %s", i, ret.String())
		}
	}

	postMixer, compressor, err := buildAudioOutputChain(job.Output, job.OutputPath)
	if err != nil {
		return nil, nil, err
	}
	controls.compressor = compressor
	pipeline.AddMany(postMixer...)
	chain := append([]*gst.Element{mixer}, postMixer...)
	if err := linkElements(chain...); err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, nil, err
	}
	return pipeline, controls, nil
}

func buildAudioSourceBranch(src dsl.EffectiveAudioSource, out dsl.AudioOutputSettings) ([]*gst.Element, *gst.Element, error) {
	pulsesrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, nil, fmt.Errorf("create pulsesrc: %w", err)
	}
	pulsesrc.Set("device", src.Device)
	caps, err := newCapsFilter(audioRawCaps(out))
	if err != nil {
		return nil, nil, err
	}
	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, nil, fmt.Errorf("create audioconvert: %w", err)
	}
	audioresample, err := gst.NewElement("audioresample")
	if err != nil {
		return nil, nil, fmt.Errorf("create audioresample: %w", err)
	}
	volume, err := gst.NewElement("volume")
	if err != nil {
		return nil, nil, fmt.Errorf("create volume: %w", err)
	}
	gain := src.Settings.Gain
	if gain <= 0 {
		gain = 1.0
	}
	volume.Set("volume", gain)
	return []*gst.Element{pulsesrc, caps, audioconvert, audioresample, volume}, volume, nil
}

func buildAudioOutputChain(out dsl.AudioOutputSettings, outputPath string) ([]*gst.Element, *gst.Element, error) {
	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, nil, fmt.Errorf("create audioconvert: %w", err)
	}
	audioresample, err := gst.NewElement("audioresample")
	if err != nil {
		return nil, nil, fmt.Errorf("create audioresample: %w", err)
	}
	caps, err := newCapsFilter(audioEncoderCaps(out))
	if err != nil {
		return nil, nil, err
	}
	compressor, err := gst.NewElement("audiodynamic")
	if err != nil {
		return nil, nil, fmt.Errorf("create audiodynamic: %w", err)
	}
	compressor.Set("mode", 0)
	compressor.Set("ratio", 1.0)
	compressor.Set("threshold", 1.0)
	level, err := gst.NewElement("level")
	if err != nil {
		return nil, nil, fmt.Errorf("create level: %w", err)
	}
	level.Set("interval", uint64(50_000_000))
	level.Set("post-messages", true)
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, nil, fmt.Errorf("create filesink: %w", err)
	}
	filesink.Set("location", outputPath)

	switch strings.ToLower(strings.TrimSpace(out.Codec)) {
	case "", "pcm_s16le", "wav":
		wavenc, err := gst.NewElement("wavenc")
		if err != nil {
			return nil, nil, fmt.Errorf("create wavenc: %w", err)
		}
		return []*gst.Element{audioconvert, audioresample, caps, compressor, level, wavenc, filesink}, compressor, nil
	case "opus":
		opusenc, err := gst.NewElement("opusenc")
		if err != nil {
			return nil, nil, fmt.Errorf("create opusenc: %w", err)
		}
		opusenc.Set("bitrate", 160000)
		oggmux, err := gst.NewElement("oggmux")
		if err != nil {
			return nil, nil, fmt.Errorf("create oggmux: %w", err)
		}
		return []*gst.Element{audioconvert, audioresample, caps, compressor, level, opusenc, oggmux, filesink}, compressor, nil
	default:
		return nil, nil, fmt.Errorf("unsupported gstreamer audio codec %q", out.Codec)
	}
}

func newVideoMuxer(container string) (*gst.Element, error) {
	switch strings.ToLower(strings.TrimSpace(container)) {
	case "", "mp4":
		mux, err := gst.NewElement("mp4mux")
		if err != nil {
			return nil, fmt.Errorf("create mp4mux: %w", err)
		}
		return mux, nil
	case "mov", "qt":
		mux, err := gst.NewElement("qtmux")
		if err != nil {
			return nil, fmt.Errorf("create qtmux: %w", err)
		}
		return mux, nil
	default:
		return nil, fmt.Errorf("unsupported gstreamer video container %q", container)
	}
}

func qualityToBitrate(quality int) int {
	if quality < 1 {
		quality = 75
	}
	if quality > 100 {
		quality = 100
	}
	return 1000 + (quality-1)*80
}

func audioRawCaps(out dsl.AudioOutputSettings) string {
	rate := out.SampleRateHz
	if rate <= 0 {
		rate = 48000
	}
	channels := out.Channels
	if channels <= 0 {
		channels = 2
	}
	return fmt.Sprintf("audio/x-raw,format=S16LE,rate=%d,channels=%d", rate, channels)
}

func audioEncoderCaps(out dsl.AudioOutputSettings) string {
	rate := out.SampleRateHz
	if rate <= 0 {
		rate = 48000
	}
	channels := out.Channels
	if channels <= 0 {
		channels = 2
	}
	switch strings.ToLower(strings.TrimSpace(out.Codec)) {
	case "opus":
		return fmt.Sprintf("audio/x-raw,format=S16LE,rate=48000,channels=%d", channels)
	default:
		return fmt.Sprintf("audio/x-raw,format=S16LE,rate=%d,channels=%d", rate, channels)
	}
}

func audioLevelEventFromMessage(job dsl.AudioMixJob, msg *gst.Message) (media.RecordingEvent, bool) {
	if msg == nil || msg.Type() != gst.MessageElement {
		return media.RecordingEvent{}, false
	}
	st := msg.GetStructure()
	if st == nil || st.Name() != "level" {
		return media.RecordingEvent{}, false
	}
	rms, err := st.GetValue("rms")
	if err != nil {
		audioLevelParseFailures.Inc(map[string]string{"reason": "missing_rms", "rms_type": ""})
		log.Debug().Str("event", "recording.gst.audio_level.parse.missing_rms").Str("structure", st.Name()).Interface("values", st.Values()).Msg("audio level message missing rms field")
		return media.RecordingEvent{}, false
	}
	left, right, ok := extractLevels(rms)
	if !ok {
		audioLevelParseFailures.Inc(map[string]string{"reason": "extract_failed", "rms_type": fmt.Sprintf("%T", rms)})
		log.Debug().Str("event", "recording.gst.audio_level.parse.failed").Str("structure", st.Name()).Str("rms_type", fmt.Sprintf("%T", rms)).Interface("rms", rms).Interface("values", st.Values()).Msg("failed to parse audio level message")
		return media.RecordingEvent{
			Type:         media.RecordingEventAudioLevel,
			ProcessLabel: job.Name,
			OutputPath:   job.OutputPath,
			DeviceID:     firstAudioDevice(job.Sources),
			LeftLevel:    0.5,
			RightLevel:   0.5,
			Available:    true,
		}, true
	}
	return media.RecordingEvent{
		Type:         media.RecordingEventAudioLevel,
		ProcessLabel: job.Name,
		OutputPath:   job.OutputPath,
		DeviceID:     firstAudioDevice(job.Sources),
		LeftLevel:    dbToLinear(left),
		RightLevel:   dbToLinear(right),
		Available:    true,
	}, true
}

func extractLevels(value interface{}) (float64, float64, bool) {
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return 0, 0, false
		}
		left, ok := asFloat64(v[0])
		if !ok {
			return 0, 0, false
		}
		right := left
		if len(v) > 1 {
			if parsed, ok := asFloat64(v[1]); ok {
				right = parsed
			}
		}
		return left, right, true
	case []float64:
		if len(v) == 0 {
			return 0, 0, false
		}
		right := v[0]
		if len(v) > 1 {
			right = v[1]
		}
		return v[0], right, true
	default:
		return 0, 0, false
	}
}

func asFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func dbToLinear(db float64) float64 {
	if db <= -60 {
		return 0
	}
	return mathPow10(db / 20)
}

func mathPow10(power float64) float64 {
	return math.Pow(10, power)
}

func firstAudioDevice(sources []dsl.EffectiveAudioSource) string {
	if len(sources) == 0 {
		return ""
	}
	return sources[0].Device
}

func emitRecordingEvent(sink func(media.RecordingEvent), event media.RecordingEvent) {
	if sink == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	sink(event)
}

func recordingContextReason(ctx context.Context) string {
	if ctx == nil {
		return "recording stopped"
	}
	cause := context.Cause(ctx)
	switch {
	case cause == nil:
		if ctx.Err() != nil {
			return ctx.Err().Error()
		}
		return "recording stopped"
	case stderrors.Is(cause, errRecordingStopRequested):
		return "recording stop requested"
	case stderrors.Is(cause, errRecordingMaxDuration):
		return cause.Error()
	default:
		return cause.Error()
	}
}
