import type { Meta, StoryObj } from '@storybook/react';
import { AddSourceButton } from '../components/source-card/AddSourceButton';

const meta = {
  title: 'Studio/AddSourceButton',
  component: AddSourceButton,
  tags: ['autodocs'],
} satisfies Meta<typeof AddSourceButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    onAdd: (kind) => console.log('Add source:', kind),
  },
};
