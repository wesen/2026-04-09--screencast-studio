import React from 'react';
import { Win, Radio, Sel, Slider, Btn } from '../primitives';

interface OutputPanelProps {
  format: 'MOV' | 'AVI' | 'MP4';
  fps: string;
  quality: number;
  audio: string;
  multiTrack: boolean;
  isRecording: boolean;
  isPaused: boolean;
  elapsed: number;
  armedCount: number;
  onFormatChange: (format: 'MOV' | 'AVI' | 'MP4') => void;
  onFpsChange: (fps: string) => void;
  onQualityChange: (quality: number) => void;
  onAudioChange: (audio: string) => void;
  onMultiTrackChange: (multiTrack: boolean) => void;
  onToggleRecording: () => void;
  onTogglePause: () => void;
  className?: string;
}

const formatTime = (seconds: number): string => {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
};

const formatSize = (seconds: number, armedCount: number): string => {
  return `~${(seconds * 0.4 * armedCount).toFixed(1)} MB`;
};

export const OutputPanel: React.FC<OutputPanelProps> = ({
  format,
  fps,
  quality,
  audio,
  multiTrack,
  isRecording,
  isPaused,
  elapsed,
  armedCount,
  onFormatChange,
  onFpsChange,
  onQualityChange,
  onAudioChange,
  onMultiTrackChange,
  onToggleRecording,
  onTogglePause,
  className,
}) => {
  return (
    <Win title="Output Parameters" className={className}>
      <div className="studio-output-grid">
        <span className="studio-output-label">Format:</span>
        <div style={{ display: 'flex', gap: 8 }}>
          {(['MOV', 'AVI', 'MP4'] as const).map((f) => (
            <span
              key={f}
              className="studio-output-radio"
              onClick={() => onFormatChange(f)}
            >
              <Radio on={format === f} />
              {f}
            </span>
          ))}
        </div>
        <span className="studio-output-label">Framerate:</span>
        <Sel
          value={fps}
          opts={['10 fps', '15 fps', '24 fps', '30 fps']}
          onChange={onFpsChange}
          width={100}
        />

        <span className="studio-output-label">Quality:</span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <div style={{ flex: 1 }}>
            <Slider value={quality} onChange={onQualityChange} />
          </div>
          <span style={{ fontSize: '9px', minWidth: 24 }}>{quality}%</span>
        </div>
        <span className="studio-output-label">Audio:</span>
        <Sel
          value={audio}
          opts={['22 kHz, 8-bit', '44 kHz, 16-bit', '48 kHz, 16-bit']}
          onChange={onAudioChange}
          width={120}
        />

        <span className="studio-output-label">Multi-track:</span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <Btn
            active={multiTrack}
            onClick={() => onMultiTrackChange(!multiTrack)}
            style={{ fontSize: '9px', padding: '2px 8px' }}
          >
            {multiTrack ? '◉ Each source → own file' : '○ Merge all sources'}
          </Btn>
        </div>
        <span className="studio-output-label">Save to:</span>
        <Sel
          value="Macintosh HD"
          opts={['Macintosh HD', 'Desktop', 'Documents']}
          onChange={() => {}}
          width={120}
        />
      </div>

      {/* Transport */}
      <div className="studio-transport">
        <Btn
          accent={!isRecording}
          active={isRecording}
          onClick={onToggleRecording}
          style={!isRecording ? { color: 'var(--studio-cream)', background: 'var(--studio-red)' } : {}}
        >
          {isRecording ? '◼ Stop' : '● Rec'}
        </Btn>
        <Btn
          active={isPaused}
          onClick={onTogglePause}
          disabled={!isRecording}
        >
          {isPaused ? '▶ Resume' : '❚❚ Pause'}
        </Btn>
        <div style={{ flex: 1, textAlign: 'center' }}>
          {isRecording && (
            <span style={{ fontSize: '8px', color: 'var(--studio-mid)' }}>
              {multiTrack ? `${armedCount} file${armedCount !== 1 ? 's' : ''}:` : 'merged:'} {formatSize(elapsed, armedCount)}
            </span>
          )}
        </div>
        <span
          className={`studio-transport__time ${isRecording && !isPaused ? 'studio-transport__time--recording' : ''}`}
        >
          {formatTime(elapsed)}
        </span>
      </div>
    </Win>
  );
};
