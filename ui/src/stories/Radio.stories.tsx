import type { Meta, StoryObj } from '@storybook/react';
import { Radio } from '../components/primitives/Radio';

const meta = {
  title: 'Studio/Radio',
  component: Radio,
  tags: ['autodocs'],
  argTypes: {
    on: { control: 'boolean' },
  },
} satisfies Meta<typeof Radio>;

export default meta;
type Story = StoryObj<typeof meta>;

export const On: Story = {
  args: {
    on: true,
  },
};

export const Off: Story = {
  args: {
    on: false,
  },
};
