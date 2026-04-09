import type { Meta, StoryObj } from '@storybook/react';
import { Waveform } from '../components/Waveform';

const meta = {
  title: 'Studio/Waveform',
  component: Waveform,
  tags: ['autodocs'],
  argTypes: {
    active: { control: 'boolean' },
  },
} satisfies Meta<typeof Waveform>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Inactive: Story = {
  args: {
    active: false,
  },
};

export const Active: Story = {
  args: {
    active: true,
  },
};
