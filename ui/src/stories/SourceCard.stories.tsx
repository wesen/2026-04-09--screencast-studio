import type { Meta, StoryObj } from '@storybook/react';
import { SourceCard } from '../components/source-card/SourceCard';
import type { StudioSource } from '../components/source-card';

const meta = {
  title: 'Studio/SourceCard',
  component: SourceCard,
  tags: ['autodocs'],
} satisfies Meta<typeof SourceCard>;

export default meta;
type Story = StoryObj<typeof meta>;

const createSource = (overrides: Partial<StudioSource> = {}): StudioSource => ({
  id: 'display-main',
  sourceId: 'display-main',
  kind: 'Display',
  scene: 'Desktop 1',
  armed: true,
  label: 'Display 1',
  ...overrides,
});

const defaultHandlers = {
  onRemove: () => {},
  onToggleArmed: () => {},
  onChangeScene: () => {},
  onMoveUp: () => {},
  onMoveDown: () => {},
};

export const DisplayArmed: Story = {
  args: {
    source: createSource({ kind: 'Display', label: 'Display 1' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const DisplayDisarmed: Story = {
  args: {
    source: createSource({ kind: 'Display', label: 'Display 1', armed: false }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const WindowTerminal: Story = {
  args: {
    source: createSource({ kind: 'Window', label: 'Terminal', scene: 'Terminal' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const WindowFinder: Story = {
  args: {
    source: createSource({ kind: 'Window', label: 'Finder', scene: 'Finder' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const WindowBrowser: Story = {
  args: {
    source: createSource({ kind: 'Window', label: 'Browser', scene: 'Browser' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const WindowCodeEditor: Story = {
  args: {
    source: createSource({ kind: 'Window', label: 'Code Editor', scene: 'Code Editor' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const RegionTopHalf: Story = {
  args: {
    source: createSource({ kind: 'Region', label: 'Region 1', scene: 'Top Half' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const RegionBottomHalf: Story = {
  args: {
    source: createSource({ kind: 'Region', label: 'Region 1', scene: 'Bottom Half' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const RegionCustom: Story = {
  args: {
    source: createSource({ kind: 'Region', label: 'Region 1', scene: 'Custom Region' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const CameraBuiltin: Story = {
  args: {
    source: createSource({ kind: 'Camera', label: 'Camera 1', scene: 'Built-in' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const CameraUSB: Story = {
  args: {
    source: createSource({ kind: 'Camera', label: 'USB Camera', scene: 'USB Camera' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const CameraFaceTime: Story = {
  args: {
    source: createSource({ kind: 'Camera', label: 'FaceTime HD', scene: 'FaceTime HD' }),
    isRecording: false,
    ...defaultHandlers,
  },
};

export const WhileRecordingArmed: Story = {
  args: {
    source: createSource({ kind: 'Display', label: 'Display 1', armed: true }),
    isRecording: true,
    ...defaultHandlers,
  },
};

export const WhileRecordingDisarmed: Story = {
  args: {
    source: createSource({ kind: 'Display', label: 'Display 2', armed: false }),
    isRecording: true,
    ...defaultHandlers,
  },
};

export const ReadOnlyNormalized: Story = {
  args: {
    source: createSource({
      id: 'window-browser',
      sourceId: 'window-browser',
      kind: 'Window',
      label: 'Firefox',
      scene: 'Firefox',
    }),
    isRecording: false,
    editable: false,
  },
};
