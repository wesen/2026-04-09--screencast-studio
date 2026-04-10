import React from 'react';
import type {
  CameraDescriptor,
  DisplayDescriptor,
  WindowDescriptor,
} from '@/api/types';
import type { RegionPreset } from '@/features/setup-draft/conversion';
import { Btn, Win } from '../primitives';
import type { StudioSourceKind } from '../source-card';

interface SourcePickerProps {
  kind: StudioSourceKind;
  displays: DisplayDescriptor[];
  windows: WindowDescriptor[];
  cameras: CameraDescriptor[];
  onClose: () => void;
  onPickDisplay: (display: DisplayDescriptor) => void;
  onPickWindow: (window: WindowDescriptor) => void;
  onPickCamera: (camera: CameraDescriptor) => void;
  onPickRegion: (display: DisplayDescriptor, preset: RegionPreset) => void;
}

const REGION_PRESETS: { preset: RegionPreset; label: string }[] = [
  { preset: 'full', label: 'Full Display' },
  { preset: 'top-half', label: 'Top Half' },
  { preset: 'bottom-half', label: 'Bottom Half' },
  { preset: 'left-half', label: 'Left Half' },
  { preset: 'right-half', label: 'Right Half' },
];

export const SourcePicker: React.FC<SourcePickerProps> = ({
  kind,
  displays,
  windows,
  cameras,
  onClose,
  onPickDisplay,
  onPickWindow,
  onPickCamera,
  onPickRegion,
}) => {
  const title = kind === 'Display'
    ? 'Add Display Source'
    : kind === 'Window'
      ? 'Add Window Source'
      : kind === 'Region'
        ? 'Add Region Source'
        : 'Add Camera Source';

  return (
    <Win title={title} onClose={onClose} className="studio-source-picker">
      {kind === 'Display' && (
        <div className="studio-source-picker__list">
          {displays.map((display) => (
            <Btn
              key={display.id}
              onClick={() => onPickDisplay(display)}
              style={{ width: '100%', justifyContent: 'flex-start' }}
            >
              {display.name} {display.primary ? '(Primary)' : ''}
            </Btn>
          ))}
          {displays.length === 0 && (
            <div className="studio-source-picker__empty">No displays discovered.</div>
          )}
        </div>
      )}

      {kind === 'Window' && (
        <div className="studio-source-picker__list">
          {windows.map((window) => (
            <Btn
              key={window.id}
              onClick={() => onPickWindow(window)}
              style={{ width: '100%', justifyContent: 'flex-start' }}
            >
              {window.title || window.id}
            </Btn>
          ))}
          {windows.length === 0 && (
            <div className="studio-source-picker__empty">No windows discovered.</div>
          )}
        </div>
      )}

      {kind === 'Camera' && (
        <div className="studio-source-picker__list">
          {cameras.map((camera) => (
            <Btn
              key={camera.id}
              onClick={() => onPickCamera(camera)}
              style={{ width: '100%', justifyContent: 'flex-start' }}
            >
              {camera.label || camera.device}
            </Btn>
          ))}
          {cameras.length === 0 && (
            <div className="studio-source-picker__empty">No cameras discovered.</div>
          )}
        </div>
      )}

      {kind === 'Region' && (
        <div className="studio-source-picker__regions">
          {displays.map((display) => (
            <div key={display.id} className="studio-source-picker__region-group">
              <div className="studio-source-picker__region-title">
                {display.name} {display.primary ? '(Primary)' : ''}
              </div>
              <div className="studio-source-picker__region-buttons">
                {REGION_PRESETS.map((preset) => (
                  <Btn
                    key={`${display.id}-${preset.preset}`}
                    onClick={() => onPickRegion(display, preset.preset)}
                    style={{ fontSize: '9px', padding: '3px 6px' }}
                  >
                    {preset.label}
                  </Btn>
                ))}
              </div>
            </div>
          ))}
          {displays.length === 0 && (
            <div className="studio-source-picker__empty">No displays discovered.</div>
          )}
        </div>
      )}
    </Win>
  );
};
