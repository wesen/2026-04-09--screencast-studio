import React from 'react';
import { Btn } from '../primitives';
import { PreviewStream } from '../preview';
import type { StudioSource, StudioSourceKind } from './types';

interface SourceCardProps {
  source: StudioSource;
  isRecording: boolean;
  editable?: boolean;
  onRemove?: () => void;
  onToggleArmed?: () => void;
  onChangeScene?: (scene: string) => void;
  onMoveUp?: () => void;
  onMoveDown?: () => void;
  className?: string;
}

const ICON_MAP: Record<StudioSourceKind, string> = {
  Display: '🖥',
  Window: '☐',
  Region: '⊞',
  Camera: '◉',
};

export const SourceCard: React.FC<SourceCardProps> = ({
  source,
  isRecording,
  editable = true,
  onRemove,
  onToggleArmed,
  onChangeScene,
  onMoveUp,
  onMoveDown,
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
        <PreviewStream
          sourceId={source.sourceId}
          state={source.previewState}
          reason={source.previewReason}
          streamUrl={source.previewUrl}
        />
        <div className="studio-source-card__preview">
          {editable ? (
            <>
              <input
                className="studio-source-card__name-input"
                value={source.scene}
                onChange={(event) => onChangeScene?.(event.target.value)}
              />
              {source.detail ? (
                <div className="studio-source-card__detail">{source.detail}</div>
              ) : null}
              <div className="studio-source-card__toolbar">
                <Btn
                  active={source.armed}
                  onClick={onToggleArmed}
                  style={{ fontSize: '9px', padding: '2px 6px' }}
                >
                  {source.armed ? '◉ Enabled' : '○ Disabled'}
                </Btn>
                <Btn
                  onClick={onMoveUp}
                  disabled={!onMoveUp}
                  style={{ fontSize: '9px', padding: '2px 6px' }}
                >
                  ↑
                </Btn>
                <Btn
                  onClick={onMoveDown}
                  disabled={!onMoveDown}
                  style={{ fontSize: '9px', padding: '2px 6px' }}
                >
                  ↓
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
