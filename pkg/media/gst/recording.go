package gst

import (
	"context"
	"fmt"
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
)

type RecordingRuntime struct{}

func NewRecordingRuntime() *RecordingRuntime {
	return &RecordingRuntime{}
}

func (RecordingRuntime) StartRecording(ctx context.Context, plan *dsl.CompiledPlan, opts media.RecordingOptions) (media.RecordingSession, error) {
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

	runCtx, cancel := context.WithCancel(ctx)
	workers := make([]*recordingWorker, 0, len(plan.VideoJobs)+len(plan.AudioJobs))
	for _, job := range plan.VideoJobs {
		worker, err := startVideoRecordingWorker(runCtx, job)
		if err != nil {
			cancel()
			for _, started := range workers {
				_ = started.stop(250 * time.Millisecond)
				started.cleanup()
			}
			return nil, err
		}
		workers = append(workers, worker)
	}
	for _, job := range plan.AudioJobs {
		worker, err := startAudioRecordingWorker(runCtx, job)
		if err != nil {
			cancel()
			for _, started := range workers {
				_ = started.stop(250 * time.Millisecond)
				started.cleanup()
			}
			return nil, err
		}
		workers = append(workers, worker)
	}

	session := &recordingSession{
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go session.run(runCtx, plan, opts, workers)
	return session, nil
}

type recordingSession struct {
	cancel context.CancelFunc
	done   chan struct{}

	mu     sync.RWMutex
	result *media.RecordingResult
	err    error
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
		s.cancel()
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

func (s *recordingSession) run(ctx context.Context, plan *dsl.CompiledPlan, opts media.RecordingOptions, workers []*recordingWorker) {
	defer close(s.done)
	startedAt := time.Now()
	emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventStateChanged, Timestamp: startedAt, State: media.RecordingStateStarting, Reason: "gstreamer recording session created"})
	for _, worker := range workers {
		emitRecordingEvent(opts.EventSink, media.RecordingEvent{Type: media.RecordingEventProcessStarted, ProcessLabel: worker.label, OutputPath: worker.outputPath})
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
	canceled := false
	for remaining > 0 {
		select {
		case wr := <-results:
			remaining--
			if wr.err != nil && result.State != media.RecordingStateFailed {
				result.State = media.RecordingStateFailed
				result.Reason = fmt.Sprintf("%s failed: %v", wr.worker.label, wr.err)
				if s.cancel != nil {
					s.cancel()
				}
			}
		case <-ctx.Done():
			if !canceled {
				canceled = true
				result.State = media.RecordingStateFinished
				result.Reason = ctx.Err().Error()
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

func (s *recordingSession) setResult(result *media.RecordingResult, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result = result
	s.err = err
}

type recordingWorker struct {
	label      string
	outputPath string
	pipeline   *gst.Pipeline
	watch      *busWatch
	resultCh   chan error
}

func startVideoRecordingWorker(ctx context.Context, job dsl.VideoJob) (*recordingWorker, error) {
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
			err := fmt.Errorf("recording pipeline error: %w", gstErr)
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
		return nil, fmt.Errorf("start recording pipeline: %w", err)
	}
	log.Info().
		Str("event", "recording.gst.start").
		Str("process_label", resolvedSource.Name).
		Str("output_path", job.OutputPath).
		Str("source_type", resolvedSource.Type).
		Msg("started gstreamer recording pipeline")
	return &recordingWorker{label: resolvedSource.Name, outputPath: job.OutputPath, pipeline: pipeline, watch: watch, resultCh: resultCh}, nil
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
	if w == nil || w.pipeline == nil {
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
	if w.watch != nil {
		w.watch.Stop()
	}
	if w.pipeline != nil {
		w.pipeline.BlockSetState(gst.StateNull)
	}
}

func startAudioRecordingWorker(ctx context.Context, job dsl.AudioMixJob) (*recordingWorker, error) {
	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	pipeline, err := buildAudioRecordingPipeline(job)
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
	return &recordingWorker{label: job.Name, outputPath: job.OutputPath, pipeline: pipeline, watch: watch, resultCh: resultCh}, nil
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
	fpsCaps, err := newCapsFilter(fmt.Sprintf("video/x-raw,framerate=%d/1", fps))
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

func buildAudioRecordingPipeline(job dsl.AudioMixJob) (*gst.Pipeline, error) {
	if len(job.Sources) == 0 {
		return nil, errors.New("audio mix job has no sources")
	}
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	mixer, err := gst.NewElement("audiomixer")
	if err != nil {
		return nil, fmt.Errorf("create audiomixer: %w", err)
	}
	pipeline.Add(mixer)

	for i, src := range job.Sources {
		branch, volume, err := buildAudioSourceBranch(src, job.Output)
		if err != nil {
			return nil, err
		}
		pipeline.AddMany(branch...)
		if err := linkElements(branch...); err != nil {
			pipeline.BlockSetState(gst.StateNull)
			return nil, err
		}
		srcPad := volume.GetStaticPad("src")
		if srcPad == nil {
			return nil, fmt.Errorf("audio branch %d missing src pad", i)
		}
		sinkPad := mixer.GetRequestPad("sink_%u")
		if sinkPad == nil {
			return nil, fmt.Errorf("audio mixer request pad %d failed", i)
		}
		if ret := srcPad.Link(sinkPad); ret != gst.PadLinkOK {
			return nil, fmt.Errorf("link audio branch %d to mixer: %s", i, ret.String())
		}
	}

	postMixer, err := buildAudioOutputChain(job.Output, job.OutputPath)
	if err != nil {
		return nil, err
	}
	pipeline.AddMany(postMixer...)
	chain := append([]*gst.Element{mixer}, postMixer...)
	if err := linkElements(chain...); err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, err
	}
	return pipeline, nil
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

func buildAudioOutputChain(out dsl.AudioOutputSettings, outputPath string) ([]*gst.Element, error) {
	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, fmt.Errorf("create audioconvert: %w", err)
	}
	audioresample, err := gst.NewElement("audioresample")
	if err != nil {
		return nil, fmt.Errorf("create audioresample: %w", err)
	}
	caps, err := newCapsFilter(audioEncoderCaps(out))
	if err != nil {
		return nil, err
	}
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, fmt.Errorf("create filesink: %w", err)
	}
	filesink.Set("location", outputPath)

	switch strings.ToLower(strings.TrimSpace(out.Codec)) {
	case "", "pcm_s16le", "wav":
		wavenc, err := gst.NewElement("wavenc")
		if err != nil {
			return nil, fmt.Errorf("create wavenc: %w", err)
		}
		return []*gst.Element{audioconvert, audioresample, caps, wavenc, filesink}, nil
	case "opus":
		opusenc, err := gst.NewElement("opusenc")
		if err != nil {
			return nil, fmt.Errorf("create opusenc: %w", err)
		}
		opusenc.Set("bitrate", 160000)
		oggmux, err := gst.NewElement("oggmux")
		if err != nil {
			return nil, fmt.Errorf("create oggmux: %w", err)
		}
		return []*gst.Element{audioconvert, audioresample, caps, opusenc, oggmux, filesink}, nil
	default:
		return nil, fmt.Errorf("unsupported gstreamer audio codec %q", out.Codec)
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

func emitRecordingEvent(sink func(media.RecordingEvent), event media.RecordingEvent) {
	if sink == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	sink(event)
}
