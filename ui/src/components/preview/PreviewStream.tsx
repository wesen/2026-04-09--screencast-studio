import React, { useMemo, useState } from 'react';

interface PreviewStreamProps {
  sourceId: string;
  state?: string;
  reason?: string;
  streamUrl?: string;
  className?: string;
  aspectRatio?: number;
}

export const PreviewStream: React.FC<PreviewStreamProps> = ({
  sourceId,
  state,
  reason,
  streamUrl,
  className,
  aspectRatio,
}) => {
  const [error, setError] = useState(false);
  const [naturalAspectRatio, setNaturalAspectRatio] = useState<number | null>(null);
  const effectiveAspectRatio = useMemo(() => {
    if (naturalAspectRatio && Number.isFinite(naturalAspectRatio) && naturalAspectRatio > 0) {
      return naturalAspectRatio;
    }
    if (aspectRatio && Number.isFinite(aspectRatio) && aspectRatio > 0) {
      return aspectRatio;
    }
    return 4 / 3;
  }, [aspectRatio, naturalAspectRatio]);

  if (!streamUrl || error) {
    const message = error
      ? 'Preview stream failed'
      : state === 'starting'
        ? 'Preview starting'
        : state === 'stopping'
          ? 'Preview stopping'
          : state === 'failed'
            ? reason || 'Preview failed'
            : 'Preview unavailable';

    return (
      <div
        className={className}
        style={{
          width: '100%',
          aspectRatio: String(effectiveAspectRatio),
          background: 'var(--studio-black)',
          borderRadius: 2,
          border: '1.5px solid var(--studio-dark)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'var(--studio-mid)',
          fontSize: 9,
          textAlign: 'center',
          padding: '0 12px',
        }}
      >
        {message}
      </div>
    );
  }

  return (
    <div
      className={className}
      style={{
        width: '100%',
        aspectRatio: String(effectiveAspectRatio),
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 2,
        border: '1.5px solid var(--studio-dark)',
        background: 'var(--studio-black)',
      }}
    >
      <img
        src={streamUrl}
        alt={`Preview of ${sourceId}`}
        style={{
          width: '100%',
          height: '100%',
          objectFit: 'contain',
          display: 'block',
          background: 'var(--studio-black)',
        }}
        onLoad={(event) => {
          const { naturalWidth, naturalHeight } = event.currentTarget;
          if (naturalWidth > 0 && naturalHeight > 0) {
            setNaturalAspectRatio(naturalWidth / naturalHeight);
          }
        }}
        onError={() => setError(true)}
      />
    </div>
  );
};
