package web

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type telemetryTarget struct {
	AudioDevice string
	DiskPath    string
}

type audioMeterSnapshot struct {
	DeviceID   string
	LeftLevel  float64
	RightLevel float64
	Available  bool
	Reason     string
}

type diskTelemetrySnapshot struct {
	Path        string
	Filesystem  string
	UsedPercent float64
	FreeGiB     float64
	TotalGiB    float64
	LowSpace    bool
	Available   bool
	Reason      string
}

type TelemetryManager struct {
	publish func(ServerEvent)

	mu         sync.RWMutex
	target     telemetryTarget
	audioMeter audioMeterSnapshot
	diskStatus diskTelemetrySnapshot
}

func NewTelemetryManager(publish func(ServerEvent)) *TelemetryManager {
	if publish == nil {
		publish = func(ServerEvent) {}
	}
	return &TelemetryManager{
		publish: publish,
		audioMeter: audioMeterSnapshot{
			Available: false,
			Reason:    "telemetry target unavailable",
		},
		diskStatus: diskTelemetrySnapshot{
			Available: false,
			Reason:    "telemetry target unavailable",
		},
	}
}

func (m *TelemetryManager) Run(ctx context.Context) error {
	log.Info().
		Str("event", "telemetry.run.start").
		Msg("telemetry manager run loop starting")
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return m.runDiskLoop(groupCtx) })
	group.Go(func() error { return m.runAudioLoop(groupCtx) })
	err := group.Wait()
	if err != nil {
		log.Error().
			Str("event", "telemetry.run.exit").
			Err(err).
			Msg("telemetry manager run loop exited with error")
		return err
	}
	log.Info().
		Str("event", "telemetry.run.exit").
		Str("reason", telemetryContextReason(ctx)).
		Msg("telemetry manager run loop exited")
	return nil
}

func (m *TelemetryManager) UpdateFromPlan(plan *dsl.CompiledPlan) {
	if plan == nil {
		return
	}
	target := telemetryTarget{}
	if len(plan.Outputs) > 0 {
		target.DiskPath = filepath.Dir(plan.Outputs[0].Path)
	}
	if len(plan.AudioJobs) > 0 && len(plan.AudioJobs[0].Sources) > 0 {
		target.AudioDevice = strings.TrimSpace(plan.AudioJobs[0].Sources[0].Device)
	}

	m.mu.Lock()
	m.target = target
	m.mu.Unlock()
}

func (m *TelemetryManager) AudioMeter() audioMeterSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.audioMeter
}

func (m *TelemetryManager) DiskStatus() diskTelemetrySnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.diskStatus
}

func (m *TelemetryManager) currentTarget() telemetryTarget {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.target
}

func (m *TelemetryManager) setAudioMeter(snapshot audioMeterSnapshot) {
	m.mu.Lock()
	m.audioMeter = snapshot
	m.mu.Unlock()
	m.publish(ServerEvent{
		Type:      "telemetry.audio_meter",
		Timestamp: time.Now(),
		Payload:   mapAudioMeterSnapshot(snapshot),
	})
}

func (m *TelemetryManager) setDiskStatus(snapshot diskTelemetrySnapshot) {
	m.mu.Lock()
	m.diskStatus = snapshot
	m.mu.Unlock()
	m.publish(ServerEvent{
		Type:      "telemetry.disk_status",
		Timestamp: time.Now(),
		Payload:   mapDiskTelemetrySnapshot(snapshot),
	})
}

func (m *TelemetryManager) runDiskLoop(ctx context.Context) error {
	log.Info().
		Str("event", "telemetry.disk.start").
		Dur("interval", 2*time.Second).
		Msg("disk telemetry loop starting")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	publish := func() {
		target := m.currentTarget()
		path := target.DiskPath
		if strings.TrimSpace(path) == "" {
			path = "."
		}
		snapshot, err := collectDiskTelemetry(path)
		if err != nil {
			snapshot = diskTelemetrySnapshot{
				Path:      path,
				Available: false,
				Reason:    err.Error(),
			}
		}
		m.setDiskStatus(snapshot)
	}

	publish()
	for {
		select {
		case <-ctx.Done():
			log.Info().
				Str("event", "telemetry.disk.exit").
				Str("reason", telemetryContextReason(ctx)).
				Msg("disk telemetry loop exiting")
			return nil
		case <-ticker.C:
			publish()
		}
	}
}

func (m *TelemetryManager) runAudioLoop(ctx context.Context) error {
	log.Info().
		Str("event", "telemetry.audio.start").
		Dur("interval", 500*time.Millisecond).
		Msg("audio telemetry loop starting")
	var (
		currentDevice string
		runnerCancel  context.CancelFunc
		runnerDone    chan struct{}
	)
	defer func() {
		if runnerCancel != nil {
			log.Info().
				Str("event", "telemetry.audio.runner.cancel").
				Str("device_id", currentDevice).
				Msg("cancelling audio telemetry runner during loop teardown")
			runnerCancel()
		}
		if runnerDone != nil {
			<-runnerDone
		}
		log.Info().
			Str("event", "telemetry.audio.exit").
			Str("reason", telemetryContextReason(ctx)).
			Msg("audio telemetry loop exiting")
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	refreshRunner := func(target telemetryTarget) {
		device := strings.TrimSpace(target.AudioDevice)
		if device == currentDevice {
			return
		}
		if runnerCancel != nil {
			log.Info().
				Str("event", "telemetry.audio.runner.replace").
				Str("device_id", currentDevice).
				Msg("replacing existing audio telemetry runner")
			runnerCancel()
			<-runnerDone
			runnerCancel = nil
			runnerDone = nil
		}
		currentDevice = device
		if device == "" {
			log.Info().
				Str("event", "telemetry.audio.runner.idle").
				Msg("no audio device selected for telemetry")
			m.setAudioMeter(audioMeterSnapshot{
				DeviceID:  "",
				Available: false,
				Reason:    "no audio input selected",
			})
			return
		}
		log.Info().
			Str("event", "telemetry.audio.runner.start").
			Str("device_id", device).
			Msg("starting audio telemetry runner")
		runnerCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		runnerCancel = cancel
		runnerDone = done
		go func(deviceID string) {
			defer close(done)
			if err := m.streamAudioMeter(runnerCtx, deviceID); err != nil && runnerCtx.Err() == nil {
				log.Error().
					Str("event", "telemetry.audio.runner.error").
					Str("device_id", deviceID).
					Err(err).
					Msg("audio telemetry runner failed")
				m.setAudioMeter(audioMeterSnapshot{
					DeviceID:  deviceID,
					Available: false,
					Reason:    err.Error(),
				})
				return
			}
			log.Info().
				Str("event", "telemetry.audio.runner.exit").
				Str("device_id", deviceID).
				Str("reason", telemetryContextReason(runnerCtx)).
				Msg("audio telemetry runner exited")
		}(device)
	}

	refreshRunner(m.currentTarget())
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			refreshRunner(m.currentTarget())
		}
	}
}

func (m *TelemetryManager) streamAudioMeter(ctx context.Context, deviceID string) error {
	log.Info().
		Str("event", "telemetry.process.start.requested").
		Str("device_id", deviceID).
		Strs("argv", []string{"parec", "--device=" + deviceID, "--format=s16le", "--rate=16000", "--channels=2", "--raw"}).
		Msg("starting parec audio telemetry process")
	cmd := exec.CommandContext(ctx, "parec",
		"--device="+deviceID,
		"--format=s16le",
		"--rate=16000",
		"--channels=2",
		"--raw",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "open parec stdout")
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		log.Error().
			Str("event", "telemetry.process.start.error").
			Str("device_id", deviceID).
			Err(err).
			Msg("failed to start parec audio telemetry process")
		return errors.Wrap(err, "start parec")
	}
	log.Info().
		Str("event", "telemetry.process.start.done").
		Str("device_id", deviceID).
		Int("pid", cmd.Process.Pid).
		Msg("parec audio telemetry process started")
	defer func() {
		log.Info().
			Str("event", "telemetry.process.wait.begin").
			Str("device_id", deviceID).
			Int("pid", cmd.Process.Pid).
			Msg("waiting for parec audio telemetry process to exit")
		if waitErr := cmd.Wait(); waitErr != nil && ctx.Err() == nil {
			log.Error().
				Str("event", "telemetry.process.wait.done").
				Str("device_id", deviceID).
				Int("pid", cmd.Process.Pid).
				Err(waitErr).
				Msg("parec audio telemetry process exited with error")
			return
		}
		log.Info().
			Str("event", "telemetry.process.wait.done").
			Str("device_id", deviceID).
			Int("pid", cmd.Process.Pid).
			Str("reason", telemetryContextReason(ctx)).
			Msg("parec audio telemetry process exited")
	}()

	buffer := make([]byte, 4096)
	lastPublish := time.Now()
	var leftPeak float64
	var rightPeak float64

	for {
		n, readErr := stdout.Read(buffer)
		if n > 0 {
			l, r := peakLevels(buffer[:n])
			leftPeak = math.Max(leftPeak, l)
			rightPeak = math.Max(rightPeak, r)
			if time.Since(lastPublish) >= 150*time.Millisecond {
				m.setAudioMeter(audioMeterSnapshot{
					DeviceID:   deviceID,
					LeftLevel:  leftPeak,
					RightLevel: rightPeak,
					Available:  true,
				})
				leftPeak = 0
				rightPeak = 0
				lastPublish = time.Now()
			}
		}
		if readErr != nil {
			if ctx.Err() != nil {
				log.Info().
					Str("event", "telemetry.process.read.exit").
					Str("device_id", deviceID).
					Str("reason", telemetryContextReason(ctx)).
					Msg("parec read loop exiting after context cancellation")
				return nil
			}
			if readErr == io.EOF {
				return errors.Errorf("parec ended for %s: %s", deviceID, strings.TrimSpace(stderr.String()))
			}
			return errors.Wrap(readErr, "read parec stream")
		}
	}
}

func telemetryContextReason(ctx context.Context) string {
	if ctx == nil || ctx.Err() == nil {
		return "running"
	}
	return ctx.Err().Error()
}

func peakLevels(buffer []byte) (float64, float64) {
	var leftPeak float64
	var rightPeak float64
	for i := 0; i+3 < len(buffer); i += 4 {
		left := math.Abs(float64(int16(binary.LittleEndian.Uint16(buffer[i:i+2])))) / 32768.0
		right := math.Abs(float64(int16(binary.LittleEndian.Uint16(buffer[i+2:i+4])))) / 32768.0
		if left > leftPeak {
			leftPeak = left
		}
		if right > rightPeak {
			rightPeak = right
		}
	}
	return leftPeak, rightPeak
}

func collectDiskTelemetry(path string) (diskTelemetrySnapshot, error) {
	resolved := existingPathOrParent(path)
	var stat syscall.Statfs_t
	if err := syscall.Statfs(resolved, &stat); err != nil {
		return diskTelemetrySnapshot{}, errors.Wrap(err, "statfs")
	}

	totalBytes := float64(stat.Blocks) * float64(stat.Bsize)
	freeBytes := float64(stat.Bavail) * float64(stat.Bsize)
	usedBytes := totalBytes - freeBytes
	usedPercent := 0.0
	if totalBytes > 0 {
		usedPercent = (usedBytes / totalBytes) * 100
	}
	totalGiB := totalBytes / (1024 * 1024 * 1024)
	freeGiB := freeBytes / (1024 * 1024 * 1024)

	return diskTelemetrySnapshot{
		Path:        path,
		Filesystem:  resolved,
		UsedPercent: usedPercent,
		FreeGiB:     freeGiB,
		TotalGiB:    totalGiB,
		LowSpace:    freeGiB < 5 || usedPercent >= 90,
		Available:   true,
	}, nil
}

func existingPathOrParent(path string) string {
	current := filepath.Clean(path)
	for {
		if current == "" || current == "." {
			return "."
		}
		if _, err := os.Stat(current); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "."
		}
		current = parent
	}
}

func mapAudioMeterSnapshot(snapshot audioMeterSnapshot) *studiov1.AudioMeterEvent {
	return &studiov1.AudioMeterEvent{
		DeviceId:   snapshot.DeviceID,
		LeftLevel:  snapshot.LeftLevel,
		RightLevel: snapshot.RightLevel,
		Available:  snapshot.Available,
		Reason:     snapshot.Reason,
	}
}

func mapDiskTelemetrySnapshot(snapshot diskTelemetrySnapshot) *studiov1.DiskTelemetryEvent {
	return &studiov1.DiskTelemetryEvent{
		Path:        snapshot.Path,
		Filesystem:  snapshot.Filesystem,
		UsedPercent: snapshot.UsedPercent,
		FreeGib:     snapshot.FreeGiB,
		TotalGib:    snapshot.TotalGiB,
		LowSpace:    snapshot.LowSpace,
		Available:   snapshot.Available,
		Reason:      snapshot.Reason,
	}
}
