import React from 'react';
import { Win, Slider } from '../primitives';
import { MicMeter } from '../MicMeter';
import { Waveform } from '../Waveform';

interface MicOption {
  value: string;
  label: string;
}

interface MicPanelProps {
  leftLevel?: number;
  rightLevel?: number;
  micInput: string;
  micOptions: MicOption[];
  gain: number;
  isRecording: boolean;
  onMicInputChange: (input: string) => void;
  onGainChange: (gain: number) => void;
  className?: string;
}

export const MicPanel: React.FC<MicPanelProps> = ({
  leftLevel,
  rightLevel,
  micInput,
  micOptions,
  gain,
  isRecording,
  onMicInputChange,
  onGainChange,
  className,
}) => {
  const effectiveLeftLevel = leftLevel ?? 0;
  const effectiveRightLevel = rightLevel ?? effectiveLeftLevel;

  return (
    <Win title="Microphone" className={className}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 8 }}>L</span>
          <MicMeter level={effectiveLeftLevel} />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 8 }}>R</span>
          <MicMeter level={effectiveRightLevel} />
        </div>
        <Waveform active={isRecording} />
        {leftLevel === undefined && rightLevel === undefined ? (
          <div style={{ fontSize: '8px', color: 'var(--studio-mid)' }}>
            Live meter unavailable
          </div>
        ) : null}
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 26 }}>Input</span>
          <select
            value={micInput}
            onChange={(event) => onMicInputChange(event.target.value)}
            className="studio-source-card__editor-input"
            style={{ width: 120 }}
          >
            {micOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 26 }}>Gain</span>
          <div style={{ flex: 1 }}>
            <Slider value={gain} onChange={onGainChange} />
          </div>
        </div>
      </div>
    </Win>
  );
};
