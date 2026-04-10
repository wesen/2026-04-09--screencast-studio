import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import type { EffectiveConfig } from '@/api/types';
import { effectiveConfigToSetupDraft } from './conversion';
import type {
  SetupDraftAudioSource,
  SetupDraftAudioOutput,
  SetupDraftDocument,
  SetupDraftVideoSource,
} from './types';

export interface SetupDraftState extends SetupDraftDocument {
  hydratedFromSessionId: string;
}

const initialState: SetupDraftState = {
  schema: '',
  sessionId: '',
  destinationTemplates: {},
  audioMixTemplate: '',
  audioOutput: {
    codec: 'pcm_s16le',
    sampleRateHz: 48000,
    channels: 2,
  },
  videoSources: [],
  audioSources: [],
  hydratedFromSessionId: '',
};

const findSourceIndex = (
  sources: SetupDraftVideoSource[],
  sourceId: string
): number => sources.findIndex((source) => source.id === sourceId);

const findAudioSourceIndex = (
  sources: SetupDraftAudioSource[],
  sourceId: string
): number => sources.findIndex((source) => source.id === sourceId);

const setupDraftSlice = createSlice({
  name: 'setupDraft',
  initialState,
  reducers: {
    hydrateFromEffectiveConfig(state, action: PayloadAction<EffectiveConfig>) {
      const next = effectiveConfigToSetupDraft(action.payload);
      state.schema = next.schema;
      state.sessionId = next.sessionId;
      state.destinationTemplates = next.destinationTemplates;
      state.audioMixTemplate = next.audioMixTemplate;
      state.audioOutput = next.audioOutput;
      state.videoSources = next.videoSources;
      state.audioSources = next.audioSources;
      state.hydratedFromSessionId = action.payload.sessionId;
    },
    setSessionId(state, action: PayloadAction<string>) {
      state.sessionId = action.payload;
    },
    replaceDestinationTemplates(state, action: PayloadAction<Record<string, string>>) {
      state.destinationTemplates = action.payload;
    },
    setAudioOutput(state, action: PayloadAction<SetupDraftAudioOutput>) {
      state.audioOutput = action.payload;
    },
    replaceVideoSources(state, action: PayloadAction<SetupDraftVideoSource[]>) {
      state.videoSources = action.payload;
    },
    addVideoSource(state, action: PayloadAction<SetupDraftVideoSource>) {
      state.videoSources.push(action.payload);
    },
    updateVideoSource(state, action: PayloadAction<SetupDraftVideoSource>) {
      const index = findSourceIndex(state.videoSources, action.payload.id);
      if (index === -1) {
        return;
      }
      state.videoSources[index] = action.payload;
    },
    removeVideoSource(state, action: PayloadAction<string>) {
      state.videoSources = state.videoSources.filter((source) => source.id !== action.payload);
    },
    moveVideoSource(
      state,
      action: PayloadAction<{ sourceId: string; direction: 'up' | 'down' }>
    ) {
      const index = findSourceIndex(state.videoSources, action.payload.sourceId);
      if (index === -1) {
        return;
      }

      const targetIndex = action.payload.direction === 'up'
        ? index - 1
        : index + 1;
      if (targetIndex < 0 || targetIndex >= state.videoSources.length) {
        return;
      }

      const [source] = state.videoSources.splice(index, 1);
      state.videoSources.splice(targetIndex, 0, source);
    },
    renameVideoSource(
      state,
      action: PayloadAction<{ sourceId: string; name: string }>
    ) {
      const index = findSourceIndex(state.videoSources, action.payload.sourceId);
      if (index === -1) {
        return;
      }
      state.videoSources[index] = {
        ...state.videoSources[index],
        name: action.payload.name,
      };
    },
    setVideoSourceEnabled(
      state,
      action: PayloadAction<{ sourceId: string; enabled: boolean }>
    ) {
      const index = findSourceIndex(state.videoSources, action.payload.sourceId);
      if (index === -1) {
        return;
      }
      state.videoSources[index] = {
        ...state.videoSources[index],
        enabled: action.payload.enabled,
      };
    },
    replaceAudioSources(state, action: PayloadAction<SetupDraftAudioSource[]>) {
      state.audioSources = action.payload;
    },
    addAudioSource(state, action: PayloadAction<SetupDraftAudioSource>) {
      state.audioSources.push(action.payload);
    },
    updateAudioSource(state, action: PayloadAction<SetupDraftAudioSource>) {
      const index = findAudioSourceIndex(state.audioSources, action.payload.id);
      if (index === -1) {
        return;
      }
      state.audioSources[index] = action.payload;
    },
    clearSetupDraft: () => initialState,
  },
});

export const {
  hydrateFromEffectiveConfig,
  setSessionId,
  replaceDestinationTemplates,
  setAudioOutput,
  replaceVideoSources,
  addVideoSource,
  updateVideoSource,
  removeVideoSource,
  moveVideoSource,
  renameVideoSource,
  setVideoSourceEnabled,
  replaceAudioSources,
  addAudioSource,
  updateAudioSource,
  clearSetupDraft,
} = setupDraftSlice.actions;

export const setupDraftReducer = setupDraftSlice.reducer;

export const selectSetupDraftSessionId = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft.sessionId;
export const selectSetupDraftDocument = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft;
export const selectSetupDraftVideoSources = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft.videoSources;
export const selectSetupDraftAudioSources = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft.audioSources;
export const selectSetupDraftHydratedFromSessionId = (
  state: { setupDraft: SetupDraftState }
) => state.setupDraft.hydratedFromSessionId;
