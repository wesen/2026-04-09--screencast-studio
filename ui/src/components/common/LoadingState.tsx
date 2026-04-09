import React from 'react';

interface LoadingStateProps {
  message?: string;
  className?: string;
}

export const LoadingState: React.FC<LoadingStateProps> = ({
  message = 'Loading...',
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
      }}
    >
      <div
        style={{
          width: 24,
          height: 24,
          border: '2px solid var(--studio-light)',
          borderTopColor: 'var(--studio-black)',
          borderRadius: '50%',
          animation: 'spin 1s linear infinite',
        }}
      />
      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
      {message && (
        <span style={{ marginTop: 8, fontSize: 10 }}>{message}</span>
      )}
    </div>
  );
};
