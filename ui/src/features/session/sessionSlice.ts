import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { RecordingState, WsEvent } from '@/api/types';

// ── Types ──

export interface SessionState {
  session: RecordingState;
  logs: LogEntry[];
  audioLevels: Record<string, number>;
  wsConnected: boolean;
  elapsed: number;
}

export interface LogEntry {
  timestamp: number;
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
}

// ── Initial State ──

const initialState: SessionState = {
  session: {
    active: false,
    session_id: '',
    state: 'idle',
    reason: '',
    outputs: [],
    warnings: [],
  },
  logs: [],
  audioLevels: {},
  wsConnected: false,
  elapsed: 0,
};

// ── Slice ──

const sessionSlice = createSlice({
  name: 'session',
  initialState,
  reducers: {
    setSession: (state, action: PayloadAction<RecordingState>) => {
      state.session = action.payload;
    },
    updateSessionState: (
      state,
      action: PayloadAction<{ state: string; reason: string }>
    ) => {
      state.session.state = action.payload.state as SessionState['session']['state'];
      state.session.reason = action.payload.reason;
    },
    setWsConnected: (state, action: PayloadAction<boolean>) => {
      state.wsConnected = action.payload;
    },
    addLog: (state, action: PayloadAction<LogEntry>) => {
      state.logs.push(action.payload);
      // Keep last 1000 logs
      if (state.logs.length > 1000) {
        state.logs = state.logs.slice(-1000);
      }
    },
    clearLogs: (state) => {
      state.logs = [];
    },
    setAudioLevel: (
      state,
      action: PayloadAction<{ deviceId: string; level: number }>
    ) => {
      state.audioLevels[action.payload.deviceId] = action.payload.level;
    },
    setElapsed: (state, action: PayloadAction<number>) => {
      state.elapsed = action.payload;
    },
    setMicLevel: (state, action: PayloadAction<number>) => {
      // Store mic level in audioLevels for the default device
      state.audioLevels['default'] = action.payload;
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
  clearLogs,
  setAudioLevel,
  setElapsed,
  setMicLevel,
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

export const selectAudioLevels = (state: { session: SessionState }) =>
  state.session.audioLevels;

export const selectElapsed = (state: { session: SessionState }) =>
  state.session.elapsed;

// ── Event Handler ──

export function handleWsEvent(
  state: SessionState,
  event: WsEvent
): SessionState {
  switch (event.type) {
    case 'session.state':
      return {
        ...state,
        session: {
          ...state.session,
          state: event.payload.state as SessionState['session']['state'],
          reason: event.payload.reason,
          active: event.payload.state === 'running',
        },
      };
    case 'session.log':
      return {
        ...state,
        logs: [
          ...state.logs.slice(-999),
          {
            timestamp: Date.now(),
            level: event.payload.level,
            message: event.payload.line,
          },
        ],
      };
    case 'session.output':
      return {
        ...state,
        session: {
          ...state.session,
          outputs: [
            ...state.session.outputs,
            {
              kind: event.payload.kind as 'video' | 'audio',
              name: event.payload.path.split('/').pop() || event.payload.path,
              path: event.payload.path,
            },
          ],
        },
      };
    case 'meter.audio':
      return {
        ...state,
        audioLevels: {
          ...state.audioLevels,
          [event.payload.device_id]: event.payload.level,
        },
      };
    default:
      return state;
  }
}
