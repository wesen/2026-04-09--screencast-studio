package discovery

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	monitorGeometryRe = regexp.MustCompile(`^(\d+)/\d+x(\d+)/\d+\+(-?\d+)\+(-?\d+)$`)
	windowIDRe        = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	quotedValueRe     = regexp.MustCompile(`"([^"]*)"`)
	geometryIntRe     = regexp.MustCompile(`(-?\d+)`)
)

func SnapshotAll(ctx context.Context) (*Snapshot, error) {
	displays, err := ListDisplays(ctx)
	if err != nil {
		return nil, err
	}
	windows, err := ListWindows(ctx)
	if err != nil {
		return nil, err
	}
	cameras, err := ListCameras(ctx)
	if err != nil {
		return nil, err
	}
	audio, err := ListAudioInputs(ctx)
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		Displays: displays,
		Windows:  windows,
		Cameras:  cameras,
		Audio:    audio,
	}, nil
}

func ListDisplays(ctx context.Context) ([]Display, error) {
	output, err := runOutput(ctx, "xrandr", "--listmonitors")
	if err != nil {
		return nil, errors.Wrap(err, "list displays with xrandr")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	displays := []Display{}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 || !strings.HasSuffix(fields[0], ":") {
			continue
		}

		width, height, x, y, err := parseMonitorGeometry(fields[2])
		if err != nil {
			return nil, err
		}

		displays = append(displays, Display{
			ID:        strings.TrimSuffix(fields[0], ":"),
			Name:      fields[3],
			Primary:   strings.Contains(fields[1], "*"),
			X:         x,
			Y:         y,
			Width:     width,
			Height:    height,
			Connector: fields[3],
		})
	}

	return displays, nil
}

func parseMonitorGeometry(token string) (int, int, int, int, error) {
	match := monitorGeometryRe.FindStringSubmatch(token)
	if match == nil {
		return 0, 0, 0, 0, errors.Errorf("unexpected monitor geometry token %q", token)
	}
	width, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, 0, 0, 0, errors.Wrap(err, "parse monitor width")
	}
	height, err := strconv.Atoi(match[2])
	if err != nil {
		return 0, 0, 0, 0, errors.Wrap(err, "parse monitor height")
	}
	x, err := strconv.Atoi(match[3])
	if err != nil {
		return 0, 0, 0, 0, errors.Wrap(err, "parse monitor x")
	}
	y, err := strconv.Atoi(match[4])
	if err != nil {
		return 0, 0, 0, 0, errors.Wrap(err, "parse monitor y")
	}
	return width, height, x, y, nil
}

func ListWindows(ctx context.Context) ([]Window, error) {
	output, err := runOutput(ctx, "xprop", "-root", "_NET_CLIENT_LIST")
	if err != nil {
		return nil, errors.Wrap(err, "list windows with xprop")
	}

	ids := windowIDRe.FindAllString(output, -1)
	windows := make([]Window, 0, len(ids))
	for _, id := range ids {
		title, err := windowTitle(ctx, id)
		if err != nil {
			return nil, err
		}
		x, y, width, height, err := windowGeometry(ctx, id)
		if err != nil {
			return nil, err
		}

		windows = append(windows, Window{
			ID:     id,
			Title:  title,
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		})
	}

	return windows, nil
}

func windowTitle(ctx context.Context, id string) (string, error) {
	for _, args := range [][]string{
		{"-id", id, "_NET_WM_NAME"},
		{"-id", id, "WM_NAME"},
	} {
		output, err := runOutput(ctx, "xprop", args...)
		if err != nil {
			continue
		}
		match := quotedValueRe.FindStringSubmatch(output)
		if match != nil {
			return match[1], nil
		}
	}
	return id, nil
}

func windowGeometry(ctx context.Context, id string) (int, int, int, int, error) {
	output, err := runOutput(ctx, "xwininfo", "-id", id)
	if err != nil {
		return 0, 0, 0, 0, errors.Wrapf(err, "query geometry for window %s", id)
	}

	x, err := findIntAfter(output, "Absolute upper-left X:")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	y, err := findIntAfter(output, "Absolute upper-left Y:")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	width, err := findIntAfter(output, "Width:")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	height, err := findIntAfter(output, "Height:")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return x, y, width, height, nil
}

func findIntAfter(output, prefix string) (int, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		match := geometryIntRe.FindString(line[len(prefix):])
		if match == "" {
			return 0, errors.Errorf("no integer found after %q", prefix)
		}
		value, err := strconv.Atoi(match)
		if err != nil {
			return 0, errors.Wrapf(err, "parse integer for %q", prefix)
		}
		return value, nil
	}
	return 0, errors.Errorf("prefix %q not found", prefix)
}

func ListCameras(ctx context.Context) ([]Camera, error) {
	output, err := runOutput(ctx, "v4l2-ctl", "--list-devices")
	if err != nil {
		return nil, errors.Wrap(err, "list cameras with v4l2-ctl")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	cameras := []Camera{}
	currentLabel := ""
	currentCard := ""
	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, " \t")
		if line == "" {
			continue
		}
		if !strings.HasPrefix(rawLine, "\t") && !strings.HasPrefix(rawLine, " ") {
			currentLabel = strings.TrimSuffix(strings.TrimSpace(line), ":")
			currentCard = currentLabel
			continue
		}
		device := strings.TrimSpace(line)
		if !strings.HasPrefix(device, "/dev/video") {
			continue
		}
		cameras = append(cameras, Camera{
			ID:       device,
			Label:    currentLabel,
			Device:   device,
			CardName: currentCard,
		})
	}

	return cameras, nil
}

func ListAudioInputs(ctx context.Context) ([]AudioInput, error) {
	output, err := runOutput(ctx, "pactl", "list", "short", "sources")
	if err != nil {
		return nil, errors.Wrap(err, "list audio inputs with pactl")
	}

	audio := []AudioInput{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			return nil, errors.Errorf("unexpected pactl source line %q", line)
		}
		audio = append(audio, AudioInput{
			ID:         parts[1],
			Name:       parts[1],
			Driver:     parts[2],
			SampleSpec: strings.Join(parts[3:len(parts)-1], " "),
			State:      parts[len(parts)-1],
		})
	}

	return audio, nil
}

func runOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("%s %s: %s", name, strings.Join(args, " "), strings.TrimSpace(string(output))))
	}
	return string(output), nil
}
