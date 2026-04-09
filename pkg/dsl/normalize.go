package dsl

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const SchemaVersion = "recorder.config/v1"

func ParseAndNormalize(body []byte) (*EffectiveConfig, error) {
	var cfg Config
	if err := decodeDSL(body, &cfg); err != nil {
		return nil, fmt.Errorf("invalid DSL: %w", err)
	}
	if cfg.Schema == "" {
		cfg.Schema = SchemaVersion
	}
	if cfg.Schema != SchemaVersion {
		return nil, fmt.Errorf("unsupported schema %q", cfg.Schema)
	}
	if len(cfg.DestinationTemplates) == 0 {
		return nil, errors.New("destination_templates is required")
	}
	if len(cfg.VideoSources) == 0 && len(cfg.AudioSources) == 0 {
		return nil, errors.New("at least one video or audio source is required")
	}

	eff := &EffectiveConfig{
		Schema:               cfg.Schema,
		SessionID:            cfg.SessionID,
		DestinationTemplates: map[string]string{},
		RawDSL:               string(body),
	}
	if eff.SessionID == "" {
		eff.SessionID = "session-" + time.Now().Format("20060102-150405")
	}
	for k, v := range cfg.DestinationTemplates {
		eff.DestinationTemplates[k] = v
	}

	eff.AudioOutput = cfg.AudioDefaults.Output
	if eff.AudioOutput.Codec == "" {
		eff.AudioOutput.Codec = "pcm_s16le"
	}
	if eff.AudioOutput.SampleRateHz <= 0 {
		eff.AudioOutput.SampleRateHz = 48000
	}
	if eff.AudioOutput.Channels <= 0 {
		eff.AudioOutput.Channels = 2
	}

	if cfg.AudioMix.DestinationTemplate != "" {
		if _, ok := eff.DestinationTemplates[cfg.AudioMix.DestinationTemplate]; !ok {
			return nil, fmt.Errorf("audio_mix.destination_template %q not found", cfg.AudioMix.DestinationTemplate)
		}
		eff.AudioMixTemplate = cfg.AudioMix.DestinationTemplate
	} else if _, ok := eff.DestinationTemplates["audio_mix"]; ok {
		eff.AudioMixTemplate = "audio_mix"
	}

	seen := map[string]struct{}{}
	for i, src := range cfg.VideoSources {
		v, warnings, err := normalizeVideoSource(src, i, cfg.ScreenCaptureDefaults, cfg.CameraCaptureDefaults, eff.DestinationTemplates)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[v.ID]; ok {
			return nil, fmt.Errorf("duplicate source id %q", v.ID)
		}
		seen[v.ID] = struct{}{}
		eff.VideoSources = append(eff.VideoSources, v)
		eff.Warnings = append(eff.Warnings, warnings...)
	}
	for i, src := range cfg.AudioSources {
		v, warnings, err := normalizeAudioSource(src, i)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[v.ID]; ok {
			return nil, fmt.Errorf("duplicate source id %q", v.ID)
		}
		seen[v.ID] = struct{}{}
		eff.AudioSources = append(eff.AudioSources, v)
		eff.Warnings = append(eff.Warnings, warnings...)
	}

	sort.Strings(eff.Warnings)
	return eff, nil
}

func normalizeVideoSource(src VideoSource, idx int, screenDefaults, cameraDefaults VideoDefaults, templates map[string]string) (EffectiveVideoSource, []string, error) {
	warnings := []string{}
	eff := EffectiveVideoSource{
		ID:      src.ID,
		Name:    src.Name,
		Type:    strings.ToLower(strings.TrimSpace(src.Type)),
		Enabled: boolValue(src.Enabled, true),
		Target:  src.Target,
	}
	if eff.ID == "" {
		base := src.Name
		if base == "" {
			base = fmt.Sprintf("source-%d", idx+1)
		}
		eff.ID = slugify(base)
	}
	if eff.Name == "" {
		eff.Name = strings.ReplaceAll(eff.ID, "_", " ")
	}

	var defaults VideoDefaults
	switch eff.Type {
	case "display", "window", "region":
		defaults = screenDefaults
	case "camera":
		defaults = cameraDefaults
	default:
		return eff, nil, fmt.Errorf("video_sources[%d]: unsupported type %q", idx, src.Type)
	}

	eff.Capture = defaults.Capture
	mergeVideoCapture(&eff.Capture, src.Settings.Capture)
	eff.Output = defaults.Output
	mergeVideoOutput(&eff.Output, src.Settings.Output)
	if eff.Capture.FPS <= 0 {
		eff.Capture.FPS = 24
	}
	if eff.Output.Container == "" {
		eff.Output.Container = "mov"
	}
	if eff.Output.VideoCodec == "" {
		eff.Output.VideoCodec = "h264"
	}
	if eff.Output.Quality <= 0 {
		eff.Output.Quality = 75
	}

	eff.DestinationTemplate = src.DestinationTemplate
	if eff.DestinationTemplate == "" {
		return eff, nil, fmt.Errorf("video_sources[%s]: destination_template is required", eff.ID)
	}
	if _, ok := templates[eff.DestinationTemplate]; !ok {
		return eff, nil, fmt.Errorf("video_sources[%s]: destination_template %q not found", eff.ID, eff.DestinationTemplate)
	}

	if eff.Capture.FollowResize != nil && *eff.Capture.FollowResize {
		warnings = append(warnings, fmt.Sprintf("video source %s: follow_resize is not implemented in this runtime yet", eff.ID))
	}

	switch eff.Type {
	case "display":
		if eff.Target.Display == "" {
			eff.Target.Display = defaultDisplay()
		}
	case "region":
		if eff.Target.Display == "" {
			eff.Target.Display = defaultDisplay()
		}
		if eff.Target.Rect == nil || eff.Target.Rect.W <= 0 || eff.Target.Rect.H <= 0 {
			return eff, nil, fmt.Errorf("video_sources[%s]: region sources require target.rect with positive w/h", eff.ID)
		}
	case "window":
		if eff.Target.Display == "" {
			eff.Target.Display = defaultDisplay()
		}
		if strings.TrimSpace(eff.Target.WindowID) == "" {
			return eff, nil, fmt.Errorf("video_sources[%s]: window sources require target.window_id", eff.ID)
		}
	case "camera":
		if strings.TrimSpace(eff.Target.Device) == "" {
			return eff, nil, fmt.Errorf("video_sources[%s]: camera sources require target.device", eff.ID)
		}
	}

	return eff, warnings, nil
}

func normalizeAudioSource(src AudioSource, idx int) (EffectiveAudioSource, []string, error) {
	warnings := []string{}
	eff := EffectiveAudioSource{
		ID:       src.ID,
		Name:     src.Name,
		Enabled:  boolValue(src.Enabled, true),
		Device:   strings.TrimSpace(src.Device),
		Settings: src.Settings,
	}
	if eff.ID == "" {
		base := src.Name
		if base == "" {
			base = fmt.Sprintf("audio-%d", idx+1)
		}
		eff.ID = slugify(base)
	}
	if eff.Name == "" {
		eff.Name = strings.ReplaceAll(eff.ID, "_", " ")
	}
	if eff.Device == "" {
		return eff, nil, fmt.Errorf("audio_sources[%s]: device is required", eff.ID)
	}
	if eff.Settings.Gain == 0 {
		eff.Settings.Gain = 1.0
	}
	if eff.Settings.NoiseGate {
		warnings = append(warnings, fmt.Sprintf("audio source %s: noise_gate is not implemented in this runtime yet", eff.ID))
	}
	if eff.Settings.Denoise {
		warnings = append(warnings, fmt.Sprintf("audio source %s: denoise is not implemented in this runtime yet", eff.ID))
	}
	return eff, warnings, nil
}

func mergeVideoCapture(dst *VideoCaptureSettings, src VideoCaptureSettings) {
	if src.FPS > 0 {
		dst.FPS = src.FPS
	}
	if src.Cursor != nil {
		dst.Cursor = src.Cursor
	}
	if src.FollowResize != nil {
		dst.FollowResize = src.FollowResize
	}
	if src.Mirror != nil {
		dst.Mirror = src.Mirror
	}
	if src.Size != "" {
		dst.Size = src.Size
	}
}

func mergeVideoOutput(dst *VideoOutputSettings, src VideoOutputSettings) {
	if src.Container != "" {
		dst.Container = src.Container
	}
	if src.VideoCodec != "" {
		dst.VideoCodec = src.VideoCodec
	}
	if src.Quality > 0 {
		dst.Quality = src.Quality
	}
}
