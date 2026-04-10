import type { Meta, StoryObj } from '@storybook/react';
import { MicPanel } from '../components/studio/MicPanel';

const meta = {
  title: 'Studio/MicPanel',
  component: MicPanel,
  tags: ['autodocs'],
} satisfies Meta<typeof MicPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    leftLevel: 0.12,
    rightLevel: 0.09,
    micInput: 'default',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 55,
    isRecording: false,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const Recording: Story = {
  args: {
    leftLevel: 0.45,
    rightLevel: 0.41,
    micInput: 'default',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 55,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const RecordingHighLevel: Story = {
  args: {
    leftLevel: 0.78,
    rightLevel: 0.74,
    micInput: 'default',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 75,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const ExternalMic: Story = {
  args: {
    leftLevel: 0.35,
    rightLevel: 0.31,
    micInput: 'usb',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 60,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const LineIn: Story = {
  args: {
    leftLevel: 0.5,
    rightLevel: 0.48,
    micInput: 'line-in',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'line-in', label: 'Line In' },
    ],
    gain: 50,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const LowGain: Story = {
  args: {
    leftLevel: 0.15,
    rightLevel: 0.12,
    micInput: 'default',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 20,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const HighGain: Story = {
  args: {
    leftLevel: 0.92,
    rightLevel: 0.88,
    micInput: 'default',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 95,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const Unavailable: Story = {
  args: {
    micInput: 'default',
    micOptions: [
      { value: 'default', label: 'Built-in Mic' },
      { value: 'usb', label: 'USB Interface' },
    ],
    gain: 55,
    isRecording: false,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};
