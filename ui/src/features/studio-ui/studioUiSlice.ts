import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export type StudioTab = 'studio' | 'logs' | 'raw';

export interface StudioUiState {
  activeTab: StudioTab;
}

const initialState: StudioUiState = {
  activeTab: 'studio',
};

const studioUiSlice = createSlice({
  name: 'studioUi',
  initialState,
  reducers: {
    setActiveTab(state, action: PayloadAction<StudioTab>) {
      state.activeTab = action.payload;
    },
  },
});

export const { setActiveTab } = studioUiSlice.actions;
export const studioUiReducer = studioUiSlice.reducer;

export const selectActiveTab = (state: { studioUi: StudioUiState }) =>
  state.studioUi.activeTab;
