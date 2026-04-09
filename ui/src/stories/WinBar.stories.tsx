import type { Meta, StoryObj } from '@storybook/react';
import { WinBar } from '../components/primitives/WinBar';

const meta = {
  title: 'Studio/WinBar',
  component: WinBar,
  tags: ['autodocs'],
  argTypes: {
    children: { control: 'text' },
  },
} satisfies Meta<typeof WinBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: 'Window Title',
  },
};

export const WithClose: Story = {
  args: {
    children: 'Closable Window',
    onClose: () => console.log('close clicked'),
  },
};

export const Sources: Story = {
  args: {
    children: 'Sources (3)',
  },
};

export const Output: Story = {
  args: {
    children: 'Output Parameters',
  },
};

export const Microphone: Story = {
  args: {
    children: 'Microphone',
  },
};

export const Status: Story = {
  args: {
    children: 'Status',
  },
};
