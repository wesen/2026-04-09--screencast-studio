import type { Meta, StoryObj } from '@storybook/react';
import { FakeScreen } from '../components/FakeScreen';

const meta = {
  title: 'Studio/FakeScreen',
  component: FakeScreen,
  tags: ['autodocs'],
  argTypes: {
    kind: { control: 'select' },
    scene: { control: 'text' },
  },
} satisfies Meta<typeof FakeScreen>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Display: Story = {
  args: {
    kind: 'Display',
    scene: 'Desktop 1',
  },
};

export const Display2: Story = {
  args: {
    kind: 'Display',
    scene: 'Desktop 2',
  },
};

export const Window: Story = {
  args: {
    kind: 'Window',
    scene: 'Terminal',
  },
};

export const WindowFinder: Story = {
  args: {
    kind: 'Window',
    scene: 'Finder',
  },
};

export const WindowBrowser: Story = {
  args: {
    kind: 'Window',
    scene: 'Browser',
  },
};

export const WindowCodeEditor: Story = {
  args: {
    kind: 'Window',
    scene: 'Code Editor',
  },
};

export const Region: Story = {
  args: {
    kind: 'Region',
    scene: 'Top Half',
  },
};

export const RegionCustom: Story = {
  args: {
    kind: 'Region',
    scene: 'Custom Region',
  },
};

export const RegionBottom: Story = {
  args: {
    kind: 'Region',
    scene: 'Bottom Half',
  },
};

export const Camera: Story = {
  args: {
    kind: 'Camera',
    scene: 'Built-in',
  },
};

export const CameraUSB: Story = {
  args: {
    kind: 'Camera',
    scene: 'USB Camera',
  },
};

export const CameraFaceTime: Story = {
  args: {
    kind: 'Camera',
    scene: 'FaceTime HD',
  },
};
