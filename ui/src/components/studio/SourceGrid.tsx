import React from 'react';
import { Win } from '../primitives';
import { AddSourceButton, SourceCard } from '../source-card';
import type { StudioSource, StudioSourceKind } from '../source-card';

interface SourceGridProps {
  sources: StudioSource[];
  isRecording: boolean;
  editable?: boolean;
  onRemove?: (id: string) => void;
  onToggleArmed?: (id: string) => void;
  onToggleSolo?: (id: string) => void;
  onChangeScene?: (id: string, scene: string) => void;
  onAdd?: (kind: StudioSourceKind) => void;
  className?: string;
}

export const SourceGrid: React.FC<SourceGridProps> = ({
  sources,
  isRecording,
  editable = true,
  onRemove,
  onToggleArmed,
  onToggleSolo,
  onChangeScene,
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
            onRemove={onRemove ? () => onRemove(source.id) : undefined}
            onToggleArmed={onToggleArmed ? () => onToggleArmed(source.id) : undefined}
            onToggleSolo={onToggleSolo ? () => onToggleSolo(source.id) : undefined}
            onChangeScene={onChangeScene ? (scene) => onChangeScene(source.id, scene) : undefined}
          />
        ))}
        {editable && onAdd ? <AddSourceButton onAdd={onAdd} /> : null}
      </div>
    </Win>
  );
};
