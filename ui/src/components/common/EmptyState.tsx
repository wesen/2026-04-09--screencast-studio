import React from 'react';

interface EmptyStateProps {
  title?: string;
  message?: string;
  icon?: string;
  className?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'No data',
  message,
  icon = '○',
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
        color: 'var(--studio-mid)',
        textAlign: 'center',
      }}
    >
      <span style={{ fontSize: 24 }}>{icon}</span>
      <span style={{ fontSize: 10, fontWeight: 'bold', marginTop: 4 }}>
        {title}
      </span>
      {message && (
        <span style={{ fontSize: 9, marginTop: 4 }}>{message}</span>
      )}
    </div>
  );
};
