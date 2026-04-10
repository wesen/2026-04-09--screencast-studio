import type { EffectiveAudioSource, EffectiveConfig, EffectiveVideoSource } from '@/api/types';
import type {
  SetupDraftAudioSource,
  SetupDraftDocument,
  SetupDraftRect,
  SetupDraftVideoCapture,
  SetupDraftVideoOutput,
  SetupDraftVideoSource,
  SetupDraftVideoSourceKind,
} from './types';

const normalizeVideoSourceKind = (value: string): SetupDraftVideoSourceKind => {
  const normalized = value.trim().toLowerCase();
  if (
    normalized === 'display' ||
    normalized === 'window' ||
    normalized === 'region' ||
    normalized === 'camera'
  ) {
    return normalized;
  }
  return 'display';
};

const toDraftRect = (
  rect?: { x?: number; y?: number; w?: number; h?: number } | undefined
): SetupDraftRect => ({
  x: rect?.x ?? 0,
  y: rect?.y ?? 0,
  w: rect?.w ?? 0,
  h: rect?.h ?? 0,
});

const toDraftCapture = (
  capture?: {
    fps?: number;
    cursor?: boolean;
    followResize?: boolean;
    mirror?: boolean;
    size?: string;
  } | undefined
): SetupDraftVideoCapture => ({
  fps: capture?.fps ?? 0,
  cursor: capture?.cursor ?? null,
  followResize: capture?.followResize ?? null,
  mirror: capture?.mirror ?? null,
  size: capture?.size ?? '',
});

const toDraftOutput = (
  output?: {
    container?: string;
    videoCodec?: string;
    quality?: number;
  } | undefined
): SetupDraftVideoOutput => ({
  container: output?.container ?? '',
  videoCodec: output?.videoCodec ?? '',
  quality: output?.quality ?? 0,
});

export const effectiveVideoSourceToDraft = (
  source: EffectiveVideoSource
): SetupDraftVideoSource => {
  const base = {
    id: source.id,
    name: source.name,
    enabled: source.enabled,
    destinationTemplate: source.destinationTemplate,
    capture: toDraftCapture(source.capture),
    output: toDraftOutput(source.output),
  };

  switch (normalizeVideoSourceKind(source.type)) {
    case 'window':
      return {
        ...base,
        kind: 'window',
        target: {
          windowId: source.target?.windowId ?? '',
        },
      };
    case 'region':
      return {
        ...base,
        kind: 'region',
        target: {
          displayId: source.target?.display ?? '',
          rect: toDraftRect(source.target?.rect),
        },
      };
    case 'camera':
      return {
        ...base,
        kind: 'camera',
        target: {
          deviceId: source.target?.device ?? '',
        },
      };
    case 'display':
    default:
      return {
        ...base,
        kind: 'display',
        target: {
          displayId: source.target?.display ?? '',
        },
      };
  }
};

export const effectiveAudioSourceToDraft = (
  source: EffectiveAudioSource
): SetupDraftAudioSource => ({
  id: source.id,
  name: source.name,
  deviceId: source.device,
  enabled: source.enabled,
  gain: source.settings?.gain ?? 0,
  noiseGate: source.settings?.noiseGate ?? false,
  denoise: source.settings?.denoise ?? false,
});

export const effectiveConfigToSetupDraft = (
  config: EffectiveConfig
): SetupDraftDocument => ({
  sessionId: config.sessionId,
  videoSources: config.videoSources.map(effectiveVideoSourceToDraft),
  audioSources: config.audioSources.map(effectiveAudioSourceToDraft),
});
