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
    format: 'MOV',
    fps: '24 fps',
    quality: 75,
    audio: '48 kHz, 16-bit',
    multiTrack: true,
    isRecording: false,
    isPaused: false,
    elapsed: 0,
    armedCount: 1,
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
    format: 'MOV',
    fps: '30 fps',
    quality: 85,
    audio: '48 kHz, 16-bit',
    multiTrack: true,
    isRecording: true,
    isPaused: false,
    elapsed: 125,
    armedCount: 2,
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
    format: 'MP4',
    fps: '24 fps',
    quality: 75,
    audio: '44 kHz, 16-bit',
    multiTrack: false,
    isRecording: true,
    isPaused: true,
    elapsed: 3600,
    armedCount: 1,
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
    format: 'AVI',
    fps: '15 fps',
    quality: 30,
    audio: '22 kHz, 8-bit',
    multiTrack: false,
    isRecording: false,
    isPaused: false,
    elapsed: 0,
    armedCount: 1,
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
    format: 'MOV',
    fps: '24 fps',
    quality: 75,
    audio: '48 kHz, 16-bit',
    multiTrack: true,
    isRecording: true,
    isPaused: false,
    elapsed: 3665, // 1 hour, 1 minute, 5 seconds
    armedCount: 3,
    onFormatChange: () => {},
    onFpsChange: () => {},
    onQualityChange: () => {},
    onAudioChange: () => {},
    onMultiTrackChange: () => {},
    onToggleRecording: () => {},
    onTogglePause: () => {},
  },
};
