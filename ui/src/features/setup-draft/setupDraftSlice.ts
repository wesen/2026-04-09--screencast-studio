import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import type { EffectiveConfig } from '@/api/types';
import { effectiveConfigToSetupDraft } from './conversion';
import type {
  SetupDraftAudioSource,
  SetupDraftDocument,
  SetupDraftVideoSource,
} from './types';

export interface SetupDraftState extends SetupDraftDocument {
  hydratedFromSessionId: string;
}

const initialState: SetupDraftState = {
  sessionId: '',
  videoSources: [],
  audioSources: [],
  hydratedFromSessionId: '',
};

const findSourceIndex = (
  sources: SetupDraftVideoSource[],
  sourceId: string
): number => sources.findIndex((source) => source.id === sourceId);

const setupDraftSlice = createSlice({
  name: 'setupDraft',
  initialState,
  reducers: {
    hydrateFromEffectiveConfig(state, action: PayloadAction<EffectiveConfig>) {
      const next = effectiveConfigToSetupDraft(action.payload);
      state.sessionId = next.sessionId;
      state.videoSources = next.videoSources;
      state.audioSources = next.audioSources;
      state.hydratedFromSessionId = action.payload.sessionId;
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
    clearSetupDraft: () => initialState,
  },
});

export const {
  hydrateFromEffectiveConfig,
  replaceVideoSources,
  addVideoSource,
  updateVideoSource,
  removeVideoSource,
  moveVideoSource,
  renameVideoSource,
  setVideoSourceEnabled,
  replaceAudioSources,
  clearSetupDraft,
} = setupDraftSlice.actions;

export const setupDraftReducer = setupDraftSlice.reducer;

export const selectSetupDraftSessionId = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft.sessionId;
export const selectSetupDraftVideoSources = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft.videoSources;
export const selectSetupDraftAudioSources = (state: { setupDraft: SetupDraftState }) =>
  state.setupDraft.audioSources;
export const selectSetupDraftHydratedFromSessionId = (
  state: { setupDraft: SetupDraftState }
) => state.setupDraft.hydratedFromSessionId;
