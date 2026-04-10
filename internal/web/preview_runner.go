package web

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
)

type PreviewRunner interface {
	Run(ctx context.Context, source dsl.EffectiveVideoSource, onFrame func([]byte), onLog func(string, string)) error
}

type FFmpegPreviewRunner struct{}

func (FFmpegPreviewRunner) Run(ctx context.Context, source dsl.EffectiveVideoSource, onFrame func([]byte), onLog func(string, string)) error {
	args, err := recording.BuildPreviewArgs(source)
	if err != nil {
		log.Error().
			Str("event", "preview.process.args.error").
			Str("source_id", source.ID).
			Str("source_name", source.Name).
			Err(err).
			Msg("failed to build preview ffmpeg args")
		return err
	}

	log.Info().
		Str("event", "preview.process.start.requested").
		Str("source_id", source.ID).
		Str("source_name", source.Name).
		Str("source_type", source.Type).
		Strs("argv", append([]string{"ffmpeg"}, args...)).
		Msg("starting preview ffmpeg process")

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "open preview stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "open preview stderr")
	}
	if err := cmd.Start(); err != nil {
		log.Error().
			Str("event", "preview.process.start.error").
			Str("source_id", source.ID).
			Err(err).
			Msg("failed to start preview ffmpeg")
		return errors.Wrap(err, "start preview ffmpeg")
	}
	log.Info().
		Str("event", "preview.process.start.done").
		Str("source_id", source.ID).
		Int("pid", cmd.Process.Pid).
		Msg("preview ffmpeg started")

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return drainTextStream(groupCtx, stderr, func(line string) {
			if strings.TrimSpace(line) == "" {
				return
			}
			onLog("stderr", line)
		})
	})
	group.Go(func() error {
		reader := bufio.NewReader(stdout)
		for {
			frame, err := readJPEGFrame(reader)
			if err != nil {
				if errors.Is(err, io.EOF) || ctx.Err() != nil {
					return nil
				}
				return err
			}
			onFrame(frame)
		}
	})
	group.Go(func() error {
		log.Info().
			Str("event", "preview.process.wait.begin").
			Str("source_id", source.ID).
			Int("pid", cmd.Process.Pid).
			Msg("waiting for preview ffmpeg to exit")
		err := cmd.Wait()
		if err != nil && ctx.Err() != nil {
			log.Info().
				Str("event", "preview.process.wait.done").
				Str("source_id", source.ID).
				Int("pid", cmd.Process.Pid).
				Str("reason", ctx.Err().Error()).
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

	err = group.Wait()
	if err != nil {
		log.Error().
			Str("event", "preview.process.summary").
			Str("source_id", source.ID).
			Err(err).
			Msg("preview runner finished with error")
		return err
	}
	log.Info().
		Str("event", "preview.process.summary").
		Str("source_id", source.ID).
		Str("reason", previewContextReason(ctx)).
		Msg("preview runner finished")
	return nil
}

func computePreviewSignature(source dsl.EffectiveVideoSource) string {
	payload, err := json.Marshal(source)
	if err != nil {
		log.Error().
			Str("event", "preview.signature.marshal.error").
			Str("source_id", source.ID).
			Err(err).
			Msg("failed to marshal preview source for signature; falling back to source id")
		payload = []byte(source.ID)
	}
	sum := sha1.Sum(payload)
	return hex.EncodeToString(sum[:])
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
