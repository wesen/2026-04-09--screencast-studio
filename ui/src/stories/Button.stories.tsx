import type { Meta, StoryObj } from '@storybook/react';
import { Btn } from '../components/primitives/Btn';

const meta = {
  title: 'Studio/Button',
  component: Btn,
  tags: ['autodocs'],
  argTypes: {
    active: { control: 'boolean' },
    accent: { control: 'boolean' },
    disabled: { control: 'boolean' },
  },
} satisfies Meta<typeof Btn>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: 'Button',
  },
};

export const Active: Story = {
  args: {
    children: 'Active Button',
    active: true,
  },
};

export const Accent: Story = {
  args: {
    children: 'Record',
    accent: true,
  },
};

export const Disabled: Story = {
  args: {
    children: 'Disabled',
    disabled: true,
  },
};
