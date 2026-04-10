import type { Meta, StoryObj } from '@storybook/react';
import { SourceGrid } from '../components/studio/SourceGrid';
import type { StudioSource } from '../components/source-card';

const meta = {
  title: 'Studio/SourceGrid',
  component: SourceGrid,
  tags: ['autodocs'],
} satisfies Meta<typeof SourceGrid>;

export default meta;
type Story = StoryObj<typeof meta>;

const createSource = (
  id: string,
  overrides: Partial<StudioSource> = {}
): StudioSource => ({
  id,
  sourceId: id,
  kind: 'Display',
  scene: 'Desktop 1',
  armed: true,
  label: 'Display 1',
  ...overrides,
});

export const Empty: Story = {
  args: {
    sources: [],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onChangeScene: () => {},
    onMoveUp: () => {},
    onMoveDown: () => {},
    onAdd: () => {},
  },
};

export const SingleSource: Story = {
  args: {
    sources: [createSource('display-main', { kind: 'Display', label: 'Display 1' })],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onChangeScene: () => {},
    onMoveUp: () => {},
    onMoveDown: () => {},
    onAdd: () => {},
  },
};

export const MultipleSources: Story = {
  args: {
    sources: [
      createSource('display-main', { kind: 'Display', label: 'Display 1' }),
      createSource('camera-face', { kind: 'Camera', label: 'Camera 1', scene: 'Built-in' }),
      createSource('window-terminal', { kind: 'Window', label: 'Terminal', scene: 'Terminal' }),
    ],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onChangeScene: () => {},
    onMoveUp: () => {},
    onMoveDown: () => {},
    onAdd: () => {},
  },
};

export const AllSourceTypes: Story = {
  args: {
    sources: [
      createSource('display-main', { kind: 'Display', label: 'Display 1' }),
      createSource('window-finder', { kind: 'Window', label: 'Window 1', scene: 'Finder' }),
      createSource('region-top-half', { kind: 'Region', label: 'Region 1', scene: 'Top Half' }),
      createSource('camera-main', { kind: 'Camera', label: 'Camera 1', scene: 'Built-in' }),
    ],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onChangeScene: () => {},
    onMoveUp: () => {},
    onMoveDown: () => {},
    onAdd: () => {},
  },
};

export const WhileRecording: Story = {
  args: {
    sources: [
      createSource('display-main', { kind: 'Display', label: 'Display 1', armed: true }),
      createSource('display-secondary', { kind: 'Display', label: 'Display 2', armed: false }),
    ],
    isRecording: true,
    onRemove: () => {},
    onToggleArmed: () => {},
    onChangeScene: () => {},
    onMoveUp: () => {},
    onMoveDown: () => {},
    onAdd: () => {},
  },
};

export const ReadOnlyNormalized: Story = {
  args: {
    sources: [
      createSource('display-main', { kind: 'Display', label: 'DELL U2720Q', scene: 'DELL U2720Q' }),
      createSource('window-browser', { kind: 'Window', label: 'Firefox', scene: 'Firefox' }),
      createSource('camera-main', { kind: 'Camera', label: 'USB Camera', scene: 'USB Camera' }),
    ],
    isRecording: false,
    editable: false,
  },
};
