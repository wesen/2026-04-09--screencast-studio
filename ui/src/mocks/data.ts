import { create } from '@bufbuild/protobuf';
import type {
  CompileResponse,
  DiscoveryResponse,
  PreviewDescriptor,
  RecordingSession,
} from '@/api/types';
import {
  CompileResponseSchema,
  DiscoveryResponseSchema,
  RecordingSessionSchema,
} from '@/gen/proto/screencast/studio/v1/web_pb';

export const mockDiscoveryData: DiscoveryResponse = create(DiscoveryResponseSchema, {
  displays: [
    {
      id: 'display-1',
      name: 'Display 1',
      primary: true,
      x: 0,
      y: 0,
      width: 1920,
      height: 1080,
      connector: 'HDMI-A-1',
    },
    {
      id: 'display-2',
      name: 'Display 2',
      primary: false,
      x: 1920,
      y: 0,
      width: 2560,
      height: 1440,
      connector: 'DP-1',
    },
  ],
  windows: [
    {
      id: '0x3a00007',
      title: 'Browser Window',
      x: 10,
      y: 20,
      width: 1200,
      height: 800,
    },
    {
      id: '0x3a00008',
      title: 'Terminal',
      x: 100,
      y: 100,
      width: 800,
      height: 600,
    },
  ],
  cameras: [
    {
      id: '/dev/video0',
      label: 'Built-in Camera',
      device: '/dev/video0',
      cardName: 'FaceTime HD',
    },
  ],
  audio: [
    {
      id: 'alsa_input.pci-0000_00_1f.3.analog-stereo',
      name: 'Built-in Mic',
      driver: 'ALSA',
      sampleSpec: 'S16LE 48000 Hz 2 channels',
      state: 'RUNNING',
    },
  ],
});

export const mockRecordingSession: RecordingSession = create(RecordingSessionSchema, {
  active: false,
  outputs: [],
  warnings: [],
  logs: [],
});

export const mockCompileResponse: CompileResponse = create(CompileResponseSchema, {
  sessionId: 'demo-session',
  warnings: [],
  outputs: [
    {
      kind: 'video',
      sourceId: 'desktop-1',
      name: 'Full Desktop',
      path: 'recordings/demo-session/Full Desktop.mov',
    },
  ],
  videoJobs: [],
  audioJobs: [],
});

export const mockPreviews: PreviewDescriptor[] = [];
