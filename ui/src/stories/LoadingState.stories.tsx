import type { Meta, StoryObj } from '@storybook/react';
import { LoadingState } from '../components/common/LoadingState';

const meta = {
  title: 'Studio/LoadingState',
  component: LoadingState,
  tags: ['autodocs'],
} satisfies Meta<typeof LoadingState>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const WithMessage: Story = {
  args: {
    message: 'Loading sources...',
  },
};

export const FetchingDiscovery: Story = {
  args: {
    message: 'Discovering displays and windows...',
  },
};

export const Compiling: Story = {
  args: {
    message: 'Compiling setup...',
  },
};
