import React, { useState } from 'react';
import { Win } from '../primitives/Win';
import { Btn } from '../primitives/Btn';

interface DSLEditorProps {
  value: string;
  onChange: (value: string) => void;
  onCompile: () => void;
  isCompiling?: boolean;
  warnings?: string[];
  errors?: string[];
  className?: string;
}

export const DSLEditor: React.FC<DSLEditorProps> = ({
  value,
  onChange,
  onCompile,
  isCompiling = false,
  warnings = [],
  errors = [],
  className,
}) => {
  const [localValue, setLocalValue] = useState(value);

  const handleBlur = () => {
    if (localValue !== value) {
      onChange(localValue);
    }
  };

  const handleReset = () => {
    setLocalValue(value);
  };

  const hasChanges = localValue !== value;

  return (
    <Win title="Raw DSL (Debug)" className={className}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        <textarea
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          onBlur={handleBlur}
          spellCheck={false}
          style={{
            width: '100%',
            minHeight: 200,
            fontFamily: '"Monaco", monospace',
            fontSize: 11,
            padding: 8,
            border: '1.5px solid var(--studio-black)',
            borderRadius: 2,
            background: '#fafaf8',
            resize: 'vertical',
            boxSizing: 'border-box',
          }}
        />

        {/* Warnings */}
        {warnings.length > 0 && (
          <div
            style={{
              padding: '4px 8px',
              background: 'rgba(184, 152, 64, 0.1)',
              border: '1px solid var(--studio-amber)',
              borderRadius: 2,
              fontSize: 9,
              color: 'var(--studio-amber)',
            }}
          >
            <strong>Warnings:</strong>
            <ul style={{ margin: '4px 0 0 16px', padding: 0 }}>
              {warnings.map((w, i) => (
                <li key={i}>{w}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Errors */}
        {errors.length > 0 && (
          <div
            style={{
              padding: '4px 8px',
              background: 'rgba(192, 64, 64, 0.1)',
              border: '1px solid var(--studio-red)',
              borderRadius: 2,
              fontSize: 9,
              color: 'var(--studio-red)',
            }}
          >
            <strong>Errors:</strong>
            <ul style={{ margin: '4px 0 0 16px', padding: 0 }}>
              {errors.map((e, i) => (
                <li key={i}>{e}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Actions */}
        <div style={{ display: 'flex', gap: 8 }}>
          <Btn onClick={onCompile} disabled={isCompiling}>
            {isCompiling ? 'Compiling...' : 'Compile'}
          </Btn>
          {hasChanges && (
            <Btn onClick={handleReset}>Reset</Btn>
          )}
          {hasChanges && (
            <span
              style={{
                marginLeft: 'auto',
                fontSize: 9,
                color: 'var(--studio-amber)',
                alignSelf: 'center',
              }}
            >
              Unsaved changes
            </span>
          )}
        </div>
      </div>
    </Win>
  );
};
