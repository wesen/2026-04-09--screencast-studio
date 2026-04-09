import React, { useRef, useEffect } from 'react';
import { Win } from '../primitives/Win';
import type { ProcessLog } from '@/api/types';

interface LogPanelProps {
  logs: ProcessLog[];
  maxLines?: number;
  autoScroll?: boolean;
  className?: string;
}

const STREAM_COLORS: Record<string, string> = {
  stdout: 'var(--studio-black)',
  stderr: 'var(--studio-red)',
};

const STREAM_PREFIXES: Record<string, string> = {
  stdout: 'O',
  stderr: 'E',
};

export const LogPanel: React.FC<LogPanelProps> = ({
  logs,
  maxLines = 500,
  autoScroll = true,
  className,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const displayedLogs = logs.slice(-maxLines);

  return (
    <Win title="Logs" className={className}>
      <div
        ref={containerRef}
        style={{
          maxHeight: 200,
          overflowY: 'auto',
          fontFamily: '"Monaco", monospace',
          fontSize: 9,
          background: '#1a1a1a',
          color: '#e8e4dc',
          padding: 4,
          borderRadius: 2,
        }}
      >
        {displayedLogs.length === 0 ? (
          <div
            style={{
              color: 'var(--studio-mid)',
              fontStyle: 'italic',
              padding: 4,
            }}
          >
            No logs yet
          </div>
        ) : (
          displayedLogs.map((log, index) => (
            <div
              key={index}
              style={{
                display: 'flex',
                gap: 8,
                padding: '1px 4px',
                borderBottom: '1px solid #2c2c2c',
              }}
            >
              <span
                style={{
                  color: STREAM_COLORS[log.stream] || 'var(--studio-mid)',
                  fontWeight: 'bold',
                  minWidth: 12,
                }}
              >
                {STREAM_PREFIXES[log.stream] || '?'}
              </span>
              <span style={{ color: 'var(--studio-mid)', flexShrink: 0 }}>
                {new Date(log.timestamp).toLocaleTimeString()}
              </span>
              <span style={{ color: 'var(--studio-mid)', flexShrink: 0 }}>
                {log.processLabel}
              </span>
              <span style={{ flex: 1, wordBreak: 'break-word' }}>
                {log.message}
              </span>
            </div>
          ))
        )}
      </div>
    </Win>
  );
};
