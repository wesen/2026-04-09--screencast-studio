import React, { useState } from 'react';
import { Btn } from '../primitives';
import type { SourceType } from '@/features/studio-draft/studioDraftSlice';

interface AddSourceButtonProps {
  onAdd: (kind: SourceType) => void;
  className?: string;
}

const SOURCE_TYPES: { kind: SourceType; icon: string; label: string }[] = [
  { kind: 'Display', icon: '🖥', label: 'Display' },
  { kind: 'Window', icon: '☐', label: 'Window' },
  { kind: 'Region', icon: '⊞', label: 'Region' },
  { kind: 'Camera', icon: '◉', label: 'Camera' },
];

export const AddSourceButton: React.FC<AddSourceButtonProps> = ({
  onAdd,
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div
      className={`studio-add-source ${className || ''}`}
      onClick={() => setIsOpen(!isOpen)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter') setIsOpen(!isOpen);
      }}
    >
      {!isOpen ? (
        <>
          <span className="studio-add-source__icon">+</span>
          <span className="studio-add-source__label">Add Source</span>
        </>
      ) : (
        <div
          className="studio-add-source__menu"
          onClick={(e) => e.stopPropagation()}
        >
          {SOURCE_TYPES.map(({ kind, icon, label }) => (
            <Btn
              key={kind}
              onClick={() => {
                onAdd(kind);
                setIsOpen(false);
              }}
              style={{ fontSize: '10px', textAlign: 'left' }}
            >
              {icon} {label}
            </Btn>
          ))}
        </div>
      )}
    </div>
  );
};
