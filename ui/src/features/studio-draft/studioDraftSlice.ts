import { createSlice, PayloadAction } from '@reduxjs/toolkit';

// ── Types ──

export type SourceType = 'Display' | 'Window' | 'Region' | 'Camera';

export interface Source {
  id: number;
  kind: SourceType;
  scene: string;
  armed: boolean;
  solo: boolean;
  label: string;
}

export interface StudioDraft {
  sources: Source[];
  format: 'MOV' | 'AVI' | 'MP4';
  fps: string;
  quality: number;
  audio: string;
  multiTrack: boolean;
  micInput: string;
  gain: number;
  micLevel: number;
}

const SOURCE_SCENES: Record<SourceType, string[]> = {
  Display: ['Desktop 1', 'Desktop 2'],
  Window: ['Finder', 'Terminal', 'Browser', 'Code Editor'],
  Region: ['Top Half', 'Bottom Half', 'Custom Region'],
  Camera: ['Built-in', 'USB Camera', 'FaceTime HD'],
};

const ICONS: Record<SourceType, string> = {
  Display: '🖥',
  Window: '☐',
  Region: '⊞',
  Camera: '◉',
};

// ── Initial State ──

let nextId = 1;

const makeSource = (kind: SourceType = 'Display'): Source => ({
  id: nextId++,
  kind,
  scene: SOURCE_SCENES[kind][0],
  armed: true,
  solo: false,
  label: `${kind} ${nextId - 1}`,
});

const initialState: StudioDraft = {
  sources: [makeSource('Display'), makeSource('Camera')],
  format: 'MOV',
  fps: '24 fps',
  quality: 75,
  audio: '48 kHz, 16-bit',
  multiTrack: true,
  micInput: 'Built-in Mic',
  gain: 55,
  micLevel: 0.12,
};

// ── Slice ──

const studioDraftSlice = createSlice({
  name: 'studioDraft',
  initialState,
  reducers: {
    addSource: (state, action: PayloadAction<SourceType>) => {
      state.sources.push(makeSource(action.payload));
    },
    removeSource: (state, action: PayloadAction<number>) => {
      state.sources = state.sources.filter((s) => s.id !== action.payload);
    },
    updateSource: (
      state,
      action: PayloadAction<{ id: number; patch: Partial<Source> }>
    ) => {
      const source = state.sources.find((s) => s.id === action.payload.id);
      if (source) {
        Object.assign(source, action.payload.patch);
      }
    },
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
    setMicLevel: (state, action: PayloadAction<number>) => {
      state.micLevel = action.payload;
    },
    resetDraft: () => {
      nextId = 1;
      return initialState;
    },
  },
});

// ── Exports ──

export const {
  addSource,
  removeSource,
  updateSource,
  setFormat,
  setFps,
  setQuality,
  setAudio,
  setMultiTrack,
  setMicInput,
  setGain,
  setMicLevel,
  resetDraft,
} = studioDraftSlice.actions;

export const studioDraftReducer = studioDraftSlice.reducer;

// ── Selectors ──

export const selectSources = (state: { studioDraft: StudioDraft }) =>
  state.studioDraft.sources;

export const selectArmedSources = (state: { studioDraft: StudioDraft }) =>
  state.studioDraft.sources.filter((s) => s.armed);

export const selectSourceById = (
  state: { studioDraft: StudioDraft },
  id: number
) => state.studioDraft.sources.find((s) => s.id === id);

export const selectOutputSettings = (state: { studioDraft: StudioDraft }) => ({
  format: state.studioDraft.format,
  fps: state.studioDraft.fps,
  quality: state.studioDraft.quality,
  audio: state.studioDraft.audio,
  multiTrack: state.studioDraft.multiTrack,
});

export const selectMicSettings = (state: { studioDraft: StudioDraft }) => ({
  micInput: state.studioDraft.micInput,
  gain: state.studioDraft.gain,
  micLevel: state.studioDraft.micLevel,
});

// Re-export for convenience
export { SOURCE_SCENES, ICONS };
