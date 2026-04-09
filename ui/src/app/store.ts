import { configureStore } from '@reduxjs/toolkit';
import { baseApi } from '@/api/baseApi';
import { editorReducer } from '@/features/editor/editorSlice';
import { previewReducer } from '@/features/previews/previewSlice';
import { setupReducer } from '@/features/setup/setupSlice';
import { studioDraftReducer } from '@/features/studio-draft/studioDraftSlice';
import { studioUiReducer } from '@/features/studio-ui/studioUiSlice';
import { sessionReducer } from '@/features/session/sessionSlice';

export const store = configureStore({
  reducer: {
    [baseApi.reducerPath]: baseApi.reducer,
    editor: editorReducer,
    previews: previewReducer,
    setup: setupReducer,
    studioDraft: studioDraftReducer,
    studioUi: studioUiReducer,
    session: sessionReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(baseApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
