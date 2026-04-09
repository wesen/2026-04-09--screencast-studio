import { create } from '@bufbuild/protobuf';
import React, { useEffect, useMemo, useState } from 'react';
import {
  useGetCurrentSessionQuery,
  useStartRecordingMutation,
  useStopRecordingMutation,
} from '@/api/recordingApi';
import { useCompileSetupMutation, useNormalizeSetupMutation } from '@/api/setupApi';
import type { ApiErrorResponse, EffectiveVideoSource } from '@/api/types';
import { useAppDispatch, useAppSelector } from '@/app/hooks';
import type { StudioSource } from '@/components/source-card';
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
  setAudio,
  setFormat,
  setFps,
  setGain,
  setMicInput,
  setMultiTrack,
  setQuality,
} from '@/features/studio-draft/studioDraftSlice';
import { selectActiveTab, setActiveTab, type StudioTab } from '@/features/studio-ui/studioUiSlice';
import {
  addLog,
  selectSession,
  selectLogs,
  setSession,
  selectWsConnected,
} from '@/features/session/sessionSlice';
import { WsClient } from '@/features/session/wsClient';
import {
  normalizeFailed,
  normalizeStarted,
  normalizeSucceeded,
  selectNormalizedConfig,
  selectNormalizeErrors,
  selectNormalizeWarnings,
} from '@/features/setup/setupSlice';
import { MenuBar, SourceGrid, OutputPanel, MicPanel, StatusPanel } from '@/components/studio';
import { LogPanel } from '@/components/log-panel';
import { DSLEditor } from '@/components/dsl-editor';
import { Btn } from '@/components/primitives/Btn';
import { ProcessLogSchema } from '@/gen/proto/screencast/studio/v1/web_pb';

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

const toStudioSource = (source: EffectiveVideoSource): StudioSource => {
  const normalizedType = source.type.trim().toLowerCase();
  const kind = normalizedType === 'window'
    ? 'Window'
    : normalizedType === 'region'
      ? 'Region'
      : normalizedType === 'camera'
        ? 'Camera'
        : 'Display';

  return {
    id: source.id,
    sourceId: source.id,
    kind,
    scene: source.name,
    armed: source.enabled,
    solo: false,
    label: source.name,
  };
};

export const StudioPage: React.FC<StudioPageProps> = ({ className }) => {
  const dispatch = useAppDispatch();
  const session = useAppSelector(selectSession);
  const logs = useAppSelector(selectLogs);
  const wsConnected = useAppSelector(selectWsConnected);
  const activeTab = useAppSelector(selectActiveTab);
  const dslText = useAppSelector(selectDslText);
  const compileWarnings = useAppSelector(selectCompileWarnings);
  const compileErrors = useAppSelector(selectCompileErrors);
  const isCompiling = useAppSelector(selectIsCompiling);
  const normalizedConfig = useAppSelector(selectNormalizedConfig);
  const normalizeWarnings = useAppSelector(selectNormalizeWarnings);
  const normalizeErrors = useAppSelector(selectNormalizeErrors);
  const [now, setNow] = useState(() => Date.now());
  const { data: currentSessionData } = useGetCurrentSessionQuery();
  const [startRecording, startRecordingState] = useStartRecordingMutation();
  const [stopRecording, stopRecordingState] = useStopRecordingMutation();
  const [compileSetup] = useCompileSetupMutation();
  const [normalizeSetup] = useNormalizeSetupMutation();

  const isRecording = session.active;
  const isPaused = false;
  const diskPercent = 8;
  const transportBusy =
    startRecordingState.isLoading ||
    stopRecordingState.isLoading ||
    session.state === 'starting' ||
    session.state === 'stopping';
  const sources = useMemo(
    () => normalizedConfig?.videoSources.map(toStudioSource) ?? [],
    [normalizedConfig]
  );
  const armedSources = useMemo(
    () => sources.filter((source) => source.armed),
    [sources]
  );
  const editorWarnings = useMemo(
    () => [...normalizeWarnings, ...compileWarnings],
    [compileWarnings, normalizeWarnings]
  );
  const editorErrors = useMemo(
    () => [...normalizeErrors, ...compileErrors],
    [compileErrors, normalizeErrors]
  );

  useEffect(() => {
    if (currentSessionData?.session) {
      dispatch(setSession(currentSessionData.session));
    }
  }, [currentSessionData, dispatch]);

  useEffect(() => {
    const wsClient = new WsClient(dispatch);
    wsClient.connect();

    return () => {
      wsClient.disconnect();
    };
  }, [dispatch]);

  useEffect(() => {
    const startedAt = parseTimestamp(session.startedAt);
    if (!isRecording || startedAt === null) {
      return;
    }

    const id = setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => clearInterval(id);
  }, [isRecording, session.startedAt]);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void (async () => {
        dispatch(normalizeStarted());
        try {
          const response = await normalizeSetup({ dsl: dslText }).unwrap();
          if (!response.config) {
            dispatch(normalizeFailed(['normalize response missing config']));
            return;
          }
          dispatch(normalizeSucceeded({
            config: response.config,
            warnings: response.warnings,
          }));
        } catch (error) {
          dispatch(normalizeFailed([errorMessageFromUnknown(error)]));
        }
      })();
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [dispatch, dslText, normalizeSetup]);

  const elapsed = useMemo(() => {
    const startedAt = parseTimestamp(session.startedAt);
    if (startedAt === null) {
      return 0;
    }

    const finishedAt = parseTimestamp(session.finishedAt);
    const end = finishedAt ?? now;
    return Math.max(0, Math.floor((end - startedAt) / 1000));
  }, [now, session.finishedAt, session.startedAt]);

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
          if (response.session) {
            dispatch(setSession(response.session));
          }
          return;
        }

        const response = await startRecording({
          dsl: dslText,
        }).unwrap();
        if (response.session) {
          dispatch(setSession(response.session));
        }
      } catch (error) {
        dispatch(addLog(create(ProcessLogSchema, {
          timestamp: new Date().toISOString(),
          processLabel: 'ui',
          stream: 'stderr',
          message: errorMessageFromUnknown(error),
        })));
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
              editable={false}
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
            warnings={editorWarnings}
            errors={editorErrors}
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
