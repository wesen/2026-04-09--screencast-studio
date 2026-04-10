package web

import (
	"time"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapHealthResponse(previewLimit int) *studiov1.HealthResponse {
	return &studiov1.HealthResponse{
		Ok:           true,
		Service:      "screencast-studio",
		PreviewLimit: int32(previewLimit),
	}
}

func mapDiscoveryResponse(snapshot *discovery.Snapshot) *studiov1.DiscoveryResponse {
	if snapshot == nil {
		return &studiov1.DiscoveryResponse{}
	}
	response := &studiov1.DiscoveryResponse{
		Displays: make([]*studiov1.DisplayDescriptor, 0, len(snapshot.Displays)),
		Windows:  make([]*studiov1.WindowDescriptor, 0, len(snapshot.Windows)),
		Cameras:  make([]*studiov1.CameraDescriptor, 0, len(snapshot.Cameras)),
		Audio:    make([]*studiov1.AudioInputDescriptor, 0, len(snapshot.Audio)),
	}
	for _, display := range snapshot.Displays {
		response.Displays = append(response.Displays, &studiov1.DisplayDescriptor{
			Id:        display.ID,
			Name:      display.Name,
			Primary:   display.Primary,
			X:         int32(display.X),
			Y:         int32(display.Y),
			Width:     int32(display.Width),
			Height:    int32(display.Height),
			Connector: display.Connector,
		})
	}
	for _, window := range snapshot.Windows {
		response.Windows = append(response.Windows, &studiov1.WindowDescriptor{
			Id:     window.ID,
			Title:  window.Title,
			X:      int32(window.X),
			Y:      int32(window.Y),
			Width:  int32(window.Width),
			Height: int32(window.Height),
		})
	}
	for _, camera := range snapshot.Cameras {
		response.Cameras = append(response.Cameras, &studiov1.CameraDescriptor{
			Id:       camera.ID,
			Label:    camera.Label,
			Device:   camera.Device,
			CardName: camera.CardName,
		})
	}
	for _, input := range snapshot.Audio {
		response.Audio = append(response.Audio, &studiov1.AudioInputDescriptor{
			Id:         input.ID,
			Name:       input.Name,
			Driver:     input.Driver,
			SampleSpec: input.SampleSpec,
			State:      input.State,
		})
	}
	return response
}

func mapEffectiveConfig(cfg *dsl.EffectiveConfig) *studiov1.EffectiveConfig {
	if cfg == nil {
		return &studiov1.EffectiveConfig{}
	}
	response := &studiov1.EffectiveConfig{
		Schema:               cfg.Schema,
		SessionId:            cfg.SessionID,
		DestinationTemplates: cloneStringMap(cfg.DestinationTemplates),
		AudioMixTemplate:     cfg.AudioMixTemplate,
		AudioOutput:          mapAudioOutputSettings(cfg.AudioOutput),
		VideoSources:         make([]*studiov1.EffectiveVideoSource, 0, len(cfg.VideoSources)),
		AudioSources:         make([]*studiov1.EffectiveAudioSource, 0, len(cfg.AudioSources)),
	}
	for _, source := range cfg.VideoSources {
		response.VideoSources = append(response.VideoSources, mapEffectiveVideoSource(source))
	}
	for _, source := range cfg.AudioSources {
		response.AudioSources = append(response.AudioSources, mapEffectiveAudioSource(source))
	}
	return response
}

func mapNormalizeResponse(cfg *dsl.EffectiveConfig) *studiov1.NormalizeResponse {
	if cfg == nil {
		return &studiov1.NormalizeResponse{}
	}
	return &studiov1.NormalizeResponse{
		SessionId: cfg.SessionID,
		Warnings:  append([]string(nil), cfg.Warnings...),
		Config:    mapEffectiveConfig(cfg),
	}
}

func mapCompileResponse(plan *dsl.CompiledPlan) *studiov1.CompileResponse {
	if plan == nil {
		return &studiov1.CompileResponse{}
	}
	response := &studiov1.CompileResponse{
		SessionId: plan.SessionID,
		Warnings:  append([]string(nil), plan.Warnings...),
		Outputs:   make([]*studiov1.PlannedOutput, 0, len(plan.Outputs)),
		VideoJobs: make([]*studiov1.VideoJob, 0, len(plan.VideoJobs)),
		AudioJobs: make([]*studiov1.AudioMixJob, 0, len(plan.AudioJobs)),
	}
	for _, output := range plan.Outputs {
		response.Outputs = append(response.Outputs, mapPlannedOutput(output))
	}
	for _, job := range plan.VideoJobs {
		response.VideoJobs = append(response.VideoJobs, &studiov1.VideoJob{
			Source:     mapEffectiveVideoSource(job.Source),
			OutputPath: job.OutputPath,
		})
	}
	for _, job := range plan.AudioJobs {
		audioJob := &studiov1.AudioMixJob{
			Name:       job.Name,
			Output:     mapAudioOutputSettings(job.Output),
			OutputPath: job.OutputPath,
			Sources:    make([]*studiov1.EffectiveAudioSource, 0, len(job.Sources)),
		}
		for _, source := range job.Sources {
			audioJob.Sources = append(audioJob.Sources, mapEffectiveAudioSource(source))
		}
		response.AudioJobs = append(response.AudioJobs, audioJob)
	}
	return response
}

func mapEffectiveVideoSource(source dsl.EffectiveVideoSource) *studiov1.EffectiveVideoSource {
	return &studiov1.EffectiveVideoSource{
		Id:                  source.ID,
		Name:                source.Name,
		Type:                source.Type,
		Enabled:             source.Enabled,
		Target:              mapVideoTarget(source.Target),
		Capture:             mapVideoCaptureSettings(source.Capture),
		Output:              mapVideoOutputSettings(source.Output),
		DestinationTemplate: source.DestinationTemplate,
	}
}

func mapEffectiveAudioSource(source dsl.EffectiveAudioSource) *studiov1.EffectiveAudioSource {
	return &studiov1.EffectiveAudioSource{
		Id:       source.ID,
		Name:     source.Name,
		Enabled:  source.Enabled,
		Device:   source.Device,
		Settings: mapAudioSourceSettings(source.Settings),
	}
}

func mapVideoTarget(target dsl.VideoTarget) *studiov1.VideoTarget {
	response := &studiov1.VideoTarget{
		Display:  target.Display,
		WindowId: target.WindowID,
		Device:   target.Device,
	}
	if target.Rect != nil {
		response.Rect = &studiov1.Rect{
			X: int32(target.Rect.X),
			Y: int32(target.Rect.Y),
			W: int32(target.Rect.W),
			H: int32(target.Rect.H),
		}
	}
	return response
}

func mapVideoCaptureSettings(settings dsl.VideoCaptureSettings) *studiov1.VideoCaptureSettings {
	response := &studiov1.VideoCaptureSettings{
		Fps: int32(settings.FPS),
	}
	if settings.Cursor != nil {
		response.Cursor = settings.Cursor
	}
	if settings.FollowResize != nil {
		response.FollowResize = settings.FollowResize
	}
	if settings.Mirror != nil {
		response.Mirror = settings.Mirror
	}
	if settings.Size != "" {
		response.Size = &settings.Size
	}
	return response
}

func mapVideoOutputSettings(settings dsl.VideoOutputSettings) *studiov1.VideoOutputSettings {
	return &studiov1.VideoOutputSettings{
		Container:  settings.Container,
		VideoCodec: settings.VideoCodec,
		Quality:    int32(settings.Quality),
	}
}

func mapAudioOutputSettings(settings dsl.AudioOutputSettings) *studiov1.AudioOutputSettings {
	return &studiov1.AudioOutputSettings{
		Codec:        settings.Codec,
		SampleRateHz: int32(settings.SampleRateHz),
		Channels:     int32(settings.Channels),
	}
}

func mapAudioSourceSettings(settings dsl.AudioSourceSettings) *studiov1.AudioSourceSettings {
	return &studiov1.AudioSourceSettings{
		Gain:      settings.Gain,
		NoiseGate: settings.NoiseGate,
		Denoise:   settings.Denoise,
	}
}

func mapPlannedOutput(output dsl.PlannedOutput) *studiov1.PlannedOutput {
	return &studiov1.PlannedOutput{
		Kind:     output.Kind,
		SourceId: output.SourceID,
		Name:     output.Name,
		Path:     output.Path,
	}
}

func mapSessionEnvelope(state recordingSessionState) *studiov1.SessionEnvelope {
	return &studiov1.SessionEnvelope{
		Session: mapRecordingSession(state),
	}
}

func mapRecordingSession(state recordingSessionState) *studiov1.RecordingSession {
	response := &studiov1.RecordingSession{
		Active:     state.Active,
		SessionId:  state.SessionID,
		State:      state.State,
		Reason:     state.Reason,
		StartedAt:  formatTimestamp(state.StartedAt),
		FinishedAt: formatTimestamp(state.FinishedAt),
		Warnings:   append([]string(nil), state.Warnings...),
		Outputs:    make([]*studiov1.PlannedOutput, 0, len(state.Outputs)),
		Logs:       make([]*studiov1.ProcessLog, 0, len(state.Logs)),
	}
	for _, output := range state.Outputs {
		response.Outputs = append(response.Outputs, mapPlannedOutput(output))
	}
	for _, entry := range state.Logs {
		response.Logs = append(response.Logs, mapProcessLogEntry(entry))
	}
	return response
}

func mapProcessLogEntry(entry processLogEntry) *studiov1.ProcessLog {
	return &studiov1.ProcessLog{
		Timestamp:    formatTimestamp(entry.Timestamp),
		ProcessLabel: entry.ProcessLabel,
		Stream:       entry.Stream,
		Message:      entry.Message,
	}
}

func mapPreviewResponse(snapshot previewSnapshot) *studiov1.PreviewDescriptor {
	return &studiov1.PreviewDescriptor{
		Id:          snapshot.ID,
		SourceId:    snapshot.SourceID,
		Name:        snapshot.Name,
		SourceType:  snapshot.SourceType,
		State:       snapshot.State,
		Reason:      snapshot.Reason,
		Leases:      int32(snapshot.Leases),
		HasFrame:    snapshot.HasFrame,
		LastFrameAt: formatTimestamp(snapshot.LastFrameAt),
	}
}

func mapPreviewListResponse(previews []previewSnapshot) *studiov1.PreviewListResponse {
	response := &studiov1.PreviewListResponse{
		Previews: make([]*studiov1.PreviewDescriptor, 0, len(previews)),
	}
	for _, preview := range previews {
		response.Previews = append(response.Previews, mapPreviewResponse(preview))
	}
	return response
}

func mapServerEvent(event ServerEvent) *studiov1.ServerEvent {
	response := &studiov1.ServerEvent{}
	if !event.Timestamp.IsZero() {
		response.Timestamp = timestamppb.New(event.Timestamp)
	}

	switch event.Type {
	case "session.state":
		if payload, ok := event.Payload.(*studiov1.RecordingSession); ok {
			response.Kind = &studiov1.ServerEvent_SessionState{SessionState: payload}
		}
	case "session.log":
		if payload, ok := event.Payload.(*studiov1.ProcessLog); ok {
			response.Kind = &studiov1.ServerEvent_SessionLog{SessionLog: payload}
		}
	case "preview.list":
		if payload, ok := event.Payload.(*studiov1.PreviewListResponse); ok {
			response.Kind = &studiov1.ServerEvent_PreviewList{PreviewList: payload}
		}
	case "preview.state":
		if payload, ok := event.Payload.(*studiov1.PreviewDescriptor); ok {
			response.Kind = &studiov1.ServerEvent_PreviewState{PreviewState: payload}
		}
	case "preview.log":
		if payload, ok := event.Payload.(*studiov1.ProcessLog); ok {
			response.Kind = &studiov1.ServerEvent_PreviewLog{PreviewLog: payload}
		}
	case "telemetry.audio_meter":
		if payload, ok := event.Payload.(*studiov1.AudioMeterEvent); ok {
			response.Kind = &studiov1.ServerEvent_AudioMeter{AudioMeter: payload}
		}
	case "telemetry.disk_status":
		if payload, ok := event.Payload.(*studiov1.DiskTelemetryEvent); ok {
			response.Kind = &studiov1.ServerEvent_DiskStatus{DiskStatus: payload}
		}
	}

	return response
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
