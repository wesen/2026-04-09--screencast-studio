import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { PreviewDescriptor } from '@/api/types';

export interface PreviewState {
  previewsById: Record<string, PreviewDescriptor>;
  ownedPreviewIdBySourceId: Record<string, string>;
}

const initialState: PreviewState = {
  previewsById: {},
  ownedPreviewIdBySourceId: {},
};

const sortPreviews = (previews: PreviewDescriptor[]): PreviewDescriptor[] =>
  [...previews].sort((left, right) => left.name.localeCompare(right.name));

const toPreviewMap = (
  previews: PreviewDescriptor[]
): Record<string, PreviewDescriptor> =>
  previews.reduce<Record<string, PreviewDescriptor>>((acc, preview) => {
    acc[preview.id] = preview;
    return acc;
  }, {});

const pruneOwnedPreviewMap = (
  previewsById: Record<string, PreviewDescriptor>,
  ownedPreviewIdBySourceId: Record<string, string>
): Record<string, string> =>
  Object.fromEntries(
    Object.entries(ownedPreviewIdBySourceId).filter(([, previewId]) => (
      previewId in previewsById
    ))
  );

const previewsSlice = createSlice({
  name: 'previews',
  initialState,
  reducers: {
    setPreviews: (state, action: PayloadAction<PreviewDescriptor[]>) => {
      state.previewsById = toPreviewMap(action.payload);
      state.ownedPreviewIdBySourceId = pruneOwnedPreviewMap(
        state.previewsById,
        state.ownedPreviewIdBySourceId
      );
    },
    upsertPreview: (state, action: PayloadAction<PreviewDescriptor>) => {
      state.previewsById[action.payload.id] = action.payload;
    },
    clearPreviews: (state) => {
      state.previewsById = {};
      state.ownedPreviewIdBySourceId = {};
    },
    trackOwnedPreview: (
      state,
      action: PayloadAction<{ sourceId: string; previewId: string }>
    ) => {
      state.ownedPreviewIdBySourceId[action.payload.sourceId] = action.payload.previewId;
    },
    clearOwnedPreview: (state, action: PayloadAction<{ sourceId: string }>) => {
      delete state.ownedPreviewIdBySourceId[action.payload.sourceId];
    },
  },
});

export const {
  setPreviews,
  upsertPreview,
  clearPreviews,
  trackOwnedPreview,
  clearOwnedPreview,
} = previewsSlice.actions;
export const previewReducer = previewsSlice.reducer;

export const selectPreviews = (state: { previews: PreviewState }) =>
  sortPreviews(Object.values(state.previews.previewsById));

export const selectOwnedPreviewIdBySourceId = (state: { previews: PreviewState }) =>
  state.previews.ownedPreviewIdBySourceId;

export const selectPreviewsById = (state: { previews: PreviewState }) =>
  state.previews.previewsById;
