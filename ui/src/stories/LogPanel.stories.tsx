import type { Meta, StoryObj } from '@storybook/react';
import { LogPanel } from '../components/log-panel/LogPanel';
import type { ProcessLog } from '../api/types';

const meta = {
  title: 'Studio/LogPanel',
  component: LogPanel,
  tags: ['autodocs'],
} satisfies Meta<typeof LogPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const createLog = (
  stream: ProcessLog['stream'],
  message: string,
  offsetMs = 0
): ProcessLog => ({
  timestamp: new Date(Date.now() - offsetMs).toISOString(),
  process_label: 'preview-1',
  stream,
  message,
});

export const Empty: Story = {
  args: {
    logs: [],
  },
};

export const WithLogs: Story = {
  args: {
    logs: [
      createLog('stdout', 'Starting screencast studio...', 10000),
      createLog('stdout', 'Discovering displays...', 9000),
      createLog('stdout', 'Found 2 displays', 8000),
      createLog('stdout', 'Discovering windows...', 7000),
      createLog('stdout', 'Found 15 windows', 6000),
      createLog('stderr', 'Window "Untitled - LibreOffice Writer" has no title', 5000),
      createLog('stdout', 'Compiling setup...', 4000),
      createLog('stdout', 'Plan generated with 1 video job, 1 audio job', 3000),
      createLog('stdout', 'Starting recording session', 2000),
      createLog('stdout', 'Worker desktop-1 started', 1000),
      createLog('stdout', 'Worker mic-1 started', 500),
    ],
  },
};

export const WithErrors: Story = {
  args: {
    logs: [
      createLog('stdout', 'Starting recording...', 5000),
      createLog('stderr', 'FFmpeg exited with code 1', 4000),
      createLog('stderr', 'Output file not writable: permission denied', 3000),
      createLog('stderr', 'Retrying...', 2000),
      createLog('stdout', 'Recording stopped', 1000),
    ],
  },
};

export const ManyLogs: Story = {
  args: {
    logs: Array.from({ length: 100 }, (_, i) =>
      createLog(
        i % 10 === 0 ? 'stderr' : 'stdout',
        `Log entry ${i} with some additional text to make it wrap`,
        (100 - i) * 100
      )
    ),
  },
};
