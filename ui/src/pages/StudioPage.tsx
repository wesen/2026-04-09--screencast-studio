import React, { useEffect, useMemo, useState } from 'react';
import {
  useGetCurrentSessionQuery,
  useStartRecordingMutation,
  useStopRecordingMutation,
} from '@/api/recordingApi';
import { useCompileSetupMutation } from '@/api/setupApi';
import type { ApiErrorResponse, ProcessLog } from '@/api/types';
import { useAppDispatch, useAppSelector } from '@/app/hooks';
import {
  compileFailed,
  compileStarted,
  compileSucceeded,
  selectCompileErrors,
  selectCompileWarnings,
  selectDslText,
  selectIsCompiling,
  setDslText,
} from '@/features/editor/editorSlice';
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
import { selectActiveTab, setActiveTab, type StudioTab } from '@/features/studio-ui/studioUiSlice';
import {
  addLog,
  selectSession,
  selectLogs,
  setSession,
  selectWsConnected,
} from '@/features/session/sessionSlice';
import { getWsClient } from '@/features/session/wsClient';
import { MenuBar, SourceGrid, OutputPanel, MicPanel, StatusPanel } from '@/components/studio';
import { LogPanel } from '@/components/log-panel';
import { DSLEditor } from '@/components/dsl-editor';
import { Btn } from '@/components/primitives/Btn';

interface StudioPageProps {
  className?: string;
}

const parseTimestamp = (value?: string): number | null => {
  if (!value) {
    return null;
  }
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) {
    return null;
  }
  return parsed;
};

const errorMessageFromUnknown = (error: unknown): string => {
  if (typeof error === 'object' && error !== null && 'data' in error) {
    const data = (error as { data?: unknown }).data;
    if (
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as ApiErrorResponse).error?.message === 'string'
    ) {
      return (data as ApiErrorResponse).error.message;
    }
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'unknown error';
};

export const StudioPage: React.FC<StudioPageProps> = ({ className }) => {
  const dispatch = useAppDispatch();
  const sources = useAppSelector(selectSources);
  const armedSources = useAppSelector(selectArmedSources);
  const session = useAppSelector(selectSession);
  const logs = useAppSelector(selectLogs);
  const wsConnected = useAppSelector(selectWsConnected);
  const activeTab = useAppSelector(selectActiveTab);
  const dslText = useAppSelector(selectDslText);
  const compileWarnings = useAppSelector(selectCompileWarnings);
  const compileErrors = useAppSelector(selectCompileErrors);
  const isCompiling = useAppSelector(selectIsCompiling);
  const [now, setNow] = useState(() => Date.now());
  const { data: currentSessionData } = useGetCurrentSessionQuery();
  const [startRecording, startRecordingState] = useStartRecordingMutation();
  const [stopRecording, stopRecordingState] = useStopRecordingMutation();
  const [compileSetup] = useCompileSetupMutation();

  const isRecording = session.active;
  const isPaused = false;
  const diskPercent = 8;
  const transportBusy =
    startRecordingState.isLoading ||
    stopRecordingState.isLoading ||
    session.state === 'starting' ||
    session.state === 'stopping';

  useEffect(() => {
    if (currentSessionData?.session) {
      dispatch(setSession(currentSessionData.session));
    }
  }, [currentSessionData, dispatch]);

  useEffect(() => {
    const wsClient = getWsClient(dispatch);
    wsClient.connect();

    return () => {
      wsClient.disconnect();
    };
  }, [dispatch]);

  useEffect(() => {
    const startedAt = parseTimestamp(session.started_at);
    if (!isRecording || startedAt === null) {
      return;
    }

    const id = setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => clearInterval(id);
  }, [isRecording, session.started_at]);

  const elapsed = useMemo(() => {
    const startedAt = parseTimestamp(session.started_at);
    if (startedAt === null) {
      return 0;
    }

    const finishedAt = parseTimestamp(session.finished_at);
    const end = finishedAt ?? now;
    return Math.max(0, Math.floor((end - startedAt) / 1000));
  }, [now, session.finished_at, session.started_at]);

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
    void (async () => {
      try {
        if (isRecording) {
          const response = await stopRecording().unwrap();
          dispatch(setSession(response.session));
          return;
        }

        const response = await startRecording({
          dsl: dslText,
        }).unwrap();
        dispatch(setSession(response.session));
      } catch (error) {
        dispatch(addLog({
          timestamp: new Date().toISOString(),
          process_label: 'ui',
          stream: 'stderr',
          message: errorMessageFromUnknown(error),
        } satisfies ProcessLog));
      }
    })();
  };

  const handleCompile = () => {
    void (async () => {
      dispatch(compileStarted());
      try {
        const response = await compileSetup({ dsl: dslText }).unwrap();
        dispatch(compileSucceeded(response.warnings));
      } catch (error) {
        dispatch(compileFailed([errorMessageFromUnknown(error)]));
      }
    })();
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
            onClick={() => dispatch(setActiveTab(tab as StudioTab))}
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
                pauseSupported={false}
                transportBusy={transportBusy}
                elapsed={elapsed}
                armedCount={armedSources.length}
                onFormatChange={(value) => dispatch(setFormat(value))}
                onFpsChange={(value) => dispatch(setFps(value))}
                onQualityChange={(value) => dispatch(setQuality(value))}
                onAudioChange={(value) => dispatch(setAudio(value))}
                onMultiTrackChange={(value) => dispatch(setMultiTrack(value))}
                onToggleRecording={handleToggleRecording}
                onTogglePause={() => {}}
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
            onChange={(value) => dispatch(setDslText(value))}
            onCompile={handleCompile}
            isCompiling={isCompiling}
            warnings={compileWarnings}
            errors={compileErrors}
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
