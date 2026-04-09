import type { Meta, StoryObj } from '@storybook/react';
import { ErrorState } from '../components/common/ErrorState';

const meta = {
  title: 'Studio/ErrorState',
  component: ErrorState,
  tags: ['autodocs'],
} satisfies Meta<typeof ErrorState>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const WithRetry: Story = {
  args: {
    onRetry: () => console.log('retry clicked'),
  },
};

export const DiscoveryFailed: Story = {
  args: {
    title: 'Discovery Failed',
    message: 'Could not connect to X11. Is DISPLAY set?',
    onRetry: () => console.log('retry clicked'),
  },
};

export const CompileError: Story = {
  args: {
    title: 'Compile Error',
    message: 'Invalid YAML: unexpected token at line 5',
    onRetry: () => console.log('retry clicked'),
  },
};

export const RecordingError: Story = {
  args: {
    title: 'Recording Failed',
    message: 'FFmpeg process terminated unexpectedly',
    onRetry: () => console.log('retry clicked'),
  },
};
