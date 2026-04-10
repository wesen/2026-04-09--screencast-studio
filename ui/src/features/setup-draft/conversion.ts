import type {
  CameraDescriptor,
  DisplayDescriptor,
  EffectiveAudioSource,
  EffectiveConfig,
  EffectiveVideoSource,
  WindowDescriptor,
} from '@/api/types';
import type {
  SetupDraftAudioSource,
  SetupDraftAudioOutput,
  SetupDraftDocument,
  SetupDraftRect,
  SetupDraftVideoCapture,
  SetupDraftVideoOutput,
  SetupDraftVideoSource,
  SetupDraftVideoSourceKind,
} from './types';

const DEFAULT_SCHEMA = 'recorder.config/v1';
const DEFAULT_DISPLAY_TARGET = ':0.0';
const DEFAULT_DESTINATION_ROOT = 'recordings';
const PER_SOURCE_SUFFIX = '/{session_id}/{source_name}.{ext}';
const AUDIO_MIX_SUFFIX = '/{session_id}/audio-mix.{ext}';

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

const toDraftAudioOutput = (
  output?: {
    codec?: string;
    sampleRateHz?: number;
    channels?: number;
  } | undefined
): SetupDraftAudioOutput => ({
  codec: output?.codec ?? 'pcm_s16le',
  sampleRateHz: output?.sampleRateHz ?? 48000,
  channels: output?.channels ?? 2,
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
  schema: config.schema,
  sessionId: config.sessionId,
  destinationTemplates: { ...config.destinationTemplates },
  audioMixTemplate: config.audioMixTemplate,
  audioOutput: toDraftAudioOutput(config.audioOutput),
  videoSources: config.videoSources.map(effectiveVideoSourceToDraft),
  audioSources: config.audioSources.map(effectiveAudioSourceToDraft),
});

const slugify = (value: string): string => value
  .trim()
  .toLowerCase()
  .replace(/[^a-z0-9]+/g, '-')
  .replace(/^-+|-+$/g, '') || 'source';

const nextUniqueSourceId = (
  base: string,
  sources: SetupDraftVideoSource[]
): string => {
  const existing = new Set(sources.map((source) => source.id));
  if (!existing.has(base)) {
    return base;
  }
  let counter = 2;
  while (existing.has(`${base}-${counter}`)) {
    counter += 1;
  }
  return `${base}-${counter}`;
};

const copyCapture = (
  capture?: Partial<SetupDraftVideoCapture>
): SetupDraftVideoCapture => ({
  fps: capture?.fps ?? 24,
  cursor: capture?.cursor ?? true,
  followResize: capture?.followResize ?? false,
  mirror: capture?.mirror ?? false,
  size: capture?.size ?? '',
});

const copyOutput = (
  output?: Partial<SetupDraftVideoOutput>
): SetupDraftVideoOutput => ({
  container: output?.container ?? 'mov',
  videoCodec: output?.videoCodec ?? 'h264',
  quality: output?.quality ?? 75,
});

const preferredTemplateName = (draft: SetupDraftDocument): string => {
  if (draft.destinationTemplates.per_source) {
    return 'per_source';
  }
  const [firstTemplate] = Object.keys(draft.destinationTemplates);
  return firstTemplate ?? 'per_source';
};

const seedVideoSettings = (
  draft: SetupDraftDocument,
  kind: SetupDraftVideoSourceKind
): { capture: SetupDraftVideoCapture; output: SetupDraftVideoOutput } => {
  const existing = draft.videoSources.find((source) => source.kind === kind);
  if (existing) {
    return {
      capture: copyCapture(existing.capture),
      output: copyOutput(existing.output),
    };
  }

  if (kind === 'camera') {
    return {
      capture: copyCapture({ fps: 30, mirror: false, size: '1280x720' }),
      output: copyOutput({ container: 'mov', videoCodec: 'h264', quality: 80 }),
    };
  }

  return {
    capture: copyCapture({ fps: 24, cursor: true, followResize: false }),
    output: copyOutput({ container: 'mov', videoCodec: 'h264', quality: 75 }),
  };
};

export const createDisplaySourceDraft = (
  display: DisplayDescriptor,
  draft: SetupDraftDocument
): SetupDraftVideoSource => {
  const settings = seedVideoSettings(draft, 'display');
  const name = display.name || display.connector || 'Display';
  return {
    id: nextUniqueSourceId(slugify(name), draft.videoSources),
    kind: 'display',
    name,
    enabled: true,
    destinationTemplate: preferredTemplateName(draft),
    capture: settings.capture,
    output: settings.output,
    target: {
      displayId: DEFAULT_DISPLAY_TARGET,
    },
  };
};

export const createWindowSourceDraft = (
  window: WindowDescriptor,
  draft: SetupDraftDocument
): SetupDraftVideoSource => {
  const settings = seedVideoSettings(draft, 'window');
  const name = window.title || 'Window';
  return {
    id: nextUniqueSourceId(slugify(name), draft.videoSources),
    kind: 'window',
    name,
    enabled: true,
    destinationTemplate: preferredTemplateName(draft),
    capture: settings.capture,
    output: settings.output,
    target: {
      windowId: window.id,
    },
  };
};

export const createCameraSourceDraft = (
  camera: CameraDescriptor,
  draft: SetupDraftDocument
): SetupDraftVideoSource => {
  const settings = seedVideoSettings(draft, 'camera');
  const name = camera.label || camera.device || 'Camera';
  return {
    id: nextUniqueSourceId(slugify(name), draft.videoSources),
    kind: 'camera',
    name,
    enabled: true,
    destinationTemplate: preferredTemplateName(draft),
    capture: settings.capture,
    output: settings.output,
    target: {
      deviceId: camera.device,
    },
  };
};

export type RegionPreset = 'full' | 'top-half' | 'bottom-half' | 'left-half' | 'right-half';

export const presetRectForDisplay = (
  display: DisplayDescriptor,
  preset: RegionPreset
): SetupDraftRect => {
  switch (preset) {
    case 'top-half':
      return { x: display.x, y: display.y, w: display.width, h: Math.floor(display.height / 2) };
    case 'bottom-half':
      return {
        x: display.x,
        y: display.y + Math.floor(display.height / 2),
        w: display.width,
        h: display.height - Math.floor(display.height / 2),
      };
    case 'left-half':
      return { x: display.x, y: display.y, w: Math.floor(display.width / 2), h: display.height };
    case 'right-half':
      return {
        x: display.x + Math.floor(display.width / 2),
        y: display.y,
        w: display.width - Math.floor(display.width / 2),
        h: display.height,
      };
    case 'full':
    default:
      return { x: display.x, y: display.y, w: display.width, h: display.height };
  }
};

export const createRegionSourceDraft = (
  display: DisplayDescriptor,
  preset: RegionPreset,
  draft: SetupDraftDocument
): SetupDraftVideoSource => {
  const settings = seedVideoSettings(draft, 'region');
  const suffix = preset === 'full' ? 'region' : preset.replace(/-/g, ' ');
  const name = `${display.name || display.connector || 'Display'} ${suffix}`;
  return {
    id: nextUniqueSourceId(slugify(name), draft.videoSources),
    kind: 'region',
    name,
    enabled: true,
    destinationTemplate: preferredTemplateName(draft),
    capture: settings.capture,
    output: settings.output,
    target: {
      displayId: DEFAULT_DISPLAY_TARGET,
      rect: presetRectForDisplay(display, preset),
    },
  };
};

const yamlQuote = (value: string): string => JSON.stringify(value);

const renderYamlLine = (
  indent: number,
  key: string,
  value: string | number | boolean
): string => `${' '.repeat(indent)}${key}: ${typeof value === 'string' ? yamlQuote(value) : String(value)}`;

const renderCaptureBlock = (
  indent: number,
  capture: SetupDraftVideoCapture
): string[] => {
  const lines = [renderYamlLine(indent, 'fps', capture.fps)];
  if (capture.cursor !== null) {
    lines.push(renderYamlLine(indent, 'cursor', capture.cursor));
  }
  if (capture.followResize !== null) {
    lines.push(renderYamlLine(indent, 'follow_resize', capture.followResize));
  }
  if (capture.mirror !== null) {
    lines.push(renderYamlLine(indent, 'mirror', capture.mirror));
  }
  if (capture.size) {
    lines.push(renderYamlLine(indent, 'size', capture.size));
  }
  return lines;
};

const renderOutputBlock = (
  indent: number,
  output: SetupDraftVideoOutput
): string[] => [
  renderYamlLine(indent, 'container', output.container),
  renderYamlLine(indent, 'video_codec', output.videoCodec),
  renderYamlLine(indent, 'quality', output.quality),
];

export const renderSetupDraftAsDsl = (draft: SetupDraftDocument): string => {
  const lines: string[] = [
    renderYamlLine(0, 'schema', draft.schema || DEFAULT_SCHEMA),
  ];

  if (draft.sessionId) {
    lines.push(renderYamlLine(0, 'session_id', draft.sessionId));
  }

  lines.push('', 'destination_templates:');
  for (const [name, template] of Object.entries(draft.destinationTemplates)) {
    lines.push(renderYamlLine(2, name, template));
  }

  lines.push(
    '',
    'audio_defaults:',
    '  output:',
    renderYamlLine(4, 'codec', draft.audioOutput.codec),
    renderYamlLine(4, 'sample_rate_hz', draft.audioOutput.sampleRateHz),
    renderYamlLine(4, 'channels', draft.audioOutput.channels)
  );

  if (draft.audioMixTemplate) {
    lines.push('', 'audio_mix:', renderYamlLine(2, 'destination_template', draft.audioMixTemplate));
  }

  lines.push('', 'video_sources:');
  if (draft.videoSources.length === 0) {
    lines.push('  []');
  } else {
    for (const source of draft.videoSources) {
      lines.push(
        `  - id: ${yamlQuote(source.id)}`,
        renderYamlLine(4, 'name', source.name),
        renderYamlLine(4, 'type', source.kind),
        renderYamlLine(4, 'enabled', source.enabled),
        '    target:'
      );
      switch (source.kind) {
        case 'display':
          lines.push(renderYamlLine(6, 'display', source.target.displayId));
          break;
        case 'window':
          lines.push(
            renderYamlLine(6, 'display', DEFAULT_DISPLAY_TARGET),
            renderYamlLine(6, 'window_id', source.target.windowId)
          );
          break;
        case 'region':
          lines.push(
            renderYamlLine(6, 'display', source.target.displayId),
            '      rect:',
            renderYamlLine(8, 'x', source.target.rect.x),
            renderYamlLine(8, 'y', source.target.rect.y),
            renderYamlLine(8, 'w', source.target.rect.w),
            renderYamlLine(8, 'h', source.target.rect.h)
          );
          break;
        case 'camera':
          lines.push(renderYamlLine(6, 'device', source.target.deviceId));
          break;
      }
      lines.push(
        '    settings:',
        '      capture:',
        ...renderCaptureBlock(8, source.capture),
        '      output:',
        ...renderOutputBlock(8, source.output),
        renderYamlLine(4, 'destination_template', source.destinationTemplate)
      );
    }
  }

  lines.push('', 'audio_sources:');
  if (draft.audioSources.length === 0) {
    lines.push('  []');
  } else {
    for (const source of draft.audioSources) {
      lines.push(
        `  - id: ${yamlQuote(source.id)}`,
        renderYamlLine(4, 'name', source.name),
        renderYamlLine(4, 'device', source.deviceId),
        renderYamlLine(4, 'enabled', source.enabled),
        '    settings:',
        renderYamlLine(6, 'gain', source.gain),
        renderYamlLine(6, 'noise_gate', source.noiseGate),
        renderYamlLine(6, 'denoise', source.denoise)
      );
    }
  }

  return `${lines.join('\n')}\n`;
};

const trimTrailingSlashes = (value: string): string => value.replace(/\/+$/g, '');

export const destinationRootFromTemplates = (
  templates: Record<string, string>
): string | null => {
  const perSource = templates.per_source;
  const audioMix = templates.audio_mix;
  if (!perSource || !audioMix) {
    return null;
  }
  if (!perSource.endsWith(PER_SOURCE_SUFFIX) || !audioMix.endsWith(AUDIO_MIX_SUFFIX)) {
    return null;
  }
  const perSourceRoot = trimTrailingSlashes(perSource.slice(0, -PER_SOURCE_SUFFIX.length));
  const audioMixRoot = trimTrailingSlashes(audioMix.slice(0, -AUDIO_MIX_SUFFIX.length));
  if (perSourceRoot !== audioMixRoot) {
    return null;
  }
  return perSourceRoot || DEFAULT_DESTINATION_ROOT;
};

export const applyDestinationRootToTemplates = (
  destinationRoot: string,
  templates: Record<string, string>
): Record<string, string> => {
  const root = trimTrailingSlashes(destinationRoot.trim()) || DEFAULT_DESTINATION_ROOT;
  return {
    ...templates,
    per_source: `${root}${PER_SOURCE_SUFFIX}`,
    audio_mix: `${root}${AUDIO_MIX_SUFFIX}`,
  };
};
