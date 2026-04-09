import { configureStore } from '@reduxjs/toolkit';
import { baseApi } from '@/api/baseApi';
import { studioDraftReducer } from '@/features/studio-draft/studioDraftSlice';
import { sessionReducer } from '@/features/session/sessionSlice';

export const store = configureStore({
  reducer: {
    [baseApi.reducerPath]: baseApi.reducer,
    studioDraft: studioDraftReducer,
    session: sessionReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(baseApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
