// TypeScript types matching Go DSL types from pkg/dsl/types.go

// ── Discovery Types ──

export interface Display {
  kind: 'display';
  id: string;
  name: string;
  primary: boolean;
  x: number;
  y: number;
  width: number;
  height: number;
  connector: string;
}

export interface Window {
  kind: 'window';
  id: string;
  title: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface Camera {
  kind: 'camera';
  id: string;
  label: string;
  device: string;
  card_name: string;
}

export interface AudioInput {
  kind: 'audio';
  id: string;
  name: string;
  driver: string;
  sample_spec: string;
  state: string;
}

export type DiscoveryItem = Display | Window | Camera | AudioInput;

export interface DiscoveryResponse {
  generated_at: string;
  items: DiscoveryItem[];
}

// ── Setup/DSL Types ──

export interface Rect {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface VideoTarget {
  display: string;
  window_id: string;
  device: string;
  rect?: Rect;
}

export interface VideoCaptureSettings {
  fps: number;
  cursor?: boolean;
  follow_resize?: boolean;
  mirror?: boolean;
  size: string;
}

export interface VideoOutputSettings {
  container: string;
  video_codec: string;
  quality: number;
}

export interface VideoDefaults {
  capture: VideoCaptureSettings;
  output: VideoOutputSettings;
}

export interface VideoSource {
  id: string;
  name: string;
  type: string;
  enabled?: boolean;
  target: VideoTarget;
  settings: VideoDefaults;
  destination_template: string;
}

export interface AudioOutputSettings {
  codec: string;
  sample_rate_hz: number;
  channels: number;
}

export interface AudioSourceSettings {
  gain: number;
  noise_gate: boolean;
  denoise: boolean;
}

export interface AudioSource {
  id: string;
  name: string;
  device: string;
  enabled?: boolean;
  settings: AudioSourceSettings;
}

export interface Config {
  schema: string;
  session_id: string;
  destination_templates: Record<string, string>;
  screen_capture_defaults: VideoDefaults;
  camera_capture_defaults: VideoDefaults;
  audio_defaults: {
    output: AudioOutputSettings;
  };
  audio_mix: {
    destination_template: string;
  };
  video_sources: VideoSource[];
  audio_sources: AudioSource[];
}

// ── Compile Types ──

export interface PlannedOutput {
  kind: 'video' | 'audio';
  source_id?: string;
  name: string;
  path: string;
}

export interface VideoJob {
  source: {
    id: string;
    name: string;
    type: string;
    enabled: boolean;
    target: VideoTarget;
    capture: VideoCaptureSettings;
    output: VideoOutputSettings;
    destination_template: string;
  };
  output_path: string;
}

export interface AudioJob {
  name: string;
  sources: {
    id: string;
    name: string;
    enabled: boolean;
    device: string;
    settings: AudioSourceSettings;
  }[];
  output: AudioOutputSettings;
  output_path: string;
}

export interface CompiledPlan {
  session_id: string;
  video_jobs: VideoJob[];
  audio_jobs: AudioJob[];
  outputs: PlannedOutput[];
  warnings: string[];
}

export interface CompileRequest {
  dsl_format: 'yaml';
  dsl: string;
}

export interface CompileResponse {
  session_id: string;
  warnings: string[];
  outputs: PlannedOutput[];
  video_jobs: number;
  audio_jobs: number;
}

export interface NormalizeResponse {
  valid: boolean;
  config?: Config;
  errors?: string[];
}

// ── Recording Types ──

export interface RecordingState {
  active: boolean;
  session_id: string;
  state: 'idle' | 'compiling' | 'starting' | 'running' | 'paused' | 'stopping' | 'stopped' | 'error';
  reason: string;
  started_at?: string;
  finished_at?: string;
  outputs: PlannedOutput[];
  warnings: string[];
}

export interface RecordingStartRequest {
  dsl_format: 'yaml';
  dsl: string;
  max_duration_seconds?: number;
}

export interface RecordingStartResponse {
  session_id: string;
  state: string;
  reason: string;
}

export interface RecordingStopResponse {
  session_id: string;
  state: string;
  reason: string;
}

// ── Preview Types ──

export interface PreviewDescriptor {
  id: string;
  source_id: string;
  state: 'starting' | 'ready' | 'error' | 'stopped';
  stream_url: string;
}

export interface PreviewEnsureRequest {
  source_id: string;
}

export interface PreviewReleaseRequest {
  source_id: string;
}

// ── WebSocket Event Types ──

export interface WsSessionStateEvent {
  type: 'session.state';
  payload: {
    session_id: string;
    state: string;
    reason: string;
  };
}

export interface WsSessionLogEvent {
  type: 'session.log';
  payload: {
    session_id: string;
    level: 'debug' | 'info' | 'warn' | 'error';
    line: string;
  };
}

export interface WsSessionOutputEvent {
  type: 'session.output';
  payload: {
    session_id: string;
    kind: string;
    path: string;
  };
}

export interface WsPreviewStateEvent {
  type: 'preview.state';
  payload: {
    source_id: string;
    state: string;
  };
}

export interface WsAudioMeterEvent {
  type: 'meter.audio';
  payload: {
    device_id: string;
    level: number;
  };
}

export type WsEvent =
  | WsSessionStateEvent
  | WsSessionLogEvent
  | WsSessionOutputEvent
  | WsPreviewStateEvent
  | WsAudioMeterEvent;
