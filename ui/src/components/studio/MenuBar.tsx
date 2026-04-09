import React, { useState, useEffect } from 'react';

interface MenuBarProps {
  armedCount: number;
  isRecording: boolean;
  isPaused: boolean;
  className?: string;
}

const MENU_ITEMS = ['File', 'Edit', 'Capture', 'Sources', 'Help'];

export const MenuBar: React.FC<MenuBarProps> = ({
  armedCount,
  isRecording,
  isPaused,
  className,
}) => {
  const [currentTime, setCurrentTime] = useState(new Date());

  useEffect(() => {
    const id = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <div className={`studio-menubar ${className || ''}`}>
      <span className="studio-menubar__logo">⌘</span>
      {MENU_ITEMS.map((m) => (
        <span key={m} className="studio-menubar__item">
          {m}
        </span>
      ))}
      <div className="studio-menubar__spacer" />
      <span className="studio-menubar__status">
        {armedCount} source{armedCount !== 1 ? 's' : ''} armed
      </span>
      {isRecording && !isPaused && (
        <span className="studio-menubar__rec studio-blink">● REC</span>
      )}
      <span className="studio-menubar__time">
        {currentTime.toLocaleTimeString([], {
          hour: '2-digit',
          minute: '2-digit',
        })}
      </span>
    </div>
  );
};
