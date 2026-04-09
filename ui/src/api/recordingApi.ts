import { baseApi } from './baseApi';
import type { RecordingState, RecordingStartResponse, RecordingStopResponse } from './types';

export const recordingApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getCurrentSession: builder.query<RecordingState, void>({
      query: () => '/recordings/current',
      providesTags: ['Session'],
    }),
    startRecording: builder.mutation<RecordingStartResponse, { dsl: string; max_duration_seconds?: number }>({
      query: (body) => ({
        url: '/recordings/start',
        method: 'POST',
        body: { dsl_format: 'yaml', ...body },
      }),
      invalidatesTags: ['Session'],
    }),
    stopRecording: builder.mutation<RecordingStopResponse, void>({
      query: () => ({
        url: '/recordings/stop',
        method: 'POST',
      }),
      invalidatesTags: ['Session'],
    }),
  }),
});

export const {
  useGetCurrentSessionQuery,
  useStartRecordingMutation,
  useStopRecordingMutation,
} = recordingApi;
