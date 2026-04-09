import type { Meta, StoryObj } from '@storybook/react';
import { EmptyState } from '../components/common/EmptyState';

const meta = {
  title: 'Studio/EmptyState',
  component: EmptyState,
  tags: ['autodocs'],
} satisfies Meta<typeof EmptyState>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const NoSources: Story = {
  args: {
    title: 'No Sources',
    message: 'Add a source to get started',
    icon: '🖥',
  },
};

export const NoLogs: Story = {
  args: {
    title: 'No Logs',
    message: 'Session logs will appear here',
    icon: '📋',
  },
};

export const NoWarnings: Story = {
  args: {
    title: 'All Good',
    message: 'No warnings or errors',
    icon: '✓',
  },
};
