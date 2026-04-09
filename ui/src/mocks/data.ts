import type {
  DiscoveryResponse,
  RecordingState,
  CompileResponse,
  PreviewDescriptor,
} from '@/api/types';

export const mockDiscoveryData: DiscoveryResponse = {
  generated_at: new Date().toISOString(),
  items: [
    {
      kind: 'display',
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
      kind: 'display',
      id: 'display-2',
      name: 'Display 2',
      primary: false,
      x: 1920,
      y: 0,
      width: 2560,
      height: 1440,
      connector: 'DP-1',
    },
    {
      kind: 'window',
      id: '0x3a00007',
      title: 'Browser Window',
      x: 10,
      y: 20,
      width: 1200,
      height: 800,
    },
    {
      kind: 'window',
      id: '0x3a00008',
      title: 'Terminal',
      x: 100,
      y: 100,
      width: 800,
      height: 600,
    },
    {
      kind: 'camera',
      id: '/dev/video0',
      label: 'Built-in Camera',
      device: '/dev/video0',
      card_name: 'FaceTime HD',
    },
    {
      kind: 'audio',
      id: 'alsa_input.pci-0000_00_1f.3.analog-stereo',
      name: 'Built-in Mic',
      driver: 'ALSA',
      sample_spec: 'S16LE 48000 Hz 2 channels',
      state: 'RUNNING',
    },
  ],
};

export const mockRecordingState: RecordingState = {
  active: false,
  session_id: '',
  state: 'idle',
  reason: '',
  outputs: [],
  warnings: [],
};

export const mockCompileResponse: CompileResponse = {
  session_id: 'demo-session',
  warnings: [],
  outputs: [
    {
      kind: 'video',
      source_id: 'desktop-1',
      name: 'Full Desktop',
      path: 'recordings/demo-session/Full Desktop.mov',
    },
  ],
  video_jobs: 1,
  audio_jobs: 1,
};

export const mockPreviews: PreviewDescriptor[] = [];
