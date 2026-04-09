import React from 'react';
import { Win } from '../primitives';
import { SourceCard, AddSourceButton } from '../source-card';
import type { Source, SourceType } from '@/features/studio-draft/studioDraftSlice';

interface SourceGridProps {
  sources: Source[];
  isRecording: boolean;
  onRemove: (id: number) => void;
  onToggleArmed: (id: number) => void;
  onToggleSolo: (id: number) => void;
  onChangeScene: (id: number, scene: string) => void;
  onAdd: (kind: SourceType) => void;
  className?: string;
}

export const SourceGrid: React.FC<SourceGridProps> = ({
  sources,
  isRecording,
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
            onRemove={() => onRemove(source.id)}
            onToggleArmed={() => onToggleArmed(source.id)}
            onToggleSolo={() => onToggleSolo(source.id)}
            onChangeScene={(scene) => onChangeScene(source.id, scene)}
          />
        ))}
        <AddSourceButton onAdd={onAdd} />
      </div>
    </Win>
  );
};
