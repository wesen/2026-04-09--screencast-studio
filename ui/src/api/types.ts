export interface ApiError {
  code: string;
  message: string;
}

export interface ApiErrorResponse {
  error: ApiError;
}

export type {
  AudioInputDescriptor,
  AudioMixJob,
  AudioOutputSettings,
  AudioSourceSettings,
  CameraDescriptor,
  CompileResponse,
  DiscoveryResponse,
  DisplayDescriptor,
  DslRequest,
  EffectiveAudioSource,
  EffectiveConfig,
  EffectiveVideoSource,
  HealthResponse,
  NormalizeResponse,
  PlannedOutput,
  PreviewDescriptor,
  PreviewEnsureRequest,
  PreviewEnvelope,
  PreviewListResponse,
  PreviewReleaseRequest,
  ProcessLog,
  RecordingSession,
  RecordingStartRequest,
  ServerEvent,
  SessionEnvelope,
  VideoCaptureSettings,
  VideoJob,
  VideoOutputSettings,
  VideoTarget,
  WindowDescriptor,
} from '@/gen/proto/screencast/studio/v1/web_pb';
