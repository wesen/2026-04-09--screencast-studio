import React from 'react';

interface BtnProps {
  children: React.ReactNode;
  active?: boolean;
  accent?: boolean;
  solo?: boolean;
  onClick?: () => void;
  disabled?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

export const Btn: React.FC<BtnProps> = ({
  children,
  active = false,
  accent = false,
  solo = false,
  onClick,
  disabled = false,
  className,
  style,
}) => {
  const classes = [
    'studio-btn',
    active ? 'studio-btn--active' : '',
    accent ? 'studio-btn--accent' : '',
    solo ? 'studio-btn--solo' : '',
    className || '',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <button
      className={classes}
      onClick={onClick}
      disabled={disabled}
      style={style}
    >
      {children}
    </button>
  );
};
