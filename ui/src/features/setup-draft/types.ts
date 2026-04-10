export type SetupDraftVideoSourceKind = 'display' | 'window' | 'region' | 'camera';

export interface SetupDraftRect {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface SetupDraftVideoCapture {
  fps: number;
  cursor: boolean | null;
  followResize: boolean | null;
  mirror: boolean | null;
  size: string;
}

export interface SetupDraftVideoOutput {
  container: string;
  videoCodec: string;
  quality: number;
}

interface SetupDraftVideoSourceBase {
  id: string;
  kind: SetupDraftVideoSourceKind;
  name: string;
  enabled: boolean;
  destinationTemplate: string;
  capture: SetupDraftVideoCapture;
  output: SetupDraftVideoOutput;
}

export interface SetupDraftDisplaySource extends SetupDraftVideoSourceBase {
  kind: 'display';
  target: {
    displayId: string;
  };
}

export interface SetupDraftWindowSource extends SetupDraftVideoSourceBase {
  kind: 'window';
  target: {
    windowId: string;
  };
}

export interface SetupDraftRegionSource extends SetupDraftVideoSourceBase {
  kind: 'region';
  target: {
    displayId: string;
    rect: SetupDraftRect;
  };
}

export interface SetupDraftCameraSource extends SetupDraftVideoSourceBase {
  kind: 'camera';
  target: {
    deviceId: string;
  };
}

export type SetupDraftVideoSource =
  | SetupDraftDisplaySource
  | SetupDraftWindowSource
  | SetupDraftRegionSource
  | SetupDraftCameraSource;

export interface SetupDraftAudioSource {
  id: string;
  name: string;
  deviceId: string;
  enabled: boolean;
  gain: number;
  noiseGate: boolean;
  denoise: boolean;
}

export interface SetupDraftAudioOutput {
  codec: string;
  sampleRateHz: number;
  channels: number;
}

export interface SetupDraftDocument {
  schema: string;
  sessionId: string;
  destinationTemplates: Record<string, string>;
  audioMixTemplate: string;
  audioOutput: SetupDraftAudioOutput;
  videoSources: SetupDraftVideoSource[];
  audioSources: SetupDraftAudioSource[];
}
