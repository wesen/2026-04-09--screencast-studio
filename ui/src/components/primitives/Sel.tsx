import React from 'react';

interface SelProps {
  value: string;
  opts: string[];
  onChange: (value: string) => void;
  width?: number;
  className?: string;
}

export const Sel: React.FC<SelProps> = ({
  value,
  opts,
  onChange,
  width = 130,
  className,
}) => {
  const handleClick = () => {
    const i = opts.indexOf(value);
    onChange(opts[(i + 1) % opts.length]);
  };

  return (
    <div
      className={`studio-sel ${className || ''}`}
      onClick={handleClick}
      style={{ width }}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          handleClick();
        }
      }}
    >
      <span className="studio-sel__value">{value}</span>
      <span className="studio-sel__arrow">▼</span>
    </div>
  );
};
