import React from 'react';

interface WinBarProps {
  children: React.ReactNode;
  onClose?: () => void;
  className?: string;
}

export const WinBar: React.FC<WinBarProps> = ({
  children,
  onClose,
  className,
}) => {
  return (
    <div className={`studio-winbar ${className || ''}`}>
      {onClose && (
        <div
          className="studio-winbar__close"
          onClick={onClose}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter') onClose();
          }}
        >
          ✕
        </div>
      )}
      <div className="studio-winbar__title">{children}</div>
      <div className="studio-winbar__spacer" />
    </div>
  );
};
