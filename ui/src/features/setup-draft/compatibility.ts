import type { EffectiveConfig } from '@/api/types';

const normalizeConfigSignature = (config: EffectiveConfig) => JSON.stringify({
  schema: config.schema,
  sessionId: config.sessionId,
  destinationTemplates: Object.fromEntries(
    Object.entries(config.destinationTemplates ?? {}).sort(([a], [b]) => a.localeCompare(b))
  ),
  audioMixTemplate: config.audioMixTemplate,
  audioOutput: config.audioOutput
    ? {
        codec: config.audioOutput.codec,
        sampleRateHz: config.audioOutput.sampleRateHz,
        channels: config.audioOutput.channels,
      }
    : null,
  videoSources: (config.videoSources ?? []).map((source) => ({
    id: source.id,
    name: source.name,
    type: source.type,
    enabled: source.enabled,
    destinationTemplate: source.destinationTemplate,
    target: source.target
      ? {
          display: source.target.display,
          windowId: source.target.windowId,
          device: source.target.device,
          rect: source.target.rect
            ? {
                x: source.target.rect.x,
                y: source.target.rect.y,
                w: source.target.rect.w,
                h: source.target.rect.h,
              }
            : null,
        }
      : null,
    capture: source.capture
      ? {
          fps: source.capture.fps,
          cursor: source.capture.cursor ?? null,
          followResize: source.capture.followResize ?? null,
          mirror: source.capture.mirror ?? null,
          size: source.capture.size ?? '',
        }
      : null,
    output: source.output
      ? {
          container: source.output.container,
          videoCodec: source.output.videoCodec,
          quality: source.output.quality,
        }
      : null,
  })),
  audioSources: (config.audioSources ?? []).map((source) => ({
    id: source.id,
    name: source.name,
    enabled: source.enabled,
    device: source.device,
    settings: source.settings
      ? {
          gain: source.settings.gain,
          noiseGate: source.settings.noiseGate,
          denoise: source.settings.denoise,
        }
      : null,
  })),
});

export const isBuilderCompatibleEffectiveConfig = (
  original: EffectiveConfig,
  roundTripped: EffectiveConfig
): boolean => normalizeConfigSignature(original) === normalizeConfigSignature(roundTripped);
