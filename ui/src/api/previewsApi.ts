import { baseApi } from './baseApi';
import { decodeProto, encodeProto } from './proto';
import type {
  PreviewEnvelope,
  PreviewListResponse,
} from './types';
import {
  PreviewEnsureRequestSchema,
  PreviewEnvelopeSchema,
  PreviewListResponseSchema,
  PreviewReleaseRequestSchema,
} from '@/gen/proto/screencast/studio/v1/web_pb';

export interface PreviewEnsureRequestBody {
  dsl: string;
  sourceId: string;
}

export interface PreviewReleaseRequestBody {
  previewId: string;
}

export const previewsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listPreviews: builder.query<PreviewListResponse, void>({
      query: () => '/previews',
      transformResponse: (response) => decodeProto(PreviewListResponseSchema, response),
      providesTags: ['Previews'],
    }),
    ensurePreview: builder.mutation<PreviewEnvelope, PreviewEnsureRequestBody>({
      query: (body) => ({
        url: '/previews/ensure',
        method: 'POST',
        body: encodeProto(PreviewEnsureRequestSchema, body),
      }),
      transformResponse: (response) => decodeProto(PreviewEnvelopeSchema, response),
      invalidatesTags: ['Previews'],
    }),
    releasePreview: builder.mutation<PreviewEnvelope, PreviewReleaseRequestBody>({
      query: (body) => ({
        url: '/previews/release',
        method: 'POST',
        body: encodeProto(PreviewReleaseRequestSchema, body),
      }),
      transformResponse: (response) => decodeProto(PreviewEnvelopeSchema, response),
      invalidatesTags: ['Previews'],
    }),
  }),
});

export const {
  useListPreviewsQuery,
  useEnsurePreviewMutation,
  useReleasePreviewMutation,
} = previewsApi;
