import React, { useState } from 'react';
import { useAppSelector } from '@/app/hooks';
import { selectSources, selectArmedSources } from '@/features/studio-draft/studioDraftSlice';
import {
  selectSession,
  selectLogs,
  selectElapsed,
} from '@/features/session/sessionSlice';
import { MenuBar, SourceGrid, OutputPanel, MicPanel, StatusPanel } from '@/components/studio';
import { LogPanel } from '@/components/log-panel';
import { DSLEditor } from '@/components/dsl-editor';
import { Btn } from '@/components/primitives/Btn';

type Tab = 'studio' | 'logs' | 'raw';

interface StudioPageProps {
  className?: string;
}

export const StudioPage: React.FC<StudioPageProps> = ({ className }) => {
  const sources = useAppSelector(selectSources);
  const armedSources = useAppSelector(selectArmedSources);
  const session = useAppSelector(selectSession);
  const logs = useAppSelector(selectLogs);
  const elapsed = useAppSelector(selectElapsed);

  const [activeTab, setActiveTab] = useState<Tab>('studio');
  const [isPaused, setIsPaused] = useState(false);
  const [dslText, setDslText] = useState(`schema: recorder.config/v1
session_id: demo
destination_templates:
  video: recordings/{session_id}/{name}.mov
video_sources:
  - id: desktop-1
    name: Full Desktop
    type: display
    target:
      display: display-1
    settings:
      capture:
        fps: 24
      output:
        container: mov
        video_codec: h264
        quality: 75
audio_sources:
  - id: mic-1
    name: Built-in Mic
    device: default
`);

  const isRecording = session.active && session.state === 'running';

  const handleToggleRecording = () => {
    if (isRecording) {
      setIsPaused(false);
    }
  };

  const handleTogglePause = () => {
    setIsPaused(!isPaused);
  };

  return (
    <div className={className} style={{ minHeight: '100vh' }}>
      <MenuBar
        armedCount={armedSources.length}
        isRecording={isRecording}
        isPaused={isPaused}
      />

      {/* Tab Bar */}
      <div
        style={{
          display: 'flex',
          gap: 4,
          padding: '4px 10px',
          background: 'var(--studio-cream)',
          borderBottom: '1px solid var(--studio-light)',
        }}
      >
        {(['studio', 'logs', 'raw'] as const).map((tab) => (
          <Btn
            key={tab}
            active={activeTab === tab}
            onClick={() => setActiveTab(tab)}
            style={{ fontSize: 9, padding: '2px 8px' }}
          >
            {tab === 'studio' ? 'Studio' : tab === 'logs' ? 'Logs' : 'Raw DSL'}
          </Btn>
        ))}
      </div>

      {/* Content */}
      <div className="studio-main">
        {activeTab === 'studio' && (
          <>
            <SourceGrid
              sources={sources}
              isRecording={isRecording}
              onRemove={(id) => console.log('remove', id)}
              onToggleArmed={(id) => console.log('toggle armed', id)}
              onToggleSolo={(id) => console.log('toggle solo', id)}
              onChangeScene={(id, scene) => console.log('change scene', id, scene)}
              onAdd={(kind) => console.log('add source', kind)}
            />

            <div className="studio-content-row">
              <OutputPanel
                format="MOV"
                fps="24 fps"
                quality={75}
                audio="48 kHz, 16-bit"
                multiTrack={true}
                isRecording={isRecording}
                isPaused={isPaused}
                elapsed={elapsed}
                armedCount={armedSources.length}
                onFormatChange={(f) => console.log('format', f)}
                onFpsChange={(f) => console.log('fps', f)}
                onQualityChange={(q) => console.log('quality', q)}
                onAudioChange={(a) => console.log('audio', a)}
                onMultiTrackChange={(m) => console.log('multitrack', m)}
                onToggleRecording={handleToggleRecording}
                onTogglePause={handleTogglePause}
              />

              <div className="studio-panel-stack">
                <MicPanel
                  micLevel={0.35}
                  micInput="Built-in Mic"
                  gain={55}
                  isRecording={isRecording}
                  onMicInputChange={(m) => console.log('mic input', m)}
                  onGainChange={(g) => console.log('gain', g)}
                />

                <StatusPanel
                  diskPercent={15}
                  isRecording={isRecording}
                  isPaused={isPaused}
                  armedSources={armedSources}
                />
              </div>
            </div>
          </>
        )}

        {activeTab === 'logs' && (
          <LogPanel logs={logs} />
        )}

        {activeTab === 'raw' && (
          <DSLEditor
            value={dslText}
            onChange={setDslText}
            onCompile={() => console.log('compile', dslText)}
            isCompiling={false}
            warnings={session.warnings}
          />
        )}
      </div>
    </div>
  );
};
