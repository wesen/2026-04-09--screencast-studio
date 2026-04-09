import type { Meta, StoryObj } from '@storybook/react';
import { PreviewStream } from '../components/preview/PreviewStream';

const meta = {
  title: 'Studio/PreviewStream',
  component: PreviewStream,
  tags: ['autodocs'],
} satisfies Meta<typeof PreviewStream>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    sourceId: 'desktop-1',
    streamUrl: undefined,
  },
};

export const Unavailable: Story = {
  args: {
    sourceId: 'desktop-1',
    streamUrl: undefined,
  },
};

export const WithStream: Story = {
  args: {
    sourceId: 'desktop-1',
    streamUrl: '/api/previews/desktop-1/mjpeg',
  },
  name: 'With Stream URL',
};
