package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
)

type PreviewRuntime struct{}

func NewPreviewRuntime() *PreviewRuntime {
	return &PreviewRuntime{}
}

func (PreviewRuntime) StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts media.PreviewOptions) (media.PreviewSession, error) {
	args, err := recording.BuildPreviewArgs(source)
	if err != nil {
		log.Error().
			Str("event", "preview.process.args.error").
			Str("source_id", source.ID).
			Str("source_name", source.Name).
			Err(err).
			Msg("failed to build preview ffmpeg args")
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	previewCtx, cancel := context.WithCancel(ctx)

	log.Info().
		Str("event", "preview.process.start.requested").
		Str("source_id", source.ID).
		Str("source_name", source.Name).
		Str("source_type", source.Type).
		Strs("argv", append([]string{"ffmpeg"}, args...)).
		Msg("starting preview ffmpeg process")

	cmd := exec.CommandContext(previewCtx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "open preview stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "open preview stderr")
	}
	if err := cmd.Start(); err != nil {
		cancel()
		log.Error().
			Str("event", "preview.process.start.error").
			Str("source_id", source.ID).
			Err(err).
			Msg("failed to start preview ffmpeg")
		return nil, errors.Wrap(err, "start preview ffmpeg")
	}
	log.Info().
		Str("event", "preview.process.start.done").
		Str("source_id", source.ID).
		Int("pid", cmd.Process.Pid).
		Msg("preview ffmpeg started")

	session := &previewSession{
		cancel: cancel,
		done:   make(chan struct{}),
	}
	if opts.OnFrame == nil {
		opts.OnFrame = func([]byte) {}
	}
	if opts.OnLog == nil {
		opts.OnLog = func(string, string) {}
	}

	group, groupCtx := errgroup.WithContext(previewCtx)
	group.Go(func() error {
		return drainTextStream(groupCtx, stderr, func(line string) {
			if strings.TrimSpace(line) == "" {
				return
			}
			opts.OnLog("stderr", line)
		})
	})
	group.Go(func() error {
		reader := bufio.NewReader(stdout)
		for {
			frame, err := readJPEGFrame(reader)
			if err != nil {
				if errors.Is(err, io.EOF) || previewCtx.Err() != nil {
					return nil
				}
				return err
			}
			session.setLatestFrame(frame)
			opts.OnFrame(frame)
		}
	})
	group.Go(func() error {
		log.Info().
			Str("event", "preview.process.wait.begin").
			Str("source_id", source.ID).
			Int("pid", cmd.Process.Pid).
			Msg("waiting for preview ffmpeg to exit")
		err := cmd.Wait()
		if err != nil && previewCtx.Err() != nil {
			log.Info().
				Str("event", "preview.process.wait.done").
				Str("source_id", source.ID).
				Int("pid", cmd.Process.Pid).
				Str("reason", previewContextReason(previewCtx)).
				Msg("preview ffmpeg exited after context cancellation")
			return nil
		}
		if err != nil {
			log.Error().
				Str("event", "preview.process.wait.done").
				Str("source_id", source.ID).
				Int("pid", cmd.Process.Pid).
				Err(err).
				Msg("preview ffmpeg exited with error")
			return err
		}
		log.Info().
			Str("event", "preview.process.wait.done").
			Str("source_id", source.ID).
			Int("pid", cmd.Process.Pid).
			Msg("preview ffmpeg exited cleanly")
		return nil
	})

	go func() {
		defer close(session.done)
		session.setWaitResult(group.Wait())
		if session.waitErr != nil {
			log.Error().
				Str("event", "preview.process.summary").
				Str("source_id", source.ID).
				Err(session.waitErr).
				Msg("preview runner finished with error")
			return
		}
		log.Info().
			Str("event", "preview.process.summary").
			Str("source_id", source.ID).
			Str("reason", previewContextReason(previewCtx)).
			Msg("preview runner finished")
	}()

	return session, nil
}

type previewSession struct {
	cancel context.CancelFunc
	done   chan struct{}

	mu          sync.RWMutex
	latestFrame []byte
	waitErr     error
}

func (s *previewSession) Wait() error {
	if s == nil {
		return nil
	}
	if s.done != nil {
		<-s.done
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.waitErr
}

func (s *previewSession) Stop(ctx context.Context) error {
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
		return s.Wait()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *previewSession) LatestFrame() ([]byte, error) {
	if s == nil {
		return nil, errors.New("preview session is nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.latestFrame) == 0 {
		return nil, errors.New("preview frame not available yet")
	}
	return append([]byte(nil), s.latestFrame...), nil
}

func (s *previewSession) TakeScreenshot(ctx context.Context, opts media.ScreenshotOptions) ([]byte, error) {
	_ = ctx
	_ = opts
	return s.LatestFrame()
}

func (s *previewSession) setLatestFrame(frame []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latestFrame = append([]byte(nil), frame...)
}

func (s *previewSession) setWaitResult(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitErr = err
}

func drainTextStream(ctx context.Context, r io.ReadCloser, fn func(string)) error {
	defer r.Close()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		fn(scanner.Text())
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

func readJPEGFrame(reader *bufio.Reader) ([]byte, error) {
	frame := &bytes.Buffer{}
	started := false
	var previous byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if !started {
			if previous == 0xff && b == 0xd8 {
				started = true
				frame.WriteByte(0xff)
				frame.WriteByte(0xd8)
			}
			previous = b
			continue
		}
		frame.WriteByte(b)
		if previous == 0xff && b == 0xd9 {
			return frame.Bytes(), nil
		}
		previous = b
	}
}

func previewContextReason(ctx context.Context) string {
	if ctx == nil || ctx.Err() == nil {
		return "running"
	}
	return ctx.Err().Error()
}
