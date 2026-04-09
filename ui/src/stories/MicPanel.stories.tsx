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
    micLevel: 0.12,
    micInput: 'Built-in Mic',
    gain: 55,
    isRecording: false,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const Recording: Story = {
  args: {
    micLevel: 0.45,
    micInput: 'Built-in Mic',
    gain: 55,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const RecordingHighLevel: Story = {
  args: {
    micLevel: 0.78,
    micInput: 'Built-in Mic',
    gain: 75,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const ExternalMic: Story = {
  args: {
    micLevel: 0.35,
    micInput: 'External',
    gain: 60,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const LineIn: Story = {
  args: {
    micLevel: 0.5,
    micInput: 'Line In',
    gain: 50,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const LowGain: Story = {
  args: {
    micLevel: 0.15,
    micInput: 'Built-in Mic',
    gain: 20,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};

export const HighGain: Story = {
  args: {
    micLevel: 0.92,
    micInput: 'Built-in Mic',
    gain: 95,
    isRecording: true,
    onMicInputChange: () => {},
    onGainChange: () => {},
  },
};
