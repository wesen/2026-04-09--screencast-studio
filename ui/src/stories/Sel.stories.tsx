import type { Meta, StoryObj } from '@storybook/react';
import { Sel } from '../components/primitives/Sel';

const meta = {
  title: 'Studio/Sel',
  component: Sel,
  tags: ['autodocs'],
} satisfies Meta<typeof Sel>;

export default meta;
type Story = StoryObj<typeof meta>;

const OPTIONS = ['Option A', 'Option B', 'Option C'];

export const Default: Story = {
  args: {
    value: 'Option A',
    opts: OPTIONS,
    onChange: () => {},
    width: 130,
  },
};

export const Narrow: Story = {
  args: {
    value: 'Option A',
    opts: OPTIONS,
    onChange: () => {},
    width: 80,
  },
};

export const Wide: Story = {
  args: {
    value: 'Option A with a very long label',
    opts: ['Option A with a very long label', 'Option B', 'Option C'],
    onChange: () => {},
    width: 200,
  },
};

export const Interactive: Story = {
  args: {
    value: 'FPS 24',
    opts: ['FPS 10', 'FPS 15', 'FPS 24', 'FPS 30'],
    onChange: () => console.log('changed'),
    width: 100,
  },
};

export const FpsOptions: Story = {
  args: {
    value: '24 fps',
    opts: ['10 fps', '15 fps', '24 fps', '30 fps'],
    onChange: () => {},
    width: 100,
  },
};

export const AudioOptions: Story = {
  args: {
    value: '48 kHz, 16-bit',
    opts: ['22 kHz, 8-bit', '44 kHz, 16-bit', '48 kHz, 16-bit'],
    onChange: () => {},
    width: 120,
  },
};

export const SaveLocation: Story = {
  args: {
    value: 'Macintosh HD',
    opts: ['Macintosh HD', 'Desktop', 'Documents'],
    onChange: () => {},
    width: 120,
  },
};
