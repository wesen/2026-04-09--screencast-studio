import React, { useState, useEffect } from 'react';

interface WaveformProps {
  active: boolean;
  className?: string;
}

export const Waveform: React.FC<WaveformProps> = ({ active, className }) => {
  const [bars, setBars] = useState<number[]>(Array(24).fill(2));

  useEffect(() => {
    if (!active) {
      setBars(Array(24).fill(2));
      return;
    }

    const id = setInterval(() => {
      setBars((prev) => prev.map(() => 2 + Math.random() * 14));
    }, 90);

    return () => clearInterval(id);
  }, [active]);

  return (
    <div className={`studio-waveform ${className || ''}`}>
      {bars.map((height, i) => (
        <div
          key={i}
          className={`studio-waveform__bar ${active ? 'studio-waveform__bar--active' : ''}`}
          style={{ height }}
        />
      ))}
    </div>
  );
};
