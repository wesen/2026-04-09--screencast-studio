import React, { useRef, useCallback } from 'react';

interface SliderProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  className?: string;
}

export const Slider: React.FC<SliderProps> = ({
  value,
  onChange,
  min = 0,
  max = 100,
  className,
}) => {
  const ref = useRef<HTMLDivElement>(null);

  const drag = useCallback(
    (e: MouseEvent | TouchEvent) => {
      if (!ref.current) return;
      const rect = ref.current.getBoundingClientRect();
      const x =
        'touches' in e
          ? e.touches[0].clientX
          : e.clientX;
      const newValue = Math.max(
        min,
        Math.min(max, Math.round(min + ((x - rect.left) / rect.width) * (max - min)))
      );
      onChange(newValue);
    },
    [min, max, onChange]
  );

  const handleMouseDown = (e: React.MouseEvent) => {
    drag(e.nativeEvent);
    const handleMove = (v: MouseEvent) => drag(v);
    const handleUp = () => {
      window.removeEventListener('mousemove', handleMove);
      window.removeEventListener('mouseup', handleUp);
    };
    window.addEventListener('mousemove', handleMove);
    window.addEventListener('mouseup', handleUp);
  };

  const pct = ((value - min) / (max - min)) * 100;

  return (
    <div
      ref={ref}
      className={`studio-slider ${className || ''}`}
      onMouseDown={handleMouseDown}
    >
      <div className="studio-slider__track">
        <div
          className="studio-slider__fill"
          style={{ width: `${pct}%` }}
        />
      </div>
      <div
        className="studio-slider__thumb"
        style={{ left: `calc(${pct}% - 5px)` }}
      />
    </div>
  );
};
