package web

import (
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type apiErrorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiHealthResponse struct {
	OK           bool   `json:"ok"`
	Service      string `json:"service"`
	PreviewLimit int    `json:"preview_limit"`
}

type apiDiscoveryResponse struct {
	Displays []apiDisplay    `json:"displays"`
	Windows  []apiWindow     `json:"windows"`
	Cameras  []apiCamera     `json:"cameras"`
	Audio    []apiAudioInput `json:"audio"`
}

type apiDisplay struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Primary   bool   `json:"primary"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Connector string `json:"connector"`
}

type apiWindow struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type apiCamera struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Device   string `json:"device"`
	CardName string `json:"card_name"`
}

type apiAudioInput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	SampleSpec string `json:"sample_spec"`
	State      string `json:"state"`
}

type apiDSLRequest struct {
	DSL string `json:"dsl"`
}

type apiNormalizeResponse struct {
	SessionID string             `json:"session_id"`
	Warnings  []string           `json:"warnings"`
	Config    apiEffectiveConfig `json:"config"`
}

type apiCompileResponse struct {
	SessionID string              `json:"session_id"`
	Warnings  []string            `json:"warnings"`
	Outputs   []apiPlannedOutput  `json:"outputs"`
	VideoJobs []apiVideoJob       `json:"video_jobs"`
	AudioJobs []apiAudioMixJob    `json:"audio_jobs"`
}

type apiEffectiveConfig struct {
	Schema               string                        `json:"schema"`
	SessionID            string                        `json:"session_id"`
	DestinationTemplates map[string]string             `json:"destination_templates"`
	AudioMixTemplate     string                        `json:"audio_mix_template"`
	AudioOutput          apiAudioOutputSettings        `json:"audio_output"`
	VideoSources         []apiEffectiveVideoSource     `json:"video_sources"`
	AudioSources         []apiEffectiveAudioSource     `json:"audio_sources"`
}

type apiEffectiveVideoSource struct {
	ID                  string                  `json:"id"`
	Name                string                  `json:"name"`
	Type                string                  `json:"type"`
	Enabled             bool                    `json:"enabled"`
	Target              apiVideoTarget          `json:"target"`
	Capture             apiVideoCaptureSettings `json:"capture"`
	Output              apiVideoOutputSettings  `json:"output"`
	DestinationTemplate string                  `json:"destination_template"`
}

type apiEffectiveAudioSource struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Device   string                 `json:"device"`
	Settings apiAudioSourceSettings `json:"settings"`
}

type apiVideoTarget struct {
	Display  string   `json:"display"`
	WindowID string   `json:"window_id"`
	Device   string   `json:"device"`
	Rect     *apiRect `json:"rect,omitempty"`
}

type apiRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type apiVideoCaptureSettings struct {
	FPS          int    `json:"fps"`
	Cursor       *bool  `json:"cursor,omitempty"`
	FollowResize *bool  `json:"follow_resize,omitempty"`
	Mirror       *bool  `json:"mirror,omitempty"`
	Size         string `json:"size,omitempty"`
}

type apiVideoOutputSettings struct {
	Container  string `json:"container"`
	VideoCodec string `json:"video_codec"`
	Quality    int    `json:"quality"`
}

type apiAudioOutputSettings struct {
	Codec        string `json:"codec"`
	SampleRateHz int    `json:"sample_rate_hz"`
	Channels     int    `json:"channels"`
}

type apiAudioSourceSettings struct {
	Gain      float64 `json:"gain"`
	NoiseGate bool    `json:"noise_gate"`
	Denoise   bool    `json:"denoise"`
}

type apiVideoJob struct {
	Source     apiEffectiveVideoSource `json:"source"`
	OutputPath string                  `json:"output_path"`
}

type apiAudioMixJob struct {
	Name       string                   `json:"name"`
	Sources    []apiEffectiveAudioSource `json:"sources"`
	Output     apiAudioOutputSettings    `json:"output"`
	OutputPath string                    `json:"output_path"`
}

type apiPlannedOutput struct {
	Kind     string `json:"kind"`
	SourceID string `json:"source_id,omitempty"`
	Name     string `json:"name"`
	Path     string `json:"path"`
}

type apiRecordingStartRequest struct {
	DSL                string `json:"dsl"`
	MaxDurationSeconds int    `json:"max_duration_seconds,omitempty"`
	GracePeriodSeconds int    `json:"grace_period_seconds,omitempty"`
}

type apiRecordingSessionResponse struct {
	Active     bool              `json:"active"`
	SessionID  string            `json:"session_id,omitempty"`
	State      string            `json:"state,omitempty"`
	Reason     string            `json:"reason,omitempty"`
	StartedAt  string            `json:"started_at,omitempty"`
	FinishedAt string            `json:"finished_at,omitempty"`
	Warnings   []string          `json:"warnings"`
	Outputs    []apiPlannedOutput `json:"outputs"`
	Logs       []apiProcessLog   `json:"logs"`
}

type apiSessionEnvelope struct {
	Session apiRecordingSessionResponse `json:"session"`
}

type apiPreviewEnsureRequest struct {
	DSL      string `json:"dsl"`
	SourceID string `json:"source_id"`
}

type apiPreviewReleaseRequest struct {
	PreviewID string `json:"preview_id"`
}

type apiPreviewResponse struct {
	ID         string `json:"id"`
	SourceID   string `json:"source_id"`
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	State      string `json:"state"`
	Reason     string `json:"reason,omitempty"`
	Leases     int    `json:"leases"`
	HasFrame   bool   `json:"has_frame"`
	LastFrameAt string `json:"last_frame_at,omitempty"`
}

type apiPreviewEnvelope struct {
	Preview apiPreviewResponse `json:"preview"`
}

type apiPreviewListResponse struct {
	Previews []apiPreviewResponse `json:"previews"`
}

type apiProcessLog struct {
	Timestamp    string `json:"timestamp"`
	ProcessLabel string `json:"process_label"`
	Stream       string `json:"stream"`
	Message      string `json:"message"`
}

func mapDiscoveryResponse(snapshot *discovery.Snapshot) apiDiscoveryResponse {
	if snapshot == nil {
		return apiDiscoveryResponse{}
	}
	response := apiDiscoveryResponse{
		Displays: make([]apiDisplay, 0, len(snapshot.Displays)),
		Windows:  make([]apiWindow, 0, len(snapshot.Windows)),
		Cameras:  make([]apiCamera, 0, len(snapshot.Cameras)),
		Audio:    make([]apiAudioInput, 0, len(snapshot.Audio)),
	}
	for _, display := range snapshot.Displays {
		response.Displays = append(response.Displays, apiDisplay(display))
	}
	for _, window := range snapshot.Windows {
		response.Windows = append(response.Windows, apiWindow(window))
	}
	for _, camera := range snapshot.Cameras {
		response.Cameras = append(response.Cameras, apiCamera(camera))
	}
	for _, input := range snapshot.Audio {
		response.Audio = append(response.Audio, apiAudioInput(input))
	}
	return response
}

func mapEffectiveConfig(cfg *dsl.EffectiveConfig) apiEffectiveConfig {
	if cfg == nil {
		return apiEffectiveConfig{}
	}
	response := apiEffectiveConfig{
		Schema:               cfg.Schema,
		SessionID:            cfg.SessionID,
		DestinationTemplates: cloneStringMap(cfg.DestinationTemplates),
		AudioMixTemplate:     cfg.AudioMixTemplate,
		AudioOutput:          mapAudioOutputSettings(cfg.AudioOutput),
		VideoSources:         make([]apiEffectiveVideoSource, 0, len(cfg.VideoSources)),
		AudioSources:         make([]apiEffectiveAudioSource, 0, len(cfg.AudioSources)),
	}
	for _, source := range cfg.VideoSources {
		response.VideoSources = append(response.VideoSources, mapEffectiveVideoSource(source))
	}
	for _, source := range cfg.AudioSources {
		response.AudioSources = append(response.AudioSources, mapEffectiveAudioSource(source))
	}
	return response
}

func mapCompileResponse(plan *dsl.CompiledPlan) apiCompileResponse {
	if plan == nil {
		return apiCompileResponse{}
	}
	response := apiCompileResponse{
		SessionID: plan.SessionID,
		Warnings:  append([]string(nil), plan.Warnings...),
		Outputs:   make([]apiPlannedOutput, 0, len(plan.Outputs)),
		VideoJobs: make([]apiVideoJob, 0, len(plan.VideoJobs)),
		AudioJobs: make([]apiAudioMixJob, 0, len(plan.AudioJobs)),
	}
	for _, output := range plan.Outputs {
		response.Outputs = append(response.Outputs, mapPlannedOutput(output))
	}
	for _, job := range plan.VideoJobs {
		response.VideoJobs = append(response.VideoJobs, apiVideoJob{
			Source:     mapEffectiveVideoSource(job.Source),
			OutputPath: job.OutputPath,
		})
	}
	for _, job := range plan.AudioJobs {
		audioJob := apiAudioMixJob{
			Name:       job.Name,
			Output:     mapAudioOutputSettings(job.Output),
			OutputPath: job.OutputPath,
			Sources:    make([]apiEffectiveAudioSource, 0, len(job.Sources)),
		}
		for _, source := range job.Sources {
			audioJob.Sources = append(audioJob.Sources, mapEffectiveAudioSource(source))
		}
		response.AudioJobs = append(response.AudioJobs, audioJob)
	}
	return response
}

func mapEffectiveVideoSource(source dsl.EffectiveVideoSource) apiEffectiveVideoSource {
	return apiEffectiveVideoSource{
		ID:                  source.ID,
		Name:                source.Name,
		Type:                source.Type,
		Enabled:             source.Enabled,
		Target:              mapVideoTarget(source.Target),
		Capture:             mapVideoCaptureSettings(source.Capture),
		Output:              mapVideoOutputSettings(source.Output),
		DestinationTemplate: source.DestinationTemplate,
	}
}

func mapEffectiveAudioSource(source dsl.EffectiveAudioSource) apiEffectiveAudioSource {
	return apiEffectiveAudioSource{
		ID:       source.ID,
		Name:     source.Name,
		Enabled:  source.Enabled,
		Device:   source.Device,
		Settings: apiAudioSourceSettings(source.Settings),
	}
}

func mapVideoTarget(target dsl.VideoTarget) apiVideoTarget {
	response := apiVideoTarget{
		Display:  target.Display,
		WindowID: target.WindowID,
		Device:   target.Device,
	}
	if target.Rect != nil {
		response.Rect = &apiRect{
			X: target.Rect.X,
			Y: target.Rect.Y,
			W: target.Rect.W,
			H: target.Rect.H,
		}
	}
	return response
}

func mapVideoCaptureSettings(settings dsl.VideoCaptureSettings) apiVideoCaptureSettings {
	return apiVideoCaptureSettings(settings)
}

func mapVideoOutputSettings(settings dsl.VideoOutputSettings) apiVideoOutputSettings {
	return apiVideoOutputSettings(settings)
}

func mapAudioOutputSettings(settings dsl.AudioOutputSettings) apiAudioOutputSettings {
	return apiAudioOutputSettings(settings)
}

func mapPlannedOutput(output dsl.PlannedOutput) apiPlannedOutput {
	return apiPlannedOutput(output)
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func formatTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.Format(time.RFC3339)
}

func mapRecordingSessionResponse(state recordingSessionState) apiRecordingSessionResponse {
	response := apiRecordingSessionResponse{
		Active:     state.Active,
		SessionID:  state.SessionID,
		State:      state.State,
		Reason:     state.Reason,
		StartedAt:  formatTimestamp(state.StartedAt),
		FinishedAt: formatTimestamp(state.FinishedAt),
		Warnings:   append([]string(nil), state.Warnings...),
		Outputs:    make([]apiPlannedOutput, 0, len(state.Outputs)),
		Logs:       make([]apiProcessLog, 0, len(state.Logs)),
	}
	for _, output := range state.Outputs {
		response.Outputs = append(response.Outputs, mapPlannedOutput(output))
	}
	for _, entry := range state.Logs {
		response.Logs = append(response.Logs, apiProcessLog{
			Timestamp:    formatTimestamp(entry.Timestamp),
			ProcessLabel: entry.ProcessLabel,
			Stream:       entry.Stream,
			Message:      entry.Message,
		})
	}
	return response
}

func mapPreviewResponse(snapshot previewSnapshot) apiPreviewResponse {
	return apiPreviewResponse{
		ID:          snapshot.ID,
		SourceID:    snapshot.SourceID,
		Name:        snapshot.Name,
		SourceType:  snapshot.SourceType,
		State:       snapshot.State,
		Reason:      snapshot.Reason,
		Leases:      snapshot.Leases,
		HasFrame:    snapshot.HasFrame,
		LastFrameAt: formatTimestamp(snapshot.LastFrameAt),
	}
}
