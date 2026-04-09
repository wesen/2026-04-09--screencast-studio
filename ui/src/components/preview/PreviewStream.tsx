import React, { useState } from 'react';

interface PreviewStreamProps {
  sourceId: string;
  streamUrl?: string;
  className?: string;
}

export const PreviewStream: React.FC<PreviewStreamProps> = ({
  sourceId,
  streamUrl,
  className,
}) => {
  const [error, setError] = useState(false);

  if (!streamUrl || error) {
    return (
      <div
        className={className}
        style={{
          width: '100%',
          aspectRatio: '4/3',
          background: 'var(--studio-black)',
          borderRadius: 2,
          border: '1.5px solid var(--studio-dark)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'var(--studio-mid)',
          fontSize: 9,
        }}
      >
        Preview unavailable
      </div>
    );
  }

  return (
    <div
      className={className}
      style={{
        width: '100%',
        aspectRatio: '4/3',
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 2,
        border: '1.5px solid var(--studio-dark)',
      }}
    >
      <img
        src={streamUrl}
        alt={`Preview of ${sourceId}`}
        style={{
          width: '100%',
          height: '100%',
          objectFit: 'cover',
        }}
        onError={() => setError(true)}
      />
    </div>
  );
};
