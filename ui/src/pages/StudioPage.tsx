import React, { useEffect, useState } from 'react';
import { useAppDispatch, useAppSelector } from '@/app/hooks';
import {
  addSource,
  removeSource,
  selectArmedSources,
  selectSources,
  setAudio,
  setFormat,
  setFps,
  setGain,
  setMicInput,
  setMultiTrack,
  setQuality,
  updateSource,
} from '@/features/studio-draft/studioDraftSlice';
import {
  selectSession,
  selectLogs,
  selectWsConnected,
} from '@/features/session/sessionSlice';
import { getWsClient } from '@/features/session/wsClient';
import { MenuBar, SourceGrid, OutputPanel, MicPanel, StatusPanel } from '@/components/studio';
import { LogPanel } from '@/components/log-panel';
import { DSLEditor } from '@/components/dsl-editor';
import { Btn } from '@/components/primitives/Btn';

type Tab = 'studio' | 'logs' | 'raw';

interface StudioPageProps {
  className?: string;
}

export const StudioPage: React.FC<StudioPageProps> = ({ className }) => {
  const dispatch = useAppDispatch();
  const sources = useAppSelector(selectSources);
  const armedSources = useAppSelector(selectArmedSources);
  const session = useAppSelector(selectSession);
  const logs = useAppSelector(selectLogs);
  const wsConnected = useAppSelector(selectWsConnected);

  const [activeTab, setActiveTab] = useState<Tab>('studio');
  const [isPaused, setIsPaused] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [diskPercent, setDiskPercent] = useState(8);
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

  useEffect(() => {
    const wsClient = getWsClient(dispatch);
    wsClient.connect();

    return () => {
      wsClient.disconnect();
    };
  }, [dispatch]);

  useEffect(() => {
    if (!isRecording || isPaused) {
      return;
    }

    const id = setInterval(() => {
      setElapsed((prev) => prev + 1);
    }, 1000);

    return () => clearInterval(id);
  }, [isPaused, isRecording]);

  useEffect(() => {
    if (!isRecording) {
      setDiskPercent(8);
      return;
    }

    const id = setInterval(() => {
      setDiskPercent((prev) => Math.min(95, prev + 0.2 * armedSources.length));
    }, 1000);

    return () => clearInterval(id);
  }, [armedSources.length, isRecording]);

  const outputSettings = useAppSelector((state) => ({
    format: state.studioDraft.format,
    fps: state.studioDraft.fps,
    quality: state.studioDraft.quality,
    audio: state.studioDraft.audio,
    multiTrack: state.studioDraft.multiTrack,
  }));

  const micSettings = useAppSelector((state) => ({
    micInput: state.studioDraft.micInput,
    gain: state.studioDraft.gain,
    micLevel: state.studioDraft.micLevel,
  }));

  const handleToggleRecording = () => {
    if (isRecording) {
      setIsPaused(false);
      setElapsed(0);
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
              onRemove={(id) => dispatch(removeSource(id))}
              onToggleArmed={(id) => {
                const source = sources.find((item) => item.id === id);
                if (source) {
                  dispatch(updateSource({ id, patch: { armed: !source.armed } }));
                }
              }}
              onToggleSolo={(id) => {
                const source = sources.find((item) => item.id === id);
                if (source) {
                  dispatch(updateSource({ id, patch: { solo: !source.solo } }));
                }
              }}
              onChangeScene={(id, scene) => {
                dispatch(updateSource({ id, patch: { scene } }));
              }}
              onAdd={(kind) => dispatch(addSource(kind))}
            />

            <div className="studio-content-row">
              <OutputPanel
                format={outputSettings.format}
                fps={outputSettings.fps}
                quality={outputSettings.quality}
                audio={outputSettings.audio}
                multiTrack={outputSettings.multiTrack}
                isRecording={isRecording}
                isPaused={isPaused}
                elapsed={elapsed}
                armedCount={armedSources.length}
                onFormatChange={(value) => dispatch(setFormat(value))}
                onFpsChange={(value) => dispatch(setFps(value))}
                onQualityChange={(value) => dispatch(setQuality(value))}
                onAudioChange={(value) => dispatch(setAudio(value))}
                onMultiTrackChange={(value) => dispatch(setMultiTrack(value))}
                onToggleRecording={handleToggleRecording}
                onTogglePause={handleTogglePause}
              />

              <div className="studio-panel-stack">
                <MicPanel
                  micLevel={micSettings.micLevel}
                  micInput={micSettings.micInput}
                  gain={micSettings.gain}
                  isRecording={isRecording}
                  onMicInputChange={(value) => dispatch(setMicInput(value))}
                  onGainChange={(value) => dispatch(setGain(value))}
                />

                <StatusPanel
                  diskPercent={diskPercent}
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

      {!wsConnected && (
        <div
          style={{
            position: 'fixed',
            bottom: 10,
            right: 10,
            background: 'var(--studio-amber)',
            color: 'var(--studio-cream)',
            padding: '4px 8px',
            borderRadius: 3,
            fontSize: 9,
          }}
        >
          Reconnecting to server...
        </div>
      )}
    </div>
  );
};
