import type { Meta, StoryObj } from '@storybook/react';
import { OutputPanel } from '../components/studio/OutputPanel';

const meta = {
  title: 'Studio/OutputPanel',
  component: OutputPanel,
  tags: ['autodocs'],
} satisfies Meta<typeof OutputPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    recordingName: 'demo',
    destinationRoot: 'recordings',
    destinationRootEditable: true,
    filenameSuffix: '',
    filenameSuffixEditable: true,
    outputs: [
      { kind: 'video', sourceId: 'desktop-1', name: 'Full Desktop', path: 'recordings/demo/Full-Desktop.mov' },
      { kind: 'audio', sourceId: '', name: 'audio-mix', path: 'recordings/demo/audio-mix.wav' },
    ],
    format: 'MOV',
    fps: '24 fps',
    quality: 75,
    audio: '48 kHz, 16-bit',
    multiTrack: true,
    isRecording: false,
    isPaused: false,
    elapsed: 0,
    armedCount: 1,
    onRecordingNameChange: () => {},
    onDestinationRootChange: () => {},
    onFilenameSuffixChange: () => {},
    onFormatChange: () => {},
    onFpsChange: () => {},
    onQualityChange: () => {},
    onAudioChange: () => {},
    onMultiTrackChange: () => {},
    onToggleRecording: () => {},
    onTogglePause: () => {},
  },
};

export const Recording: Story = {
  args: {
    recordingName: 'interview-01',
    destinationRoot: '/tmp/captures',
    destinationRootEditable: true,
    filenameSuffix: '-{date}-{index}',
    filenameSuffixEditable: true,
    outputs: [
      { kind: 'video', sourceId: 'desktop-1', name: 'Full Desktop', path: '/tmp/captures/interview-01/Full-Desktop-2026-04-10-1.mov' },
      { kind: 'video', sourceId: 'cam-1', name: 'Camera', path: '/tmp/captures/interview-01/Camera-2026-04-10-1.mov' },
      { kind: 'audio', sourceId: '', name: 'audio-mix', path: '/tmp/captures/interview-01/audio-mix-2026-04-10-1.wav' },
    ],
    format: 'MOV',
    fps: '30 fps',
    quality: 85,
    audio: '48 kHz, 16-bit',
    multiTrack: true,
    isRecording: true,
    isPaused: false,
    elapsed: 125,
    armedCount: 2,
    onRecordingNameChange: () => {},
    onDestinationRootChange: () => {},
    onFilenameSuffixChange: () => {},
    onFormatChange: () => {},
    onFpsChange: () => {},
    onQualityChange: () => {},
    onAudioChange: () => {},
    onMultiTrackChange: () => {},
    onToggleRecording: () => {},
    onTogglePause: () => {},
  },
};

export const Paused: Story = {
  args: {
    recordingName: 'paused-demo',
    destinationRoot: 'recordings',
    destinationRootEditable: true,
    filenameSuffix: '',
    filenameSuffixEditable: true,
    outputs: [
      { kind: 'video', sourceId: 'desktop-1', name: 'Full Desktop', path: 'recordings/paused-demo/Full-Desktop.mp4' },
    ],
    format: 'MP4',
    fps: '24 fps',
    quality: 75,
    audio: '44 kHz, 16-bit',
    multiTrack: false,
    isRecording: true,
    isPaused: true,
    elapsed: 3600,
    armedCount: 1,
    onRecordingNameChange: () => {},
    onDestinationRootChange: () => {},
    onFilenameSuffixChange: () => {},
    onFormatChange: () => {},
    onFpsChange: () => {},
    onQualityChange: () => {},
    onAudioChange: () => {},
    onMultiTrackChange: () => {},
    onToggleRecording: () => {},
    onTogglePause: () => {},
  },
};

export const LowQuality: Story = {
  args: {
    recordingName: 'quick-demo',
    destinationRoot: 'recordings',
    destinationRootEditable: true,
    filenameSuffix: '',
    filenameSuffixEditable: true,
    outputs: [
      { kind: 'video', sourceId: 'desktop-1', name: 'Full Desktop', path: 'recordings/quick-demo/Full-Desktop.avi' },
    ],
    format: 'AVI',
    fps: '15 fps',
    quality: 30,
    audio: '22 kHz, 8-bit',
    multiTrack: false,
    isRecording: false,
    isPaused: false,
    elapsed: 0,
    armedCount: 1,
    onRecordingNameChange: () => {},
    onDestinationRootChange: () => {},
    onFilenameSuffixChange: () => {},
    onFormatChange: () => {},
    onFpsChange: () => {},
    onQualityChange: () => {},
    onAudioChange: () => {},
    onMultiTrackChange: () => {},
    onToggleRecording: () => {},
    onTogglePause: () => {},
  },
};

export const LongRecording: Story = {
  args: {
    recordingName: 'long-session',
    destinationRoot: '/srv/recordings',
    destinationRootEditable: true,
    filenameSuffix: '-{timestamp}',
    filenameSuffixEditable: true,
    outputs: [
      { kind: 'video', sourceId: 'desktop-1', name: 'Display 1', path: '/srv/recordings/long-session/Display-1-20260410-130500.mov' },
      { kind: 'video', sourceId: 'camera-1', name: 'Camera 1', path: '/srv/recordings/long-session/Camera-1-20260410-130500.mov' },
      { kind: 'audio', sourceId: '', name: 'audio-mix', path: '/srv/recordings/long-session/audio-mix-20260410-130500.wav' },
    ],
    format: 'MOV',
    fps: '24 fps',
    quality: 75,
    audio: '48 kHz, 16-bit',
    multiTrack: true,
    isRecording: true,
    isPaused: false,
    elapsed: 3665, // 1 hour, 1 minute, 5 seconds
    armedCount: 3,
    onRecordingNameChange: () => {},
    onDestinationRootChange: () => {},
    onFilenameSuffixChange: () => {},
    onFormatChange: () => {},
    onFpsChange: () => {},
    onQualityChange: () => {},
    onAudioChange: () => {},
    onMultiTrackChange: () => {},
    onToggleRecording: () => {},
    onTogglePause: () => {},
  },
};

export const AdvancedTemplatesLocked: Story = {
  args: {
    ...Default.args,
    destinationRoot: '',
    destinationRootEditable: false,
    filenameSuffix: '',
    filenameSuffixEditable: false,
    destinationRootReason: 'Advanced destination templates are active. Edit Raw DSL to change output paths.',
  },
};

export const InvalidDestination: Story = {
  args: {
    ...Default.args,
    destinationRoot: '/root/forbidden',
    outputPreviewErrors: ['compile failed: destination template rendered to an invalid or unwritable path'],
    outputs: [],
  },
};
