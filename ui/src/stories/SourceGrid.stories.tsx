import type { Meta, StoryObj } from '@storybook/react';
import { SourceGrid } from '../components/studio/SourceGrid';
import type { Source } from '../features/studio-draft/studioDraftSlice';

const meta = {
  title: 'Studio/SourceGrid',
  component: SourceGrid,
  tags: ['autodocs'],
} satisfies Meta<typeof SourceGrid>;

export default meta;
type Story = StoryObj<typeof meta>;

const createSource = (id: number, overrides: Partial<Source> = {}): Source => ({
  id,
  kind: 'Display',
  scene: 'Desktop 1',
  armed: true,
  solo: false,
  label: 'Display 1',
  ...overrides,
});

export const Empty: Story = {
  args: {
    sources: [],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onToggleSolo: () => {},
    onChangeScene: () => {},
    onAdd: () => {},
  },
};

export const SingleSource: Story = {
  args: {
    sources: [createSource(1, { kind: 'Display', label: 'Display 1' })],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onToggleSolo: () => {},
    onChangeScene: () => {},
    onAdd: () => {},
  },
};

export const MultipleSources: Story = {
  args: {
    sources: [
      createSource(1, { kind: 'Display', label: 'Display 1' }),
      createSource(2, { kind: 'Camera', label: 'Camera 1', scene: 'Built-in' }),
      createSource(3, { kind: 'Window', label: 'Terminal', scene: 'Terminal' }),
    ],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onToggleSolo: () => {},
    onChangeScene: () => {},
    onAdd: () => {},
  },
};

export const AllSourceTypes: Story = {
  args: {
    sources: [
      createSource(1, { kind: 'Display', label: 'Display 1' }),
      createSource(2, { kind: 'Window', label: 'Window 1', scene: 'Finder' }),
      createSource(3, { kind: 'Region', label: 'Region 1', scene: 'Top Half' }),
      createSource(4, { kind: 'Camera', label: 'Camera 1', scene: 'Built-in' }),
    ],
    isRecording: false,
    onRemove: () => {},
    onToggleArmed: () => {},
    onToggleSolo: () => {},
    onChangeScene: () => {},
    onAdd: () => {},
  },
};

export const WhileRecording: Story = {
  args: {
    sources: [
      createSource(1, { kind: 'Display', label: 'Display 1', armed: true }),
      createSource(2, { kind: 'Display', label: 'Display 2', armed: false }),
    ],
    isRecording: true,
    onRemove: () => {},
    onToggleArmed: () => {},
    onToggleSolo: () => {},
    onChangeScene: () => {},
    onAdd: () => {},
  },
};
