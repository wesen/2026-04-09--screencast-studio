import type { Meta, StoryObj } from '@storybook/react';
import { DSLEditor } from '../components/dsl-editor/DSLEditor';
import { DEFAULT_DSL_TEXT } from '../features/editor/editorSlice';

const meta = {
  title: 'Studio/DSLEditor',
  component: DSLEditor,
  tags: ['autodocs'],
} satisfies Meta<typeof DSLEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

const sampleDSL = DEFAULT_DSL_TEXT;

export const Default: Story = {
  args: {
    value: sampleDSL,
    onChange: () => {},
    onApply: () => {},
    onReset: () => {},
    isApplying: false,
    hasChanges: false,
    warnings: [],
    errors: [],
  },
};

export const WithWarnings: Story = {
  args: {
    value: sampleDSL,
    onChange: () => {},
    onApply: () => {},
    onReset: () => {},
    isApplying: false,
    hasChanges: false,
    warnings: [
      'No audio sources defined',
      'Quality set to 75, consider higher for production',
    ],
    errors: [],
  },
};

export const WithErrors: Story = {
  args: {
    value: `schema: recorder.config/v1
invalid yaml here`,
    onChange: () => {},
    onApply: () => {},
    onReset: () => {},
    isApplying: false,
    hasChanges: true,
    warnings: [],
    errors: [
      'YAML parse error at line 2: indentation mismatch',
      'Schema version "v1" not found',
    ],
  },
};

export const Compiling: Story = {
  args: {
    value: sampleDSL,
    onChange: () => {},
    onApply: () => {},
    onReset: () => {},
    isApplying: true,
    hasChanges: true,
    warnings: [],
    errors: [],
  },
};

export const WithChanges: Story = {
  args: {
    value: sampleDSL,
    onChange: () => {},
    onApply: () => {},
    onReset: () => {},
    isApplying: false,
    hasChanges: true,
    warnings: [],
    errors: [],
  },
  name: 'With Unsaved Changes',
};
