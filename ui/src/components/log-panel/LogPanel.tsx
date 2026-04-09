import React, { useRef, useEffect } from 'react';
import { Win } from '../primitives/Win';
import type { LogEntry } from '@/features/session/sessionSlice';

interface LogPanelProps {
  logs: LogEntry[];
  maxLines?: number;
  autoScroll?: boolean;
  className?: string;
}

const LEVEL_COLORS: Record<string, string> = {
  debug: 'var(--studio-mid)',
  info: 'var(--studio-black)',
  warn: 'var(--studio-amber)',
  error: 'var(--studio-red)',
};

const LEVEL_PREFIXES: Record<string, string> = {
  debug: 'D',
  info: 'I',
  warn: 'W',
  error: 'E',
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
                  color: LEVEL_COLORS[log.level] || 'var(--studio-mid)',
                  fontWeight: 'bold',
                  minWidth: 12,
                }}
              >
                {LEVEL_PREFIXES[log.level] || '?'}
              </span>
              <span style={{ color: 'var(--studio-mid)', flexShrink: 0 }}>
                {new Date(log.timestamp).toLocaleTimeString()}
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
