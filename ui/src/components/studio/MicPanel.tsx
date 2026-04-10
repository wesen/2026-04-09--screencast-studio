import React from 'react';
import { Win, Sel, Slider } from '../primitives';
import { MicMeter } from '../MicMeter';
import { Waveform } from '../Waveform';

interface MicPanelProps {
  micLevel?: number;
  micInput: string;
  gain: number;
  isRecording: boolean;
  onMicInputChange: (input: string) => void;
  onGainChange: (gain: number) => void;
  className?: string;
}

export const MicPanel: React.FC<MicPanelProps> = ({
  micLevel,
  micInput,
  gain,
  isRecording,
  onMicInputChange,
  onGainChange,
  className,
}) => {
  const effectiveLevel = micLevel ?? 0;

  return (
    <Win title="Microphone" className={className}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 8 }}>L</span>
          <MicMeter level={effectiveLevel} />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 8 }}>R</span>
          <MicMeter level={effectiveLevel * 0.85} />
        </div>
        <Waveform active={isRecording} />
        {micLevel === undefined ? (
          <div style={{ fontSize: '8px', color: 'var(--studio-mid)' }}>
            Live meter unavailable
          </div>
        ) : null}
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ fontSize: '8px', color: 'var(--studio-mid)', width: 26 }}>Input</span>
          <Sel
            value={micInput}
            opts={['Built-in Mic', 'External', 'Line In']}
            onChange={onMicInputChange}
            width={95}
          />
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
