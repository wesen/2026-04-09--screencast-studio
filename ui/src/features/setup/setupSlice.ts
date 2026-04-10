import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { EffectiveConfig, PlannedOutput } from '@/api/types';

export interface SetupState {
  normalizedConfig: EffectiveConfig | null;
  compiledOutputs: PlannedOutput[];
  compileWarnings: string[];
  compileErrors: string[];
  normalizeWarnings: string[];
  normalizeErrors: string[];
  isNormalizing: boolean;
  isCompilingPreview: boolean;
}

const initialState: SetupState = {
  normalizedConfig: null,
  compiledOutputs: [],
  compileWarnings: [],
  compileErrors: [],
  normalizeWarnings: [],
  normalizeErrors: [],
  isNormalizing: false,
  isCompilingPreview: false,
};

const setupSlice = createSlice({
  name: 'setup',
  initialState,
  reducers: {
    normalizeStarted(state) {
      state.isNormalizing = true;
      state.normalizeErrors = [];
    },
    normalizeSucceeded(
      state,
      action: PayloadAction<{ config: EffectiveConfig; warnings: string[] }>
    ) {
      state.isNormalizing = false;
      state.normalizedConfig = action.payload.config;
      state.normalizeWarnings = action.payload.warnings;
      state.normalizeErrors = [];
    },
    normalizeFailed(state, action: PayloadAction<string[]>) {
      state.isNormalizing = false;
      state.normalizeErrors = action.payload;
      state.normalizedConfig = null;
    },
    compilePreviewStarted(state) {
      state.isCompilingPreview = true;
      state.compileErrors = [];
    },
    compilePreviewSucceeded(
      state,
      action: PayloadAction<{ outputs: PlannedOutput[]; warnings: string[] }>
    ) {
      state.isCompilingPreview = false;
      state.compiledOutputs = action.payload.outputs;
      state.compileWarnings = action.payload.warnings;
      state.compileErrors = [];
    },
    compilePreviewFailed(state, action: PayloadAction<string[]>) {
      state.isCompilingPreview = false;
      state.compiledOutputs = [];
      state.compileErrors = action.payload;
    },
  },
});

export const {
  normalizeStarted,
  normalizeSucceeded,
  normalizeFailed,
  compilePreviewStarted,
  compilePreviewSucceeded,
  compilePreviewFailed,
} = setupSlice.actions;
export const setupReducer = setupSlice.reducer;

export const selectNormalizedConfig = (state: { setup: SetupState }) =>
  state.setup.normalizedConfig;
export const selectNormalizeWarnings = (state: { setup: SetupState }) =>
  state.setup.normalizeWarnings;
export const selectNormalizeErrors = (state: { setup: SetupState }) =>
  state.setup.normalizeErrors;
export const selectIsNormalizing = (state: { setup: SetupState }) =>
  state.setup.isNormalizing;
export const selectCompiledOutputs = (state: { setup: SetupState }) =>
  state.setup.compiledOutputs;
export const selectCompilePreviewWarnings = (state: { setup: SetupState }) =>
  state.setup.compileWarnings;
export const selectCompilePreviewErrors = (state: { setup: SetupState }) =>
  state.setup.compileErrors;
export const selectIsCompilingPreview = (state: { setup: SetupState }) =>
  state.setup.isCompilingPreview;
