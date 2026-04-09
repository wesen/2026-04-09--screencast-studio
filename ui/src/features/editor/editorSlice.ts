import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export const DEFAULT_DSL_TEXT = `schema: recorder.config/v1
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
`;

export interface EditorState {
  dslText: string;
}

const initialState: EditorState = {
  dslText: DEFAULT_DSL_TEXT,
};

const editorSlice = createSlice({
  name: 'editor',
  initialState,
  reducers: {
    setDslText(state, action: PayloadAction<string>) {
      state.dslText = action.payload;
    },
  },
});

export const { setDslText } = editorSlice.actions;
export const editorReducer = editorSlice.reducer;

export const selectDslText = (state: { editor: EditorState }) =>
  state.editor.dslText;
