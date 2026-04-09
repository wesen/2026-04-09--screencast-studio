import type { Meta, StoryObj } from '@storybook/react';
import { Win } from '../components/primitives/Win';

const meta = {
  title: 'Studio/Win',
  component: Win,
  tags: ['autodocs'],
  argTypes: {
    title: { control: 'text' },
  },
} satisfies Meta<typeof Win>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    title: 'Window',
    children: <div style={{ padding: 8 }}>Window content goes here</div>,
  },
};

export const Sources: Story = {
  args: {
    title: 'Sources (3)',
    children: (
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <div
          style={{
            width: 100,
            height: 80,
            background: '#e8e4dc',
            border: '1.5px solid #1a1a1a',
            borderRadius: 3,
          }}
        />
        <div
          style={{
            width: 100,
            height: 80,
            background: '#e8e4dc',
            border: '1.5px solid #1a1a1a',
            borderRadius: 3,
          }}
        />
      </div>
    ),
  },
};

export const Output: Story = {
  args: {
    title: 'Output Parameters',
    children: (
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
        <div style={{ fontSize: 9, color: '#8a8a7a' }}>Format:</div>
        <div>MOV</div>
        <div style={{ fontSize: 9, color: '#8a8a7a' }}>Quality:</div>
        <div>75%</div>
      </div>
    ),
  },
};

export const Closable: Story = {
  args: {
    title: 'Closable Window',
    onClose: () => console.log('close clicked'),
    children: <div style={{ padding: 8 }}>Click the ✕ to close</div>,
  },
};
