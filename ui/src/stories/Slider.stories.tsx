import type { Meta, StoryObj } from '@storybook/react';
import { Slider } from '../components/primitives/Slider';

const meta = {
  title: 'Studio/Slider',
  component: Slider,
  tags: ['autodocs'],
} satisfies Meta<typeof Slider>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    value: 50,
    onChange: () => {},
    min: 0,
    max: 100,
  },
};

export const Low: Story = {
  args: {
    value: 10,
    onChange: () => {},
    min: 0,
    max: 100,
  },
};

export const High: Story = {
  args: {
    value: 90,
    onChange: () => {},
    min: 0,
    max: 100,
  },
};

export const Quality: Story = {
  args: {
    value: 75,
    onChange: () => {},
    min: 0,
    max: 100,
  },
};

export const Gain: Story = {
  args: {
    value: 55,
    onChange: () => {},
    min: 0,
    max: 100,
  },
};

export const Interactive: Story = {
  args: {
    value: 50,
    onChange: () => console.log('changed'),
    min: 0,
    max: 100,
  },
};
