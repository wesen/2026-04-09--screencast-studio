import React from 'react';
import { Btn, Sel } from '../primitives';
import { FakeScreen } from '../FakeScreen';
import type { StudioSource, StudioSourceKind } from './types';

interface SourceCardProps {
  source: StudioSource;
  isRecording: boolean;
  editable?: boolean;
  onRemove?: () => void;
  onToggleArmed?: () => void;
  onToggleSolo?: () => void;
  onChangeScene?: (scene: string) => void;
  className?: string;
}

const ICON_MAP: Record<StudioSourceKind, string> = {
  Display: '🖥',
  Window: '☐',
  Region: '⊞',
  Camera: '◉',
};

const SCENE_MAP: Record<StudioSourceKind, string[]> = {
  Display: ['Desktop 1', 'Desktop 2'],
  Window: ['Finder', 'Terminal', 'Browser', 'Code Editor'],
  Region: ['Top Half', 'Bottom Half', 'Custom Region'],
  Camera: ['Built-in', 'USB Camera', 'FaceTime HD'],
};

export const SourceCard: React.FC<SourceCardProps> = ({
  source,
  isRecording,
  editable = true,
  onRemove,
  onToggleArmed,
  onToggleSolo,
  onChangeScene,
  className,
}) => {
  const classes = [
    'studio-source-card',
    !source.armed ? 'studio-source-card--disarmed' : '',
    isRecording && !source.armed ? '' : '',
    className || '',
  ]
    .filter(Boolean)
    .join(' ');

  const headerClasses = [
    'studio-source-card__header',
    source.armed ? 'studio-source-card__header--armed' : 'studio-source-card__header--disarmed',
  ].join(' ');

  const scenes = SCENE_MAP[source.kind];
  const icon = ICON_MAP[source.kind];

  return (
    <div className={classes}>
      <div className={headerClasses}>
        {editable ? (
          <div
            className="studio-winbar__close"
            onClick={onRemove}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && onRemove) onRemove();
            }}
          >
            ✕
          </div>
        ) : (
          <div style={{ width: 10 }} />
        )}
        <span className="studio-source-card__title">
          {icon} {source.label}
        </span>
        <div style={{ width: 10 }} />
      </div>
      <div className="studio-source-card__body">
        <FakeScreen kind={source.kind} scene={source.scene} />
        <div className="studio-source-card__preview">
          {editable ? (
            <>
              <Sel
                value={source.scene}
                opts={scenes}
                onChange={(value) => onChangeScene?.(value)}
                width={175}
              />
              <div className="studio-source-card__toolbar">
                <Btn
                  active={source.armed}
                  onClick={onToggleArmed}
                  style={{ fontSize: '9px', padding: '2px 0' }}
                >
                  {source.armed ? '◉ Armed' : '○ Disarmed'}
                </Btn>
                <Btn
                  active={source.solo}
                  onClick={onToggleSolo}
                  style={{
                    fontSize: '9px',
                    padding: '2px 6px',
                    color: source.solo ? 'var(--studio-cream)' : 'var(--studio-amber)',
                    background: source.solo ? 'var(--studio-amber)' : 'var(--studio-cream)',
                  }}
                >
                  S
                </Btn>
              </div>
            </>
          ) : (
            <div className="studio-source-card__toolbar">
              <Btn
                active={source.armed}
                disabled
                style={{ fontSize: '9px', padding: '2px 8px' }}
              >
                {source.armed ? '◉ Enabled' : '○ Disabled'}
              </Btn>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
