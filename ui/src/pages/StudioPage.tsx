import { create } from '@bufbuild/protobuf';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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
  CameraDescriptor,
  DisplayDescriptor,
  PreviewDescriptor,
  WindowDescriptor,
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
  selectRawDslDirty,
  selectRawDslText,
  selectStructuredEditingLocked,
  selectStructuredEditingLockReason,
  resetRawDslToApplied,
  setDslText,
  setRawDslText,
} from '@/features/editor/editorSlice';
import { selectActiveTab, setActiveTab, type StudioTab } from '@/features/studio-ui/studioUiSlice';
import {
  addLog,
  selectAudioMeter,
  selectDiskStatus,
  selectSession,
  selectLogs,
  setSession,
  selectWsConnected,
} from '@/features/session/sessionSlice';
import { WsClient } from '@/features/session/wsClient';
import {
  compilePreviewFailed,
  compilePreviewStarted,
  compilePreviewSucceeded,
  normalizeFailed,
  normalizeStarted,
  normalizeSucceeded,
  selectCompiledOutputs,
  selectCompilePreviewErrors,
  selectCompilePreviewWarnings,
  selectIsCompilingPreview,
  selectNormalizedConfig,
  selectNormalizeErrors,
  selectNormalizeWarnings,
} from '@/features/setup/setupSlice';
import {
  applyDestinationRootToTemplates,
  createCameraSourceDraft,
  createDisplaySourceDraft,
  destinationRootFromTemplates,
  effectiveConfigToSetupDraft,
  createRegionSourceDraft,
  createWindowSourceDraft,
  effectiveVideoSourceToDraft,
  presetRectForDisplay,
  renderSetupDraftAsDsl,
  type RegionPreset,
} from '@/features/setup-draft/conversion';
import { isBuilderCompatibleEffectiveConfig } from '@/features/setup-draft/compatibility';
import {
  addAudioSource,
  addVideoSource,
  hydrateFromEffectiveConfig,
  moveVideoSource,
  removeVideoSource,
  replaceVideoSources,
  selectSetupDraftDocument,
  setAudioOutput,
  setSessionId,
  replaceDestinationTemplates,
  setVideoSourceEnabled,
  updateAudioSource,
  updateVideoSource,
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
import type { SetupDraftVideoSource } from '@/features/setup-draft/types';

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

const describeSource = (source: SetupDraftVideoSource): string => {
  switch (source.kind) {
    case 'display':
      return `X11 display ${source.target.displayId}`;
    case 'window':
      return `Window ${source.target.windowId}`;
    case 'region':
      return `${source.target.rect.w}x${source.target.rect.h} at ${source.target.rect.x},${source.target.rect.y}`;
    case 'camera':
      return `Device ${source.target.deviceId}`;
    default:
      return '';
  }
};

const containerToFormat = (container?: string): 'MOV' | 'AVI' | 'MP4' => {
  switch ((container ?? '').trim().toLowerCase()) {
    case 'avi':
      return 'AVI';
    case 'mp4':
      return 'MP4';
    case 'mov':
    default:
      return 'MOV';
  }
};

const formatToContainer = (format: 'MOV' | 'AVI' | 'MP4'): string => format.toLowerCase();

const sampleRateLabel = (sampleRateHz?: number): string => {
  switch (sampleRateHz) {
    case 22050:
      return '22 kHz, 8-bit';
    case 44100:
      return '44 kHz, 16-bit';
    case 48000:
    default:
      return '48 kHz, 16-bit';
  }
};

const labelToSampleRate = (label: string): number => {
  switch (label) {
    case '22 kHz, 8-bit':
      return 22050;
    case '44 kHz, 16-bit':
      return 44100;
    case '48 kHz, 16-bit':
    default:
      return 48000;
  }
};

const toStudioSource = (
  source: SetupDraftVideoSource,
  preview?: PreviewDescriptor
): StudioSource => {
  const kind = source.kind === 'window'
    ? 'Window'
    : source.kind === 'region'
      ? 'Region'
      : source.kind === 'camera'
        ? 'Camera'
        : 'Display';

  return {
    id: source.id,
    sourceId: source.id,
    kind,
    scene: source.name,
    armed: source.enabled,
    label: source.name,
    detail: describeSource(source),
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
  const audioMeter = useAppSelector(selectAudioMeter);
  const diskStatus = useAppSelector(selectDiskStatus);
  const wsConnected = useAppSelector(selectWsConnected);
  const activeTab = useAppSelector(selectActiveTab);
  const dslText = useAppSelector(selectDslText);
  const rawDslText = useAppSelector(selectRawDslText);
  const rawDslDirty = useAppSelector(selectRawDslDirty);
  const setupDraft = useAppSelector(selectSetupDraftDocument);
  const compileWarnings = useAppSelector(selectCompileWarnings);
  const compileErrors = useAppSelector(selectCompileErrors);
  const isCompiling = useAppSelector(selectIsCompiling);
  const structuredEditingLocked = useAppSelector(selectStructuredEditingLocked);
  const structuredEditingLockReason = useAppSelector(selectStructuredEditingLockReason);
  const normalizedConfig = useAppSelector(selectNormalizedConfig);
  const compiledOutputs = useAppSelector(selectCompiledOutputs);
  const compilePreviewWarnings = useAppSelector(selectCompilePreviewWarnings);
  const compilePreviewErrors = useAppSelector(selectCompilePreviewErrors);
  const normalizeWarnings = useAppSelector(selectNormalizeWarnings);
  const normalizeErrors = useAppSelector(selectNormalizeErrors);
  const isCompilingPreview = useAppSelector(selectIsCompilingPreview);
  const previewsById = useAppSelector(selectPreviewsById);
  const ownedPreviewIdBySourceId = useAppSelector(selectOwnedPreviewIdBySourceId);
  const [now, setNow] = useState(() => Date.now());
  const [bootstrapReady, setBootstrapReady] = useState(false);
  const ownedPreviewIdBySourceIdRef = useRef<Record<string, string>>({});
  const pendingPreviewEnsuresRef = useRef<Set<string>>(new Set());
  const pendingPreviewReleasesRef = useRef<Map<string, string>>(new Map());
  const previewGenerationRef = useRef<Record<string, number>>({});
  const bootstrapAppliedRef = useRef(false);
  const { data: healthData } = useGetHealthQuery();
  const { data: discoveryData } = useGetDiscoveryQuery();
  const { data: currentSessionData } = useGetCurrentSessionQuery();
  const [startRecording, startRecordingState] = useStartRecordingMutation();
  const [stopRecording, stopRecordingState] = useStopRecordingMutation();
  const [normalizeSetup] = useNormalizeSetupMutation();
  const [compileSetup] = useCompileSetupMutation();
  const [ensurePreview] = useEnsurePreviewMutation();
  const [releasePreview] = useReleasePreviewMutation();
  const [sourcePickerKind, setSourcePickerKind] = useState<StudioSource['kind'] | null>(null);
  const [previewSyncNonce, setPreviewSyncNonce] = useState(0);

  const visibleVideoSources = useMemo(
    () => (
      structuredEditingLocked && normalizedConfig
        ? normalizedConfig.videoSources.map(effectiveVideoSourceToDraft)
        : setupDraft.videoSources
    ),
    [normalizedConfig, setupDraft.videoSources, structuredEditingLocked]
  );

  const isRecording = session.active;
  const isPaused = false;
  const previewLimit = healthData?.previewLimit || DEFAULT_PREVIEW_LIMIT;
  const transportBusy =
    startRecordingState.isLoading ||
    stopRecordingState.isLoading ||
    session.state === 'starting' ||
    session.state === 'stopping';
  const sources = useMemo(
    () => visibleVideoSources.map((source) => {
      const previewId = ownedPreviewIdBySourceId[source.id];
      const preview = previewId ? previewsById[previewId] : undefined;
      return toStudioSource(source, preview);
    }),
    [ownedPreviewIdBySourceId, previewsById, visibleVideoSources]
  );
  const armedSources = useMemo(
    () => sources.filter((source) => source.armed),
    [sources]
  );
  const desiredPreviewSourceIds = useMemo(
    () => activeTab === 'studio'
      ? sources.filter((source) => source.armed).slice(0, previewLimit).map((source) => source.sourceId)
      : [],
    [activeTab, previewLimit, sources]
  );
  const editorWarnings = useMemo(
    () => [...normalizeWarnings, ...compilePreviewWarnings, ...compileWarnings],
    [compilePreviewWarnings, compileWarnings, normalizeWarnings]
  );
  const editorErrors = useMemo(
    () => [...normalizeErrors, ...compilePreviewErrors, ...compileErrors],
    [compileErrors, compilePreviewErrors, normalizeErrors]
  );

  useEffect(() => {
    if (!healthData || bootstrapAppliedRef.current) {
      return;
    }

    if (healthData.initialDsl) {
      dispatch(setDslText(healthData.initialDsl));
    }
    bootstrapAppliedRef.current = true;
    setBootstrapReady(true);
  }, [dispatch, healthData]);

  useEffect(() => {
    if (currentSessionData?.session) {
      dispatch(setSession(currentSessionData.session));
    }
  }, [currentSessionData, dispatch]);

  useEffect(() => {
    if (!normalizedConfig) {
      return;
    }
    if (structuredEditingLocked) {
      return;
    }
    if (setupDraft.sessionId !== '') {
      return;
    }
    dispatch(hydrateFromEffectiveConfig(normalizedConfig));
  }, [dispatch, normalizedConfig, setupDraft.sessionId, structuredEditingLocked]);

  useEffect(() => {
    if (structuredEditingLocked) {
      setSourcePickerKind(null);
    }
  }, [structuredEditingLocked]);

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
    if (!bootstrapReady) {
      return;
    }
    const timeoutId = window.setTimeout(() => {
      void (async () => {
        dispatch(normalizeStarted());
        dispatch(compilePreviewStarted());
        try {
          const [normalizeResponse, compileResponse] = await Promise.all([
            normalizeSetup({ dsl: dslText }).unwrap(),
            compileSetup({ dsl: dslText }).unwrap(),
          ]);
          if (!normalizeResponse.config) {
            dispatch(normalizeFailed(['normalize response missing config']));
            dispatch(compilePreviewFailed(['normalize response missing config']));
            return;
          }
          dispatch(normalizeSucceeded({
            config: normalizeResponse.config,
            warnings: normalizeResponse.warnings,
          }));
          dispatch(compilePreviewSucceeded({
            outputs: compileResponse.outputs,
            warnings: compileResponse.warnings,
          }));
        } catch (error) {
          dispatch(normalizeFailed([errorMessageFromUnknown(error)]));
          dispatch(compilePreviewFailed([errorMessageFromUnknown(error)]));
        }
      })();
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [bootstrapReady, compileSetup, dispatch, dslText, normalizeSetup]);

  useEffect(() => {
    ownedPreviewIdBySourceIdRef.current = ownedPreviewIdBySourceId;
  }, [ownedPreviewIdBySourceId]);

  const requestPreviewSync = useCallback(() => {
    setPreviewSyncNonce((value) => value + 1);
  }, []);

  const releaseDetachedPreview = useCallback((previewId: string, sourceId: string) => {
    void (async () => {
      try {
        const response = await releasePreview({ previewId }).unwrap();
        if (response.preview) {
          dispatch(upsertPreview(response.preview));
        }
      } catch (error) {
        if (apiErrorCodeFromUnknown(error) !== 'preview_not_found') {
          dispatch(addLog(create(ProcessLogSchema, {
            timestamp: new Date().toISOString(),
            processLabel: 'ui.preview',
            stream: 'stderr',
            message: `failed to release stale preview ${previewId} for ${sourceId}: ${errorMessageFromUnknown(error)}`,
          })));
        }
      } finally {
        requestPreviewSync();
      }
    })();
  }, [dispatch, releasePreview, requestPreviewSync]);

  const releaseOwnedPreviewForSource = useCallback((sourceId: string, previewId: string) => {
    if (pendingPreviewReleasesRef.current.has(sourceId)) {
      return;
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
        requestPreviewSync();
      }
    })();
  }, [dispatch, releasePreview, requestPreviewSync]);

  const restartPreviewForSource = useCallback((sourceId: string) => {
    previewGenerationRef.current[sourceId] = (previewGenerationRef.current[sourceId] ?? 0) + 1;
    const previewId = ownedPreviewIdBySourceIdRef.current[sourceId];
    if (previewId) {
      releaseOwnedPreviewForSource(sourceId, previewId);
      return;
    }
    requestPreviewSync();
  }, [releaseOwnedPreviewForSource, requestPreviewSync]);

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
      const generation = previewGenerationRef.current[sourceId] ?? 0;
      void (async () => {
        try {
          const response = await ensurePreview({
            dsl: dslText,
            sourceId,
          }).unwrap();
          if (response.preview) {
            const currentGeneration = previewGenerationRef.current[sourceId] ?? 0;
            if (currentGeneration !== generation) {
              releaseDetachedPreview(response.preview.id, sourceId);
              return;
            }
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
      if (desired.has(sourceId) || pendingPreviewReleasesRef.current.has(sourceId)) {
        continue;
      }

      releaseOwnedPreviewForSource(sourceId, previewId);
    }
  }, [
    desiredPreviewSourceIds,
    dispatch,
    dslText,
    ensurePreview,
    ownedPreviewIdBySourceId,
    previewSyncNonce,
    releaseDetachedPreview,
    releasePreview,
    releaseOwnedPreviewForSource,
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

  const primaryVideoSource = setupDraft.videoSources[0];
  const primaryAudioSource = setupDraft.audioSources[0];
  const outputSettings = useMemo(() => ({
    format: containerToFormat(primaryVideoSource?.output.container),
    fps: `${primaryVideoSource?.capture.fps ?? 24} fps`,
    quality: primaryVideoSource?.output.quality ?? 75,
    audio: sampleRateLabel(setupDraft.audioOutput.sampleRateHz),
    multiTrack: true,
  }), [primaryVideoSource, setupDraft.audioOutput.sampleRateHz]);

  const micSettings = useMemo(() => ({
    micInput: primaryAudioSource?.deviceId ?? '',
    gain: Math.round((primaryAudioSource?.gain ?? 1) * 100),
  }), [primaryAudioSource]);

  const destinationRoot = useMemo(
    () => destinationRootFromTemplates(setupDraft.destinationTemplates),
    [setupDraft.destinationTemplates]
  );
  const destinationRootEditable = !structuredEditingLocked && destinationRoot !== null;
  const destinationRootReason = structuredEditingLocked
    ? structuredEditingLockReason
    : destinationRoot === null
      ? 'Advanced destination templates are active. Edit Raw DSL to change output paths.'
      : undefined;
  const micOptions = useMemo(
    () => {
      const options = (discoveryData?.audio ?? [])
        .filter((input) => !input.id.endsWith('.monitor'))
        .map((input) => ({
          value: input.id,
          label: input.name || input.id,
        }));

      if (primaryAudioSource?.deviceId) {
        const hasCurrent = options.some((option) => option.value === primaryAudioSource.deviceId);
        if (!hasCurrent) {
          options.unshift({
            value: primaryAudioSource.deviceId,
            label: primaryAudioSource.name
              ? `${primaryAudioSource.name} (${primaryAudioSource.deviceId})`
              : primaryAudioSource.deviceId,
          });
        }
      }

      return options;
    },
    [discoveryData?.audio, primaryAudioSource?.deviceId, primaryAudioSource?.name]
  );
  const diskPercent = diskStatus?.available ? diskStatus.usedPercent : undefined;

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

  const handleApplyDsl = () => {
    void (async () => {
      dispatch(compileStarted());
      dispatch(normalizeStarted());
      try {
        const response = await normalizeSetup({ dsl: rawDslText }).unwrap();
        if (!response.config) {
          dispatch(normalizeFailed(['normalize response missing config']));
          dispatch(compileFailed(['normalize response missing config']));
          return;
        }

        const roundTrippedDsl = renderSetupDraftAsDsl(
          effectiveConfigToSetupDraft(response.config)
        );
        const roundTripResponse = await normalizeSetup({ dsl: roundTrippedDsl }).unwrap();
        if (!roundTripResponse.config) {
          dispatch(compileFailed(['builder compatibility check missing config']));
          return;
        }

        const compatible = isBuilderCompatibleEffectiveConfig(
          response.config,
          roundTripResponse.config,
        );
        const lockReason = compatible
          ? ''
          : 'Advanced DSL is active. Structured editing is unavailable because this setup uses shapes the builder does not support yet.';

        dispatch(normalizeSucceeded({
          config: response.config,
          warnings: response.warnings,
        }));

        if (compatible) {
          dispatch(hydrateFromEffectiveConfig(response.config));
        }

        dispatch(compileSucceeded({
          dslText: rawDslText,
          warnings: response.warnings,
          structuredEditingLocked: !compatible,
          structuredEditingLockReason: lockReason,
        }));
      } catch (error) {
        dispatch(normalizeFailed([errorMessageFromUnknown(error)]));
        dispatch(compileFailed([errorMessageFromUnknown(error)]));
      }
    })();
  };

  const syncStructuredDraft = useCallback((nextDraft: typeof setupDraft) => {
    dispatch(setDslText(renderSetupDraftAsDsl(nextDraft)));
  }, [dispatch]);

  const applyAddedSource = (source: ReturnType<typeof createDisplaySourceDraft>) => {
    if (structuredEditingLocked) {
      return;
    }
    const nextDraft = {
      ...setupDraft,
      videoSources: [...setupDraft.videoSources, source],
    };
    dispatch(addVideoSource(source));
    syncStructuredDraft(nextDraft);
    setSourcePickerKind(null);
  };

  const applyUpdatedSource = (
    updatedSource: SetupDraftVideoSource,
    options?: { restartPreview?: boolean }
  ) => {
    if (structuredEditingLocked) {
      return;
    }
    const nextDraft = {
      ...setupDraft,
      videoSources: setupDraft.videoSources.map((source) => (
        source.id === updatedSource.id ? updatedSource : source
      )),
    };
    dispatch(updateVideoSource(updatedSource));
    syncStructuredDraft(nextDraft);
    if (options?.restartPreview) {
      restartPreviewForSource(updatedSource.id);
    }
  };

  const handleRenameSource = (sourceId: string, name: string) => {
    const source = setupDraft.videoSources.find((item) => item.id === sourceId);
    if (!source) {
      return;
    }
    applyUpdatedSource({
      ...source,
      name,
    });
  };

  const handleToggleSourceEnabled = (sourceId: string) => {
    const source = setupDraft.videoSources.find((item) => item.id === sourceId);
    if (!source) {
      return;
    }
    const nextDraft = {
      ...setupDraft,
      videoSources: setupDraft.videoSources.map((item) => (
        item.id === sourceId
          ? { ...item, enabled: !item.enabled }
          : item
      )),
    };
    dispatch(setVideoSourceEnabled({ sourceId, enabled: !source.enabled }));
    syncStructuredDraft(nextDraft);
  };

  const handleRemoveSource = (sourceId: string) => {
    const nextDraft = {
      ...setupDraft,
      videoSources: setupDraft.videoSources.filter((source) => source.id !== sourceId),
    };
    dispatch(removeVideoSource(sourceId));
    syncStructuredDraft(nextDraft);
  };

  const handleMoveSource = (sourceId: string, direction: 'up' | 'down') => {
    const index = setupDraft.videoSources.findIndex((source) => source.id === sourceId);
    if (index === -1) {
      return;
    }
    const targetIndex = direction === 'up' ? index - 1 : index + 1;
    if (targetIndex < 0 || targetIndex >= setupDraft.videoSources.length) {
      return;
    }
    const nextSources = [...setupDraft.videoSources];
    const [source] = nextSources.splice(index, 1);
    nextSources.splice(targetIndex, 0, source);
    const nextDraft = {
      ...setupDraft,
      videoSources: nextSources,
    };
    dispatch(moveVideoSource({ sourceId, direction }));
    syncStructuredDraft(nextDraft);
  };

  const handleRecordingNameChange = (value: string) => {
    if (structuredEditingLocked) {
      return;
    }
    const nextDraft = {
      ...setupDraft,
      sessionId: value.trim(),
    };
    dispatch(setSessionId(nextDraft.sessionId));
    syncStructuredDraft(nextDraft);
  };

  const handleDestinationRootChange = (value: string) => {
    if (structuredEditingLocked || destinationRoot === null) {
      return;
    }
    const nextTemplates = applyDestinationRootToTemplates(value, setupDraft.destinationTemplates);
    const nextDraft = {
      ...setupDraft,
      destinationTemplates: nextTemplates,
    };
    dispatch(replaceDestinationTemplates(nextTemplates));
    syncStructuredDraft(nextDraft);
  };

  const handleFormatChange = (value: 'MOV' | 'AVI' | 'MP4') => {
    if (structuredEditingLocked) {
      return;
    }
    const nextVideoSources = setupDraft.videoSources.map((source) => ({
      ...source,
      output: {
        ...source.output,
        container: formatToContainer(value),
      },
    }));
    const nextDraft = {
      ...setupDraft,
      videoSources: nextVideoSources,
    };
    dispatch(replaceVideoSources(nextVideoSources));
    syncStructuredDraft(nextDraft);
  };

  const handleFpsChange = (value: string) => {
    if (structuredEditingLocked) {
      return;
    }
    const fps = Number.parseInt(value, 10) || 24;
    const nextVideoSources = setupDraft.videoSources.map((source) => ({
      ...source,
      capture: {
        ...source.capture,
        fps,
      },
    }));
    const nextDraft = {
      ...setupDraft,
      videoSources: nextVideoSources,
    };
    dispatch(replaceVideoSources(nextVideoSources));
    syncStructuredDraft(nextDraft);
  };

  const handleQualityChange = (value: number) => {
    if (structuredEditingLocked) {
      return;
    }
    const nextVideoSources = setupDraft.videoSources.map((source) => ({
      ...source,
      output: {
        ...source.output,
        quality: value,
      },
    }));
    const nextDraft = {
      ...setupDraft,
      videoSources: nextVideoSources,
    };
    dispatch(replaceVideoSources(nextVideoSources));
    syncStructuredDraft(nextDraft);
  };

  const handleAudioOutputChange = (value: string) => {
    if (structuredEditingLocked) {
      return;
    }
    const nextAudioOutput = {
      ...setupDraft.audioOutput,
      sampleRateHz: labelToSampleRate(value),
    };
    const nextDraft = {
      ...setupDraft,
      audioOutput: nextAudioOutput,
    };
    dispatch(setAudioOutput(nextAudioOutput));
    syncStructuredDraft(nextDraft);
  };

  const handleMicInputChange = (deviceId: string) => {
    if (structuredEditingLocked) {
      return;
    }
    const descriptor = (discoveryData?.audio ?? []).find((input) => input.id === deviceId);
    const nextSource = primaryAudioSource
      ? {
        ...primaryAudioSource,
        deviceId,
        name: descriptor?.name || primaryAudioSource.name,
      }
      : {
        id: 'mic-1',
        name: descriptor?.name || 'Microphone',
        deviceId,
        enabled: true,
        gain: 1,
        noiseGate: false,
        denoise: false,
      };
    const nextAudioSources = primaryAudioSource
      ? [nextSource, ...setupDraft.audioSources.slice(1)]
      : [...setupDraft.audioSources, nextSource];
    const nextDraft = {
      ...setupDraft,
      audioSources: nextAudioSources,
    };
    if (primaryAudioSource) {
      dispatch(updateAudioSource(nextSource));
    } else {
      dispatch(addAudioSource(nextSource));
    }
    syncStructuredDraft(nextDraft);
  };

  const handleGainChange = (value: number) => {
    if (structuredEditingLocked || !primaryAudioSource) {
      return;
    }
    const nextSource = {
      ...primaryAudioSource,
      gain: Math.max(0, value) / 100,
    };
    const nextDraft = {
      ...setupDraft,
      audioSources: [nextSource, ...setupDraft.audioSources.slice(1)],
    };
    dispatch(updateAudioSource(nextSource));
    syncStructuredDraft(nextDraft);
  };

  const displays = discoveryData?.displays ?? [];
  const windows = discoveryData?.windows ?? [];
  const cameras = discoveryData?.cameras ?? [];

  const renderTargetEditor = (studioSource: StudioSource): React.ReactNode => {
    const source = setupDraft.videoSources.find((item) => item.id === studioSource.id);
    if (!source) {
      return null;
    }

    const selectClassName = 'studio-source-card__editor';
    switch (source.kind) {
      case 'window':
        return (
          <div className={selectClassName}>
            <label>
              Window
              <select
                value={source.target.windowId}
                onChange={(event) => applyUpdatedSource({
                  ...source,
                  name: windows.find((window) => window.id === event.target.value)?.title || source.name,
                  target: {
                    windowId: event.target.value,
                  },
                }, { restartPreview: true })}
              >
                {windows.map((window: WindowDescriptor) => (
                  <option key={window.id} value={window.id}>
                    {window.title || window.id}
                  </option>
                ))}
              </select>
            </label>
          </div>
        );
      case 'camera':
        return (
          <div className={selectClassName}>
            <label>
              Camera Device
              <select
                value={source.target.deviceId}
                onChange={(event) => applyUpdatedSource({
                  ...source,
                  name: cameras.find((camera) => camera.device === event.target.value)?.label || source.name,
                  target: {
                    deviceId: event.target.value,
                  },
                }, { restartPreview: true })}
              >
                {cameras.map((camera: CameraDescriptor) => (
                  <option key={camera.id} value={camera.device}>
                    {camera.label || camera.device}
                  </option>
                ))}
              </select>
            </label>
          </div>
        );
      case 'region':
        return (
          <div className={selectClassName}>
            <div className="studio-source-card__editor-grid">
              <label>
                X
                <input
                  type="number"
                  value={source.target.rect.x}
                  onChange={(event) => applyUpdatedSource({
                    ...source,
                    target: {
                      ...source.target,
                      rect: {
                        ...source.target.rect,
                        x: Number(event.target.value) || 0,
                      },
                    },
                  }, { restartPreview: true })}
                />
              </label>
              <label>
                Y
                <input
                  type="number"
                  value={source.target.rect.y}
                  onChange={(event) => applyUpdatedSource({
                    ...source,
                    target: {
                      ...source.target,
                      rect: {
                        ...source.target.rect,
                        y: Number(event.target.value) || 0,
                      },
                    },
                  }, { restartPreview: true })}
                />
              </label>
              <label>
                Width
                <input
                  type="number"
                  value={source.target.rect.w}
                  onChange={(event) => applyUpdatedSource({
                    ...source,
                    target: {
                      ...source.target,
                      rect: {
                        ...source.target.rect,
                        w: Number(event.target.value) || 0,
                      },
                    },
                  }, { restartPreview: true })}
                />
              </label>
              <label>
                Height
                <input
                  type="number"
                  value={source.target.rect.h}
                  onChange={(event) => applyUpdatedSource({
                    ...source,
                    target: {
                      ...source.target,
                      rect: {
                        ...source.target.rect,
                        h: Number(event.target.value) || 0,
                      },
                    },
                  }, { restartPreview: true })}
                />
              </label>
            </div>
            <div className="studio-source-card__editor-note">
              Presets from discovered displays:
            </div>
            <div className="studio-source-card__editor-actions">
              {displays.map((display: DisplayDescriptor) => (
                <Btn
                  key={display.id}
                  onClick={() => applyUpdatedSource({
                    ...source,
                    target: {
                      ...source.target,
                      rect: presetRectForDisplay(display, 'full'),
                    },
                  }, { restartPreview: true })}
                  style={{ fontSize: '9px', padding: '2px 6px' }}
                >
                  {display.name}
                </Btn>
              ))}
            </div>
          </div>
        );
      case 'display':
      default:
        return (
          <div className={selectClassName}>
            <div className="studio-source-card__editor-note">
              Full-display sources still use the runtime&apos;s root X11 target (`:0.0`).
              Per-monitor display selection needs a backend target-model change, so this source type is currently name/edit/reorder only.
            </div>
          </div>
        );
    }
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
            {structuredEditingLocked && (
              <div
                style={{
                  margin: '10px',
                  padding: '8px 10px',
                  border: '1px solid var(--studio-amber)',
                  background: 'rgba(184, 152, 64, 0.12)',
                  color: 'var(--studio-amber)',
                  fontSize: 10,
                }}
              >
                {structuredEditingLockReason}
              </div>
            )}
            <SourceGrid
              sources={sources}
              isRecording={isRecording}
              editable={!structuredEditingLocked}
              renderEditor={structuredEditingLocked ? undefined : renderTargetEditor}
              onRemove={structuredEditingLocked ? undefined : handleRemoveSource}
              onToggleArmed={structuredEditingLocked ? undefined : handleToggleSourceEnabled}
              onChangeScene={structuredEditingLocked ? undefined : handleRenameSource}
              onMoveUp={structuredEditingLocked ? undefined : (sourceId) => handleMoveSource(sourceId, 'up')}
              onMoveDown={structuredEditingLocked ? undefined : (sourceId) => handleMoveSource(sourceId, 'down')}
              onAdd={structuredEditingLocked ? undefined : (kind) => setSourcePickerKind(kind)}
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
                recordingName={setupDraft.sessionId}
                destinationRoot={destinationRoot ?? ''}
                destinationRootEditable={destinationRootEditable}
                destinationRootReason={destinationRootReason}
                outputs={compiledOutputs}
                outputPreviewBusy={isCompilingPreview}
                outputPreviewErrors={compilePreviewErrors}
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
                onRecordingNameChange={handleRecordingNameChange}
                onDestinationRootChange={handleDestinationRootChange}
                onFormatChange={handleFormatChange}
                onFpsChange={handleFpsChange}
                onQualityChange={handleQualityChange}
                onAudioChange={handleAudioOutputChange}
                onMultiTrackChange={() => {}}
                onToggleRecording={handleToggleRecording}
                onTogglePause={() => {}}
              />

              <div className="studio-panel-stack">
                <MicPanel
                  leftLevel={audioMeter?.available ? audioMeter.leftLevel : undefined}
                  rightLevel={audioMeter?.available ? audioMeter.rightLevel : undefined}
                  micInput={micSettings.micInput}
                  micOptions={micOptions}
                  gain={micSettings.gain}
                  isRecording={isRecording}
                  onMicInputChange={handleMicInputChange}
                  onGainChange={handleGainChange}
                />

                <StatusPanel
                  diskPercent={diskPercent}
                  destinationRoot={destinationRoot ?? undefined}
                  outputCount={compiledOutputs.length}
                  diskFreeGiB={diskStatus?.available ? diskStatus.freeGib : undefined}
                  diskTotalGiB={diskStatus?.available ? diskStatus.totalGib : undefined}
                  diskReason={!diskStatus?.available ? diskStatus?.reason : undefined}
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
            value={rawDslText}
            onChange={(value) => dispatch(setRawDslText(value))}
            onApply={handleApplyDsl}
            onReset={() => dispatch(resetRawDslToApplied())}
            isApplying={isCompiling}
            hasChanges={rawDslDirty}
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
