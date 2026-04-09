import React from 'react';
import { Btn } from '../primitives/Btn';

interface ErrorStateProps {
  title?: string;
  message?: string;
  onRetry?: () => void;
  className?: string;
}

export const ErrorState: React.FC<ErrorStateProps> = ({
  title = 'Error',
  message = 'Something went wrong',
  onRetry,
  className,
}) => {
  return (
    <div
      className={className}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 20,
        color: 'var(--studio-red)',
        textAlign: 'center',
      }}
    >
      <span style={{ fontSize: 20 }}>⚠</span>
      <span style={{ fontSize: 10, fontWeight: 'bold', marginTop: 4 }}>
        {title}
      </span>
      {message && (
        <span style={{ fontSize: 9, color: 'var(--studio-mid)', marginTop: 4 }}>
          {message}
        </span>
      )}
      {onRetry && (
        <Btn onClick={onRetry} style={{ marginTop: 8 }}>
          Retry
        </Btn>
      )}
    </div>
  );
};
