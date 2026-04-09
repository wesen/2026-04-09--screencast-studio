export interface ApiError {
  code: string;
  message: string;
}

export interface ApiErrorResponse {
  error: ApiError;
}

export interface HealthResponse {
  ok: boolean;
  service: string;
  preview_limit: number;
}

export interface DisplayDescriptor {
  id: string;
  name: string;
  primary: boolean;
  x: number;
  y: number;
  width: number;
  height: number;
  connector: string;
}

export interface WindowDescriptor {
  id: string;
  title: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface CameraDescriptor {
  id: string;
  label: string;
  device: string;
  card_name: string;
}

export interface AudioInputDescriptor {
  id: string;
  name: string;
  driver: string;
  sample_spec: string;
  state: string;
}

export interface DiscoveryResponse {
  displays: DisplayDescriptor[];
  windows: WindowDescriptor[];
  cameras: CameraDescriptor[];
  audio: AudioInputDescriptor[];
}

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
  size?: string;
}

export interface VideoOutputSettings {
  container: string;
  video_codec: string;
  quality: number;
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

export interface EffectiveVideoSource {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
  target: VideoTarget;
  capture: VideoCaptureSettings;
  output: VideoOutputSettings;
  destination_template: string;
}

export interface EffectiveAudioSource {
  id: string;
  name: string;
  enabled: boolean;
  device: string;
  settings: AudioSourceSettings;
}

export interface EffectiveConfig {
  schema: string;
  session_id: string;
  destination_templates: Record<string, string>;
  audio_mix_template: string;
  audio_output: AudioOutputSettings;
  video_sources: EffectiveVideoSource[];
  audio_sources: EffectiveAudioSource[];
}

export interface NormalizeResponse {
  session_id: string;
  warnings: string[];
  config: EffectiveConfig;
}

export interface PlannedOutput {
  kind: string;
  source_id?: string;
  name: string;
  path: string;
}

export interface VideoJob {
  source: EffectiveVideoSource;
  output_path: string;
}

export interface AudioMixJob {
  name: string;
  sources: EffectiveAudioSource[];
  output: AudioOutputSettings;
  output_path: string;
}

export interface CompileResponse {
  session_id: string;
  warnings: string[];
  outputs: PlannedOutput[];
  video_jobs: VideoJob[];
  audio_jobs: AudioMixJob[];
}

export interface RecordingStartRequest {
  dsl: string;
  max_duration_seconds?: number;
  grace_period_seconds?: number;
}

export interface ProcessLog {
  timestamp: string;
  process_label: string;
  stream: string;
  message: string;
}

export interface RecordingSession {
  active: boolean;
  session_id?: string;
  state?: string;
  reason?: string;
  started_at?: string;
  finished_at?: string;
  warnings: string[];
  outputs: PlannedOutput[];
  logs: ProcessLog[];
}

export interface SessionEnvelope {
  session: RecordingSession;
}

export interface PreviewDescriptor {
  id: string;
  source_id: string;
  name: string;
  source_type: string;
  state: string;
  reason?: string;
  leases: number;
  has_frame: boolean;
  last_frame_at?: string;
}

export interface PreviewEnsureRequest {
  dsl: string;
  source_id: string;
}

export interface PreviewReleaseRequest {
  preview_id: string;
}

export interface PreviewEnvelope {
  preview: PreviewDescriptor;
}

export interface PreviewListResponse {
  previews: PreviewDescriptor[];
}

export interface ServerEventBase {
  type: string;
  timestamp: string;
}

export interface SessionStateEvent extends ServerEventBase {
  type: 'session.state';
  payload: RecordingSession;
}

export interface SessionLogEvent extends ServerEventBase {
  type: 'session.log';
  payload: ProcessLog;
}

export interface PreviewListEvent extends ServerEventBase {
  type: 'preview.list';
  payload: PreviewListResponse;
}

export interface PreviewStateEvent extends ServerEventBase {
  type: 'preview.state';
  payload: PreviewDescriptor;
}

export type WsEvent =
  | SessionStateEvent
  | SessionLogEvent
  | PreviewListEvent
  | PreviewStateEvent;
