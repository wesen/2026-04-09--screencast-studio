import type { Meta, StoryObj } from '@storybook/react';
import { MicMeter } from '../components/MicMeter';

const meta = {
  title: 'Studio/MicMeter',
  component: MicMeter,
  tags: ['autodocs'],
  argTypes: {
    level: { control: { type: 'range', min: 0, max: 1, step: 0.05 } },
  },
} satisfies Meta<typeof MicMeter>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Silence: Story = {
  args: {
    level: 0,
  },
};

export const Low: Story = {
  args: {
    level: 0.25,
  },
};

export const Medium: Story = {
  args: {
    level: 0.5,
  },
};

export const High: Story = {
  args: {
    level: 0.75,
  },
};

export const Clipping: Story = {
  args: {
    level: 0.95,
  },
};
