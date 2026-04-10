import React from 'react';
import { Win, Radio, Sel, Slider, Btn } from '../primitives';

interface OutputPreview {
  kind: string;
  sourceId?: string;
  name: string;
  path: string;
}

interface OutputPanelProps {
  recordingName: string;
  destinationRoot: string;
  destinationRootEditable: boolean;
  destinationRootReason?: string;
  filenameSuffix: string;
  filenameSuffixEditable: boolean;
  filenameSuffixReason?: string;
  outputs: OutputPreview[];
  outputPreviewBusy?: boolean;
  outputPreviewErrors?: string[];
  format: 'MOV' | 'AVI' | 'MP4';
  fps: string;
  quality: number;
  audio: string;
  multiTrack: boolean;
  isRecording: boolean;
  isPaused: boolean;
  pauseSupported?: boolean;
  transportBusy?: boolean;
  elapsed: number;
  armedCount: number;
  onRecordingNameChange: (value: string) => void;
  onDestinationRootChange: (value: string) => void;
  onFilenameSuffixChange: (value: string) => void;
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

const formatSize = (seconds: number, armedCount: number): string => `~${(seconds * 0.4 * armedCount).toFixed(1)} MB`;

export const OutputPanel: React.FC<OutputPanelProps> = ({
  recordingName,
  destinationRoot,
  destinationRootEditable,
  destinationRootReason,
  filenameSuffix,
  filenameSuffixEditable,
  filenameSuffixReason,
  outputs,
  outputPreviewBusy = false,
  outputPreviewErrors = [],
  format,
  fps,
  quality,
  audio,
  multiTrack,
  isRecording,
  isPaused,
  pauseSupported = true,
  transportBusy = false,
  elapsed,
  armedCount,
  onRecordingNameChange,
  onDestinationRootChange,
  onFilenameSuffixChange,
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
        <span className="studio-output-label">Name:</span>
        <input
          value={recordingName}
          onChange={(event) => onRecordingNameChange(event.target.value)}
          className="studio-source-card__editor-input"
          style={{ width: 160 }}
        />
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

        <span className="studio-output-label">Save to:</span>
        <input
          value={destinationRoot}
          onChange={(event) => onDestinationRootChange(event.target.value)}
          disabled={!destinationRootEditable}
          className="studio-source-card__editor-input"
          style={{ width: 160 }}
        />
        <span className="studio-output-label">Framerate:</span>
        <Sel
          value={fps}
          opts={['10 fps', '15 fps', '24 fps', '30 fps']}
          onChange={onFpsChange}
          width={100}
        />

        <span className="studio-output-label">Filename:</span>
        <input
          value={filenameSuffix}
          onChange={(event) => onFilenameSuffixChange(event.target.value)}
          disabled={!filenameSuffixEditable}
          className="studio-source-card__editor-input"
          style={{ width: 160 }}
          placeholder="-{date}-{index}"
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
      </div>

      {destinationRootReason ? (
        <div style={{ fontSize: '8px', color: 'var(--studio-amber)', marginBottom: 6 }}>
          {destinationRootReason}
        </div>
      ) : null}
      {filenameSuffixReason && !destinationRootReason ? (
        <div style={{ fontSize: '8px', color: 'var(--studio-amber)', marginBottom: 6 }}>
          {filenameSuffixReason}
        </div>
      ) : null}
      <div style={{ fontSize: '8px', color: 'var(--studio-mid)', marginBottom: 8, lineHeight: 1.4 }}>
        Tokens can be used in <code>Name</code>, <code>Save to</code>, and <code>Filename</code>.
        {' '}Video files use <code>{'{source_name}'}</code>
        {filenameSuffix || '<suffix>'}
        <code>.{'{ext}'}</code> and mixed audio uses <code>audio-mix</code>
        {filenameSuffix || '<suffix>'}
        <code>.{'{ext}'}</code>.
        {' '}Tokens: <code>{'{date}'}</code>, <code>{'{time}'}</code>, <code>{'{timestamp}'}</code>, <code>{'{index}'}</code>.
      </div>

      <div style={{ marginBottom: 8 }}>
        <div style={{ fontSize: '9px', color: 'var(--studio-mid)', marginBottom: 4 }}>
          Planned outputs {outputPreviewBusy ? '…refreshing' : ''}
        </div>
        {outputPreviewErrors.length > 0 ? (
          <div style={{ fontSize: '8px', color: 'var(--studio-red)' }}>
            {outputPreviewErrors.join(' ')}
          </div>
        ) : outputs.length === 0 ? (
          <div style={{ fontSize: '8px', color: 'var(--studio-mid)' }}>
            No planned outputs yet
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 3, maxHeight: 96, overflowY: 'auto' }}>
            {outputs.map((output) => (
              <div key={`${output.kind}:${output.sourceId}:${output.path}`} style={{ fontSize: '8px', color: 'var(--studio-dark)' }}>
                <span style={{ color: 'var(--studio-mid)', marginRight: 6 }}>
                  {output.kind === 'audio' ? 'Audio' : output.name}
                </span>
                <code>{output.path}</code>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="studio-transport">
        <Btn
          accent={!isRecording}
          active={isRecording}
          onClick={onToggleRecording}
          disabled={transportBusy}
          style={!isRecording ? { color: 'var(--studio-cream)', background: 'var(--studio-red)' } : {}}
        >
          {transportBusy ? '…' : isRecording ? '◼ Stop' : '● Rec'}
        </Btn>
        <Btn
          active={isPaused}
          onClick={onTogglePause}
          disabled={!isRecording || !pauseSupported}
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
