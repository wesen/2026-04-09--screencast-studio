import { create } from '@bufbuild/protobuf';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useGetDiscoveryQuery, useGetHealthQuery } from '@/api/discoveryApi';
import {
  useEnsurePreviewMutation,
  useReleasePreviewMutation,
} from '@/api/previewsApi';
import {
  useGetCurrentSessionQuery,
  useStartRecordingMutation,
  useStopRecordingMutation,
} from '@/api/recordingApi';
import { useCompileSetupMutation, useNormalizeSetupMutation } from '@/api/setupApi';
import type {
  ApiErrorResponse,
  EffectiveVideoSource,
  PreviewDescriptor,
} from '@/api/types';
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
import {
  createCameraSourceDraft,
  createDisplaySourceDraft,
  createRegionSourceDraft,
  createWindowSourceDraft,
  renderSetupDraftAsDsl,
  type RegionPreset,
} from '@/features/setup-draft/conversion';
import {
  addVideoSource,
  hydrateFromEffectiveConfig,
  selectSetupDraftDocument,
} from '@/features/setup-draft/setupDraftSlice';
import {
  clearOwnedPreview,
  selectOwnedPreviewIdBySourceId,
  selectPreviewsById,
  trackOwnedPreview,
  upsertPreview,
} from '@/features/previews/previewSlice';
import { MenuBar, SourceGrid, OutputPanel, MicPanel, SourcePicker, StatusPanel } from '@/components/studio';
import { LogPanel } from '@/components/log-panel';
import { DSLEditor } from '@/components/dsl-editor';
import { Btn } from '@/components/primitives/Btn';
import { ProcessLogSchema } from '@/gen/proto/screencast/studio/v1/web_pb';

interface StudioPageProps {
  className?: string;
}

const DEFAULT_PREVIEW_LIMIT = 4;

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

const apiErrorCodeFromUnknown = (error: unknown): string | undefined => {
  if (typeof error === 'object' && error !== null && 'data' in error) {
    const data = (error as { data?: unknown }).data;
    if (
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as ApiErrorResponse).error?.code === 'string'
    ) {
      return (data as ApiErrorResponse).error.code;
    }
  }
  return undefined;
};

const previewStreamUrl = (previewId?: string): string | undefined => (
  previewId ? `/api/previews/${previewId}/mjpeg` : undefined
);

const toStudioSource = (
  source: EffectiveVideoSource,
  preview?: PreviewDescriptor
): StudioSource => {
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
    previewId: preview?.id,
    previewState: preview?.state,
    previewReason: preview?.reason,
    previewUrl: preview && preview.state !== 'failed' && preview.state !== 'finished'
      ? previewStreamUrl(preview.id)
      : undefined,
  };
};

export const StudioPage: React.FC<StudioPageProps> = ({ className }) => {
  const dispatch = useAppDispatch();
  const session = useAppSelector(selectSession);
  const logs = useAppSelector(selectLogs);
  const wsConnected = useAppSelector(selectWsConnected);
  const activeTab = useAppSelector(selectActiveTab);
  const dslText = useAppSelector(selectDslText);
  const setupDraft = useAppSelector(selectSetupDraftDocument);
  const compileWarnings = useAppSelector(selectCompileWarnings);
  const compileErrors = useAppSelector(selectCompileErrors);
  const isCompiling = useAppSelector(selectIsCompiling);
  const normalizedConfig = useAppSelector(selectNormalizedConfig);
  const normalizeWarnings = useAppSelector(selectNormalizeWarnings);
  const normalizeErrors = useAppSelector(selectNormalizeErrors);
  const previewsById = useAppSelector(selectPreviewsById);
  const ownedPreviewIdBySourceId = useAppSelector(selectOwnedPreviewIdBySourceId);
  const [now, setNow] = useState(() => Date.now());
  const ownedPreviewIdBySourceIdRef = useRef<Record<string, string>>({});
  const pendingPreviewEnsuresRef = useRef<Set<string>>(new Set());
  const pendingPreviewReleasesRef = useRef<Map<string, string>>(new Map());
  const { data: healthData } = useGetHealthQuery();
  const { data: discoveryData } = useGetDiscoveryQuery();
  const { data: currentSessionData } = useGetCurrentSessionQuery();
  const [startRecording, startRecordingState] = useStartRecordingMutation();
  const [stopRecording, stopRecordingState] = useStopRecordingMutation();
  const [compileSetup] = useCompileSetupMutation();
  const [normalizeSetup] = useNormalizeSetupMutation();
  const [ensurePreview] = useEnsurePreviewMutation();
  const [releasePreview] = useReleasePreviewMutation();
  const [sourcePickerKind, setSourcePickerKind] = useState<StudioSource['kind'] | null>(null);

  const isRecording = session.active;
  const isPaused = false;
  const previewLimit = healthData?.previewLimit || DEFAULT_PREVIEW_LIMIT;
  const transportBusy =
    startRecordingState.isLoading ||
    stopRecordingState.isLoading ||
    session.state === 'starting' ||
    session.state === 'stopping';
  const sources = useMemo(
    () => normalizedConfig?.videoSources.map((source) => {
      const previewId = ownedPreviewIdBySourceId[source.id];
      const preview = previewId ? previewsById[previewId] : undefined;
      return toStudioSource(source, preview);
    }) ?? [],
    [normalizedConfig, ownedPreviewIdBySourceId, previewsById]
  );
  const armedSources = useMemo(
    () => sources.filter((source) => source.armed),
    [sources]
  );
  const desiredPreviewSourceIds = useMemo(
    () => activeTab === 'studio'
      ? sources.slice(0, previewLimit).map((source) => source.sourceId)
      : [],
    [activeTab, previewLimit, sources]
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
    if (!normalizedConfig) {
      return;
    }
    dispatch(hydrateFromEffectiveConfig(normalizedConfig));
  }, [dispatch, normalizedConfig]);

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

  useEffect(() => {
    ownedPreviewIdBySourceIdRef.current = ownedPreviewIdBySourceId;
  }, [ownedPreviewIdBySourceId]);

  useEffect(() => {
    const desired = new Set(desiredPreviewSourceIds);

    for (const sourceId of desired) {
      if (
        ownedPreviewIdBySourceId[sourceId] ||
        pendingPreviewEnsuresRef.current.has(sourceId)
      ) {
        continue;
      }

      pendingPreviewEnsuresRef.current.add(sourceId);
      void (async () => {
        try {
          const response = await ensurePreview({
            dsl: dslText,
            sourceId,
          }).unwrap();
          if (response.preview) {
            dispatch(upsertPreview(response.preview));
            dispatch(trackOwnedPreview({
              sourceId,
              previewId: response.preview.id,
            }));
          }
        } catch (error) {
          dispatch(addLog(create(ProcessLogSchema, {
            timestamp: new Date().toISOString(),
            processLabel: 'ui.preview',
            stream: 'stderr',
            message: `failed to ensure preview for ${sourceId}: ${errorMessageFromUnknown(error)}`,
          })));
        } finally {
          pendingPreviewEnsuresRef.current.delete(sourceId);
        }
      })();
    }

    for (const [sourceId, previewId] of Object.entries(ownedPreviewIdBySourceId)) {
      if (
        desired.has(sourceId) ||
        pendingPreviewReleasesRef.current.has(sourceId)
      ) {
        continue;
      }

      pendingPreviewReleasesRef.current.set(sourceId, previewId);
      void (async () => {
        try {
          const response = await releasePreview({ previewId }).unwrap();
          if (response.preview) {
            dispatch(upsertPreview(response.preview));
          }
          dispatch(clearOwnedPreview({ sourceId }));
        } catch (error) {
          if (apiErrorCodeFromUnknown(error) === 'preview_not_found') {
            dispatch(clearOwnedPreview({ sourceId }));
            return;
          }

          dispatch(addLog(create(ProcessLogSchema, {
            timestamp: new Date().toISOString(),
            processLabel: 'ui.preview',
            stream: 'stderr',
            message: `failed to release preview ${previewId}: ${errorMessageFromUnknown(error)}`,
          })));
        } finally {
          pendingPreviewReleasesRef.current.delete(sourceId);
        }
      })();
    }
  }, [
    desiredPreviewSourceIds,
    dispatch,
    dslText,
    ensurePreview,
    ownedPreviewIdBySourceId,
    releasePreview,
  ]);

  useEffect(() => {
    return () => {
      for (const previewId of Object.values(ownedPreviewIdBySourceIdRef.current)) {
        void releasePreview({ previewId });
      }
    };
  }, [releasePreview]);

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

  const applyAddedSource = (
    source: ReturnType<typeof createDisplaySourceDraft>
  ) => {
    const nextDraft = {
      ...setupDraft,
      videoSources: [...setupDraft.videoSources, source],
    };
    dispatch(addVideoSource(source));
    dispatch(setDslText(renderSetupDraftAsDsl(nextDraft)));
    setSourcePickerKind(null);
  };

  const displays = discoveryData?.displays ?? [];
  const windows = discoveryData?.windows ?? [];
  const cameras = discoveryData?.cameras ?? [];

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
              onAdd={(kind) => setSourcePickerKind(kind)}
            />

            {sourcePickerKind && (
              <SourcePicker
                kind={sourcePickerKind}
                displays={displays}
                windows={windows}
                cameras={cameras}
                onClose={() => setSourcePickerKind(null)}
                onPickDisplay={(display) => {
                  applyAddedSource(createDisplaySourceDraft(display, setupDraft));
                }}
                onPickWindow={(window) => {
                  applyAddedSource(createWindowSourceDraft(window, setupDraft));
                }}
                onPickCamera={(camera) => {
                  applyAddedSource(createCameraSourceDraft(camera, setupDraft));
                }}
                onPickRegion={(display, preset: RegionPreset) => {
                  applyAddedSource(createRegionSourceDraft(display, preset, setupDraft));
                }}
              />
            )}

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
                  micInput={micSettings.micInput}
                  gain={micSettings.gain}
                  isRecording={isRecording}
                  onMicInputChange={(value) => dispatch(setMicInput(value))}
                  onGainChange={(value) => dispatch(setGain(value))}
                />

                <StatusPanel
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
