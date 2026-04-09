import React from 'react';

interface RadioProps {
  on: boolean;
  className?: string;
}

export const Radio: React.FC<RadioProps> = ({ on, className }) => (
  <span
    className={`studio-radio ${on ? '' : 'studio-radio--off'} ${className || ''}`}
  />
);
