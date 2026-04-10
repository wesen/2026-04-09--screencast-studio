import React from 'react';
import { Win } from '../primitives';
import { AddSourceButton, SourceCard } from '../source-card';
import type { StudioSource, StudioSourceKind } from '../source-card';

interface SourceGridProps {
  sources: StudioSource[];
  isRecording: boolean;
  editable?: boolean;
  renderEditor?: (source: StudioSource) => React.ReactNode;
  onRemove?: (id: string) => void;
  onToggleArmed?: (id: string) => void;
  onChangeScene?: (id: string, scene: string) => void;
  onMoveUp?: (id: string) => void;
  onMoveDown?: (id: string) => void;
  onAdd?: (kind: StudioSourceKind) => void;
  className?: string;
}

export const SourceGrid: React.FC<SourceGridProps> = ({
  sources,
  isRecording,
  editable = true,
  renderEditor,
  onRemove,
  onToggleArmed,
  onChangeScene,
  onMoveUp,
  onMoveDown,
  onAdd,
  className,
}) => {
  return (
    <Win title={`Sources (${sources.length})`} className={className}>
      <div className="studio-grid">
        {sources.map((source) => (
          <SourceCard
            key={source.id}
            source={source}
            isRecording={isRecording}
            editable={editable}
            editor={renderEditor?.(source)}
            onRemove={onRemove ? () => onRemove(source.id) : undefined}
            onToggleArmed={onToggleArmed ? () => onToggleArmed(source.id) : undefined}
            onChangeScene={onChangeScene ? (scene) => onChangeScene(source.id, scene) : undefined}
            onMoveUp={onMoveUp ? () => onMoveUp(source.id) : undefined}
            onMoveDown={onMoveDown ? () => onMoveDown(source.id) : undefined}
          />
        ))}
        {onAdd ? <AddSourceButton onAdd={onAdd} /> : null}
      </div>
    </Win>
  );
};
