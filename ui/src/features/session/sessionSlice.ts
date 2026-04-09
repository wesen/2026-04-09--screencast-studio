import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { ProcessLog, RecordingSession, WsEvent } from '@/api/types';

// ── Types ──

export interface SessionState {
  session: RecordingSession;
  logs: ProcessLog[];
  wsConnected: boolean;
}

// ── Initial State ──

const initialState: SessionState = {
  session: {
    active: false,
    outputs: [],
    warnings: [],
    logs: [],
  },
  logs: [],
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
      // Keep last 1000 logs
      if (state.logs.length > 1000) {
        state.logs = state.logs.slice(-1000);
      }
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

// ── Event Handler ──

export function handleWsEvent(
  state: SessionState,
  event: WsEvent
): SessionState {
  switch (event.type) {
    case 'session.state':
      return {
        ...state,
        session: event.payload,
        logs: event.payload.logs,
      };
    case 'session.log':
      return {
        ...state,
        logs: [...state.logs.slice(-999), event.payload],
      };
    case 'preview.list':
    case 'preview.state':
      return {
        ...state,
      };
    default:
      return state;
  }
}
