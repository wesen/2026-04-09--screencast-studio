import React from 'react';
import { WinBar } from './WinBar';

interface WinProps {
  title: string;
  children: React.ReactNode;
  onClose?: () => void;
  className?: string;
  style?: React.CSSProperties;
}

export const Win: React.FC<WinProps> = ({
  title,
  children,
  onClose,
  className,
  style,
}) => {
  return (
    <div className={`studio-win ${className || ''}`} style={style}>
      <WinBar onClose={onClose}>{title}</WinBar>
      <div className="studio-win__content">{children}</div>
    </div>
  );
};
