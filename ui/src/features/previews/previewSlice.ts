import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { PreviewDescriptor } from '@/api/types';

export interface PreviewState {
  previews: PreviewDescriptor[];
}

const initialState: PreviewState = {
  previews: [],
};

const sortPreviews = (previews: PreviewDescriptor[]): PreviewDescriptor[] =>
  [...previews].sort((left, right) => left.name.localeCompare(right.name));

const previewsSlice = createSlice({
  name: 'previews',
  initialState,
  reducers: {
    setPreviews(state, action: PayloadAction<PreviewDescriptor[]>) {
      state.previews = sortPreviews(action.payload);
    },
    upsertPreview(state, action: PayloadAction<PreviewDescriptor>) {
      const next = state.previews.filter((item) => item.id !== action.payload.id);
      next.push(action.payload);
      state.previews = sortPreviews(next);
    },
    clearPreviews(state) {
      state.previews = [];
    },
  },
});

export const { setPreviews, upsertPreview, clearPreviews } = previewsSlice.actions;
export const previewReducer = previewsSlice.reducer;

export const selectPreviews = (state: { previews: PreviewState }) =>
  state.previews.previews;
