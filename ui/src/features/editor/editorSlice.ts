import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export const DEFAULT_DSL_TEXT = `schema: recorder.config/v1
session_id: demo

destination_templates:
  per_source: recordings/{session_id}/{source_name}.{ext}
  audio_mix: recordings/{session_id}/audio-mix.{ext}

screen_capture_defaults:
  capture:
    fps: 24
    cursor: true
    follow_resize: false
  output:
    container: mov
    video_codec: h264
    quality: 75

camera_capture_defaults:
  capture:
    fps: 30
    mirror: false
    size: "1280x720"
  output:
    container: mov
    video_codec: h264
    quality: 80

audio_defaults:
  output:
    codec: pcm_s16le
    sample_rate_hz: 48000
    channels: 2

audio_mix:
  destination_template: audio_mix

video_sources:
  - id: desktop-1
    name: Full Desktop
    type: display
    target:
      display: ":0.0"
    destination_template: per_source

audio_sources:
  - id: mic-1
    name: Built-in Mic
    device: default
    settings:
      gain: 1.0
`;

export interface EditorState {
  dslText: string;
  rawDslText: string;
  compileWarnings: string[];
  compileErrors: string[];
  isCompiling: boolean;
  structuredEditingLocked: boolean;
  structuredEditingLockReason: string;
}

const initialState: EditorState = {
  dslText: DEFAULT_DSL_TEXT,
  rawDslText: DEFAULT_DSL_TEXT,
  compileWarnings: [],
  compileErrors: [],
  isCompiling: false,
  structuredEditingLocked: false,
  structuredEditingLockReason: '',
};

const editorSlice = createSlice({
  name: 'editor',
  initialState,
  reducers: {
    setDslText(state, action: PayloadAction<string>) {
      state.dslText = action.payload;
      state.rawDslText = action.payload;
      state.compileErrors = [];
    },
    setRawDslText(state, action: PayloadAction<string>) {
      state.rawDslText = action.payload;
    },
    resetRawDslToApplied(state) {
      state.rawDslText = state.dslText;
      state.compileErrors = [];
    },
    compileStarted(state) {
      state.isCompiling = true;
      state.compileErrors = [];
    },
    compileSucceeded(
      state,
      action: PayloadAction<{
        dslText: string;
        warnings: string[];
        structuredEditingLocked: boolean;
        structuredEditingLockReason: string;
      }>
    ) {
      state.isCompiling = false;
      state.dslText = action.payload.dslText;
      state.rawDslText = action.payload.dslText;
      state.compileWarnings = action.payload.warnings;
      state.compileErrors = [];
      state.structuredEditingLocked = action.payload.structuredEditingLocked;
      state.structuredEditingLockReason = action.payload.structuredEditingLockReason;
    },
    compileFailed(state, action: PayloadAction<string[]>) {
      state.isCompiling = false;
      state.compileErrors = action.payload;
    },
    setStructuredEditingLock(
      state,
      action: PayloadAction<{ locked: boolean; reason?: string }>
    ) {
      state.structuredEditingLocked = action.payload.locked;
      state.structuredEditingLockReason = action.payload.reason ?? '';
    },
  },
});

export const {
  setDslText,
  setRawDslText,
  resetRawDslToApplied,
  compileStarted,
  compileSucceeded,
  compileFailed,
  setStructuredEditingLock,
} = editorSlice.actions;
export const editorReducer = editorSlice.reducer;

export const selectDslText = (state: { editor: EditorState }) =>
  state.editor.dslText;
export const selectRawDslText = (state: { editor: EditorState }) =>
  state.editor.rawDslText;
export const selectCompileWarnings = (state: { editor: EditorState }) =>
  state.editor.compileWarnings;
export const selectCompileErrors = (state: { editor: EditorState }) =>
  state.editor.compileErrors;
export const selectIsCompiling = (state: { editor: EditorState }) =>
  state.editor.isCompiling;
export const selectStructuredEditingLocked = (state: { editor: EditorState }) =>
  state.editor.structuredEditingLocked;
export const selectStructuredEditingLockReason = (state: { editor: EditorState }) =>
  state.editor.structuredEditingLockReason;
export const selectRawDslDirty = (state: { editor: EditorState }) =>
  state.editor.rawDslText !== state.editor.dslText;
