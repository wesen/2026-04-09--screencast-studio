import React from 'react';
import type { StudioSourceKind } from '@/components/source-card/types';

interface FakeScreenProps {
  kind: StudioSourceKind;
  scene?: string;
  className?: string;
}

export const FakeScreen: React.FC<FakeScreenProps> = ({
  kind,
  scene,
  className,
}) => {
  if (kind === 'Camera') {
    return (
      <div className={`studio-fake-screen ${className || ''}`}>
        <svg
          width="40"
          height="50"
          viewBox="0 0 40 50"
          fill="none"
          stroke="#606058"
          strokeWidth="1.2"
        >
          <ellipse cx="20" cy="14" rx="9" ry="10" />
          <circle cx="16" cy="12" r="1.2" fill="#707068" />
          <circle cx="24" cy="12" r="1.2" fill="#707068" />
          <path d="M16 18Q20 22 24 18" />
          <path d="M8 28Q6 38 5 48L35 48Q34 38 32 28Q28 25 20 25Q12 25 8 28Z" />
        </svg>
        <div className="studio-fake-screen__scanlines" />
        <div className="studio-fake-screen__label">{scene}</div>
      </div>
    );
  }

  if (kind === 'Region') {
    return (
      <div className={`studio-fake-screen ${className || ''}`}>
        <div className="studio-fake-screen__scanlines" />
        <div
          style={{
            position: 'absolute',
            top: '15%',
            left: '10%',
            width: '80%',
            height: '70%',
            border: '1px dashed var(--studio-amber)',
            borderRadius: '1px',
          }}
        />
        <div
          style={{
            position: 'absolute',
            top: '15%',
            left: '10%',
            fontFamily: 'var(--studio-font)',
            fontSize: '7px',
            color: 'var(--studio-amber)',
            padding: '1px 3px',
          }}
        >
          {scene}
        </div>
        {['top-left', 'top-right', 'bottom-left', 'bottom-right'].map((c) => {
          const [v, h] = c.split('-') as [string, string];
          return (
            <div
              key={c}
              style={{
                position: 'absolute',
                [v]: v === 'top' ? 'calc(15% - 3px)' : 'calc(85% - 3px)',
                [h]: h === 'left' ? 'calc(10% - 3px)' : 'calc(90% - 3px)',
                width: 6,
                height: 6,
                border: '1px solid var(--studio-amber)',
                background: 'rgba(184, 152, 64, 0.3)',
              }}
            />
          );
        })}
      </div>
    );
  }

  // Display / Window
  return (
    <div className={`studio-fake-screen ${className || ''}`}>
      <div className="studio-fake-screen__scanlines" />
      <div
        style={{
          position: 'absolute',
          top: 5,
          left: 6,
          right: 6,
        }}
      >
        <div
          style={{
            fontFamily: 'var(--studio-font)',
            fontSize: '8px',
            color: '#909088',
            display: 'flex',
            justifyContent: 'space-between',
            marginBottom: 3,
          }}
        >
          <span>■ {kind === 'Window' ? scene : 'Finder'}</span>
          <span style={{ fontSize: '7px' }}>12:00</span>
        </div>
        <div
          style={{
            border: '1px solid #505048',
            borderRadius: '1px',
            padding: 3,
          }}
        >
          <div
            style={{
              display: 'flex',
              gap: 6,
              flexWrap: 'wrap',
            }}
          >
            {(kind === 'Window'
              ? ['src', 'out', 'lib']
              : ['System', 'Apps', 'Docs']
            ).map((f) => (
              <div
                key={f}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                }}
              >
                <div
                  style={{
                    width: 12,
                    height: 9,
                    border: '1px solid #606058',
                    borderRadius: '1px',
                    marginBottom: 1,
                  }}
                />
                <span
                  style={{
                    fontFamily: 'var(--studio-font)',
                    fontSize: '6px',
                    color: '#707068',
                  }}
                >
                  {f}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="studio-fake-screen__label">{scene || kind}</div>
    </div>
  );
};
