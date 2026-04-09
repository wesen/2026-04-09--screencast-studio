import type { Meta, StoryObj } from '@storybook/react';
import { StatusPanel } from '../components/studio/StatusPanel';
import type { Source } from '../features/studio-draft/studioDraftSlice';

const meta = {
  title: 'Studio/StatusPanel',
  component: StatusPanel,
  tags: ['autodocs'],
} satisfies Meta<typeof StatusPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const createSource = (id: number, label: string, armed: boolean): Source => ({
  id,
  kind: 'Display',
  scene: 'Desktop 1',
  armed,
  solo: false,
  label,
});

export const Ready: Story = {
  args: {
    diskPercent: 8,
    isRecording: false,
    isPaused: false,
    armedSources: [createSource(1, 'Display 1', true)],
  },
};

export const ReadyLowDisk: Story = {
  args: {
    diskPercent: 85,
    isRecording: false,
    isPaused: false,
    armedSources: [createSource(1, 'Display 1', true)],
  },
};

export const Recording: Story = {
  args: {
    diskPercent: 15,
    isRecording: true,
    isPaused: false,
    armedSources: [createSource(1, 'Display 1', true)],
  },
};

export const RecordingMultipleSources: Story = {
  args: {
    diskPercent: 25,
    isRecording: true,
    isPaused: false,
    armedSources: [
      createSource(1, 'Display 1', true),
      createSource(2, 'Camera 1', true),
      createSource(3, 'Display 2', true),
    ],
  },
};

export const Paused: Story = {
  args: {
    diskPercent: 40,
    isRecording: true,
    isPaused: true,
    armedSources: [createSource(1, 'Display 1', true)],
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
    armedSources: [createSource(1, 'Display 1', true)],
  },
};
