import { baseApi } from './baseApi';
import type {
  PreviewEnsureRequest,
  PreviewEnvelope,
  PreviewListResponse,
  PreviewReleaseRequest,
} from './types';

export const previewsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listPreviews: builder.query<PreviewListResponse, void>({
      query: () => '/previews',
      providesTags: ['Previews'],
    }),
    ensurePreview: builder.mutation<PreviewEnvelope, PreviewEnsureRequest>({
      query: (body) => ({
        url: '/previews/ensure',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Previews'],
    }),
    releasePreview: builder.mutation<PreviewEnvelope, PreviewReleaseRequest>({
      query: (body) => ({
        url: '/previews/release',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Previews'],
    }),
  }),
});

export const {
  useListPreviewsQuery,
  useEnsurePreviewMutation,
  useReleasePreviewMutation,
} = previewsApi;
