import type { Meta, StoryObj } from '@storybook/react';
import { LogPanel } from '../components/log-panel/LogPanel';
import type { LogEntry } from '../features/session/sessionSlice';

const meta = {
  title: 'Studio/LogPanel',
  component: LogPanel,
  tags: ['autodocs'],
} satisfies Meta<typeof LogPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const createLog = (
  level: LogEntry['level'],
  message: string,
  offsetMs = 0
): LogEntry => ({
  timestamp: Date.now() - offsetMs,
  level,
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
      createLog('info', 'Starting screencast studio...', 10000),
      createLog('info', 'Discovering displays...', 9000),
      createLog('info', 'Found 2 displays', 8000),
      createLog('info', 'Discovering windows...', 7000),
      createLog('info', 'Found 15 windows', 6000),
      createLog('warn', 'Window "Untitled - LibreOffice Writer" has no title', 5000),
      createLog('info', 'Compiling setup...', 4000),
      createLog('info', 'Plan generated with 1 video job, 1 audio job', 3000),
      createLog('info', 'Starting recording session', 2000),
      createLog('info', 'Worker desktop-1 started', 1000),
      createLog('info', 'Worker mic-1 started', 500),
    ],
  },
};

export const WithErrors: Story = {
  args: {
    logs: [
      createLog('info', 'Starting recording...', 5000),
      createLog('error', 'FFmpeg exited with code 1', 4000),
      createLog('error', 'Output file not writable: permission denied', 3000),
      createLog('warn', 'Retrying...', 2000),
      createLog('info', 'Recording stopped', 1000),
    ],
  },
};

export const ManyLogs: Story = {
  args: {
    logs: Array.from({ length: 100 }, (_, i) =>
      createLog(
        i % 10 === 0 ? 'warn' : i % 20 === 0 ? 'error' : 'info',
        `Log entry ${i} with some additional text to make it wrap`,
        (100 - i) * 100
      )
    ),
  },
};
