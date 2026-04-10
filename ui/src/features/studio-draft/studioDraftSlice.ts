import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface StudioDraft {
  format: 'MOV' | 'AVI' | 'MP4';
  fps: string;
  quality: number;
  audio: string;
  multiTrack: boolean;
  micInput: string;
  gain: number;
}

const initialState: StudioDraft = {
  format: 'MOV',
  fps: '24 fps',
  quality: 75,
  audio: '48 kHz, 16-bit',
  multiTrack: true,
  micInput: 'Built-in Mic',
  gain: 55,
};

const studioDraftSlice = createSlice({
  name: 'studioDraft',
  initialState,
  reducers: {
    setFormat: (state, action: PayloadAction<'MOV' | 'AVI' | 'MP4'>) => {
      state.format = action.payload;
    },
    setFps: (state, action: PayloadAction<string>) => {
      state.fps = action.payload;
    },
    setQuality: (state, action: PayloadAction<number>) => {
      state.quality = action.payload;
    },
    setAudio: (state, action: PayloadAction<string>) => {
      state.audio = action.payload;
    },
    setMultiTrack: (state, action: PayloadAction<boolean>) => {
      state.multiTrack = action.payload;
    },
    setMicInput: (state, action: PayloadAction<string>) => {
      state.micInput = action.payload;
    },
    setGain: (state, action: PayloadAction<number>) => {
      state.gain = action.payload;
    },
    resetDraft: () => initialState,
  },
});

export const {
  setFormat,
  setFps,
  setQuality,
  setAudio,
  setMultiTrack,
  setMicInput,
  setGain,
  resetDraft,
} = studioDraftSlice.actions;

export const studioDraftReducer = studioDraftSlice.reducer;
