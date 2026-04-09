package web

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
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
		return err
	}

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
		return errors.Wrap(err, "start preview ffmpeg")
	}

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
		err := cmd.Wait()
		if err != nil && ctx.Err() != nil {
			return nil
		}
		return err
	})

	return group.Wait()
}

func computePreviewSignature(source dsl.EffectiveVideoSource) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%#v", source)))
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
