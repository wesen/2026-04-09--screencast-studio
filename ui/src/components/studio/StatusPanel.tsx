import React from 'react';
import { Win } from '../primitives';
import type { StudioSource } from '@/components/source-card';

interface StatusPanelProps {
  diskPercent?: number;
  isRecording: boolean;
  isPaused: boolean;
  armedSources: StudioSource[];
  className?: string;
}

const getStatusText = (
  isRecording: boolean,
  isPaused: boolean
): { text: string; color: string } => {
  if (isRecording) {
    return isPaused
      ? { text: '⏸ Paused', color: 'var(--studio-amber)' }
      : { text: '● Recording', color: 'var(--studio-red)' };
  }
  return { text: '◻ Ready', color: 'var(--studio-dark)' };
};

export const StatusPanel: React.FC<StatusPanelProps> = ({
  diskPercent,
  isRecording,
  isPaused,
  armedSources,
  className,
}) => {
  const status = getStatusText(isRecording, isPaused);
  const hasDiskData = typeof diskPercent === 'number';
  const diskPct = hasDiskData ? Math.min(95, diskPercent) : 0;
  const diskRemaining = hasDiskData ? `${Math.round(100 - diskPct)}%` : 'n/a';

  return (
    <Win title="Status" className={className}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <div className="studio-status-item">
          <span className="studio-status-label">Disk</span>
          <div className="studio-status-bar">
          <div
              className={`studio-status-bar__fill ${diskPct > 85 ? 'studio-status-bar__fill--warning' : ''}`}
              style={{ width: `${diskPct}%` }}
            />
          </div>
          <span className="studio-status-value">{diskRemaining}</span>
        </div>
        {!hasDiskData ? (
          <div style={{ fontSize: '8px', color: 'var(--studio-mid)' }}>
            Disk telemetry unavailable
          </div>
        ) : null}
        <div style={{ fontSize: '9px', color: 'var(--studio-mid)' }}>
          Status:{' '}
          <span style={{ color: status.color, fontWeight: 'bold' }}>
            {status.text}
          </span>
        </div>
        <div style={{ fontSize: '8px', color: 'var(--studio-mid)' }}>
          Armed: {armedSources.map((s) => s.label).join(', ') || 'None'}
        </div>
      </div>
    </Win>
  );
};
