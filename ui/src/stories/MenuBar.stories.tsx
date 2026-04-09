import type { Meta, StoryObj } from '@storybook/react';
import { MenuBar } from '../components/studio/MenuBar';

const meta = {
  title: 'Studio/MenuBar',
  component: MenuBar,
  tags: ['autodocs'],
} satisfies Meta<typeof MenuBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    armedCount: 1,
    isRecording: false,
    isPaused: false,
  },
};

export const MultipleArmed: Story = {
  args: {
    armedCount: 3,
    isRecording: false,
    isPaused: false,
  },
};

export const NoArmed: Story = {
  args: {
    armedCount: 0,
    isRecording: false,
    isPaused: false,
  },
};

export const Recording: Story = {
  args: {
    armedCount: 2,
    isRecording: true,
    isPaused: false,
  },
};

export const Paused: Story = {
  args: {
    armedCount: 1,
    isRecording: true,
    isPaused: true,
  },
};
