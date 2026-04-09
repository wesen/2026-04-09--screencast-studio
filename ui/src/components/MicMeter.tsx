import React from 'react';

interface MicMeterProps {
  level: number; // 0-1
  className?: string;
}

export const MicMeter: React.FC<MicMeterProps> = ({ level, className }) => {
  const barCount = 16;

  return (
    <div className={`studio-mic-meter ${className || ''}`}>
      {Array.from({ length: barCount }).map((_, i) => {
        const p = i / barCount;
        const on = p < level;
        const colorClass =
          on
            ? p > 0.75
              ? 'studio-mic-meter__bar--danger'
              : p > 0.55
              ? 'studio-mic-meter__bar--warning'
              : 'studio-mic-meter__bar--on'
            : '';

        return (
          <div
            key={i}
            className={`studio-mic-meter__bar ${colorClass}`}
          />
        );
      })}
    </div>
  );
};
