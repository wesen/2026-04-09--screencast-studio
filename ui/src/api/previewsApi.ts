import { baseApi } from './baseApi';
import type { PreviewDescriptor } from './types';

export const previewsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listPreviews: builder.query<PreviewDescriptor[], void>({
      query: () => '/previews',
      providesTags: ['Previews'],
    }),
    ensurePreview: builder.mutation<PreviewDescriptor, string>({
      query: (sourceId) => ({
        url: '/previews/ensure',
        method: 'POST',
        body: { source_id: sourceId },
      }),
      invalidatesTags: ['Previews'],
    }),
    releasePreview: builder.mutation<{ success: boolean }, string>({
      query: (sourceId) => ({
        url: '/previews/release',
        method: 'POST',
        body: { source_id: sourceId },
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
