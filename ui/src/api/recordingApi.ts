import { baseApi } from './baseApi';
import { decodeProto, encodeProto } from './proto';
import type { SessionEnvelope } from './types';
import {
  RecordingStartRequestSchema,
  SessionEnvelopeSchema,
} from '@/gen/proto/screencast/studio/v1/web_pb';

export interface RecordingStartRequestBody {
  dsl: string;
  maxDurationSeconds?: number;
  gracePeriodSeconds?: number;
}

export const recordingApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getCurrentSession: builder.query<SessionEnvelope, void>({
      query: () => '/recordings/current',
      transformResponse: (response) => decodeProto(SessionEnvelopeSchema, response),
      providesTags: ['Session'],
    }),
    startRecording: builder.mutation<SessionEnvelope, RecordingStartRequestBody>({
      query: (body) => ({
        url: '/recordings/start',
        method: 'POST',
        body: encodeProto(RecordingStartRequestSchema, body),
      }),
      transformResponse: (response) => decodeProto(SessionEnvelopeSchema, response),
      invalidatesTags: ['Session'],
    }),
    stopRecording: builder.mutation<SessionEnvelope, void>({
      query: () => ({
        url: '/recordings/stop',
        method: 'POST',
      }),
      transformResponse: (response) => decodeProto(SessionEnvelopeSchema, response),
      invalidatesTags: ['Session'],
    }),
  }),
});

export const {
  useGetCurrentSessionQuery,
  useStartRecordingMutation,
  useStopRecordingMutation,
} = recordingApi;
