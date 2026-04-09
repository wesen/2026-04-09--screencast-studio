import React from 'react';
import { Btn, Sel } from '../primitives';
import { FakeScreen } from '../FakeScreen';
import type { Source, SourceType } from '@/features/studio-draft/studioDraftSlice';

interface SourceCardProps {
  source: Source;
  isRecording: boolean;
  onRemove: () => void;
  onToggleArmed: () => void;
  onToggleSolo: () => void;
  onChangeScene: (scene: string) => void;
  className?: string;
}

const ICON_MAP: Record<SourceType, string> = {
  Display: '🖥',
  Window: '☐',
  Region: '⊞',
  Camera: '◉',
};

const SCENE_MAP: Record<SourceType, string[]> = {
  Display: ['Desktop 1', 'Desktop 2'],
  Window: ['Finder', 'Terminal', 'Browser', 'Code Editor'],
  Region: ['Top Half', 'Bottom Half', 'Custom Region'],
  Camera: ['Built-in', 'USB Camera', 'FaceTime HD'],
};

export const SourceCard: React.FC<SourceCardProps> = ({
  source,
  isRecording,
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
        <div
          className="studio-winbar__close"
          onClick={onRemove}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter') onRemove();
          }}
        >
          ✕
        </div>
        <span className="studio-source-card__title">
          {icon} {source.label}
        </span>
        <div style={{ width: 10 }} />
      </div>
      <div className="studio-source-card__body">
        <FakeScreen kind={source.kind} scene={source.scene} />
        <div className="studio-source-card__preview">
          <Sel
            value={source.scene}
            opts={scenes}
            onChange={onChangeScene}
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
        </div>
      </div>
    </div>
  );
};
