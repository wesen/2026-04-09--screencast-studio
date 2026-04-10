import type { Meta, StoryObj } from '@storybook/react';
import { StatusPanel } from '../components/studio/StatusPanel';
import type { StudioSource } from '../components/source-card';

const meta = {
  title: 'Studio/StatusPanel',
  component: StatusPanel,
  tags: ['autodocs'],
} satisfies Meta<typeof StatusPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const createSource = (id: string, label: string, armed: boolean): StudioSource => ({
  id,
  sourceId: id,
  kind: 'Display',
  scene: 'Desktop 1',
  armed,
  label,
});

export const Ready: Story = {
  args: {
    diskPercent: 8,
    isRecording: false,
    isPaused: false,
    armedSources: [createSource('display-1', 'Display 1', true)],
  },
};

export const ReadyLowDisk: Story = {
  args: {
    diskPercent: 85,
    isRecording: false,
    isPaused: false,
    armedSources: [createSource('display-1', 'Display 1', true)],
  },
};

export const Recording: Story = {
  args: {
    diskPercent: 15,
    isRecording: true,
    isPaused: false,
    armedSources: [createSource('display-1', 'Display 1', true)],
  },
};

export const RecordingMultipleSources: Story = {
  args: {
    diskPercent: 25,
    isRecording: true,
    isPaused: false,
    armedSources: [
      createSource('display-1', 'Display 1', true),
      createSource('camera-1', 'Camera 1', true),
      createSource('display-2', 'Display 2', true),
    ],
  },
};

export const Paused: Story = {
  args: {
    diskPercent: 40,
    isRecording: true,
    isPaused: true,
    armedSources: [createSource('display-1', 'Display 1', true)],
  },
};

export const NoArmedSources: Story = {
  args: {
    diskPercent: 50,
    isRecording: false,
    isPaused: false,
    armedSources: [],
  },
};

export const DiskAlmostFull: Story = {
  args: {
    diskPercent: 92,
    isRecording: false,
    isPaused: false,
    armedSources: [createSource('display-1', 'Display 1', true)],
  },
};
