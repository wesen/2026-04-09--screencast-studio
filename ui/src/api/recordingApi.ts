import { baseApi } from './baseApi';
import type { RecordingStartRequest, SessionEnvelope } from './types';

export const recordingApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getCurrentSession: builder.query<SessionEnvelope, void>({
      query: () => '/recordings/current',
      providesTags: ['Session'],
    }),
    startRecording: builder.mutation<SessionEnvelope, RecordingStartRequest>({
      query: (body) => ({
        url: '/recordings/start',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Session'],
    }),
    stopRecording: builder.mutation<SessionEnvelope, void>({
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
