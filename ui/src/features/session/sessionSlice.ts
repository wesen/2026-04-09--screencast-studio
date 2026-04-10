import { create } from '@bufbuild/protobuf';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { AudioMeterEvent, DiskTelemetryEvent, ProcessLog, RecordingSession } from '@/api/types';
import { RecordingSessionSchema } from '@/gen/proto/screencast/studio/v1/web_pb';

// ── Types ──

export interface SessionState {
  session: RecordingSession;
  logs: ProcessLog[];
  audioMeter: AudioMeterEvent | null;
  diskStatus: DiskTelemetryEvent | null;
  wsConnected: boolean;
}

// ── Initial State ──

const initialState: SessionState = {
  session: create(RecordingSessionSchema, {
    active: false,
    outputs: [],
    warnings: [],
    logs: [],
  }),
  logs: [],
  audioMeter: null,
  diskStatus: null,
  wsConnected: false,
};

// ── Slice ──

const sessionSlice = createSlice({
  name: 'session',
  initialState,
  reducers: {
    setSession: (state, action: PayloadAction<RecordingSession>) => {
      state.session = action.payload;
      state.logs = action.payload.logs;
    },
    updateSessionState: (
      state,
      action: PayloadAction<RecordingSession>
    ) => {
      state.session = {
        ...state.session,
        ...action.payload,
      };
      if (action.payload.logs) {
        state.logs = action.payload.logs;
      }
    },
    setWsConnected: (state, action: PayloadAction<boolean>) => {
      state.wsConnected = action.payload;
    },
    addLog: (state, action: PayloadAction<ProcessLog>) => {
      state.logs.push(action.payload);
      state.session.logs = state.logs;
      // Keep last 1000 logs
      if (state.logs.length > 1000) {
        state.logs = state.logs.slice(-1000);
        state.session.logs = state.logs;
      }
    },
    setAudioMeter: (state, action: PayloadAction<AudioMeterEvent>) => {
      state.audioMeter = action.payload;
    },
    setDiskStatus: (state, action: PayloadAction<DiskTelemetryEvent>) => {
      state.diskStatus = action.payload;
    },
    clearLogs: (state) => {
      state.logs = [];
    },
    resetSession: () => initialState,
  },
});

// ── Exports ──

export const {
  setSession,
  updateSessionState,
  setWsConnected,
  addLog,
  setAudioMeter,
  setDiskStatus,
  clearLogs,
  resetSession,
} = sessionSlice.actions;

export const sessionReducer = sessionSlice.reducer;

// ── Selectors ──

export const selectSession = (state: { session: SessionState }) =>
  state.session.session;

export const selectIsRecording = (state: { session: SessionState }) =>
  state.session.session.active;

export const selectSessionState = (state: { session: SessionState }) =>
  state.session.session.state;

export const selectWsConnected = (state: { session: SessionState }) =>
  state.session.wsConnected;

export const selectLogs = (state: { session: SessionState }) =>
  state.session.logs;
export const selectAudioMeter = (state: { session: SessionState }) =>
  state.session.audioMeter;
export const selectDiskStatus = (state: { session: SessionState }) =>
  state.session.diskStatus;
