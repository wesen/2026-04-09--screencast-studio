import { baseApi } from './baseApi';
import type { DiscoveryResponse, DiscoveryItem } from './types';

export const discoveryApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getDiscovery: builder.query<DiscoveryResponse, void>({
      query: () => '/discovery',
      providesTags: ['Discovery'],
    }),
    getDiscoveryByKind: builder.query<DiscoveryItem[], string>({
      query: (kind) => `/discovery/${kind}`,
      providesTags: (_result, _error, kind) => [{ type: 'DiscoveryItem', id: kind }],
    }),
    refreshDiscovery: builder.mutation<{ success: boolean }, void>({
      query: () => ({
        url: '/discovery/refresh',
        method: 'POST',
      }),
      invalidatesTags: ['Discovery'],
    }),
  }),
});

export const {
  useGetDiscoveryQuery,
  useGetDiscoveryByKindQuery,
  useRefreshDiscoveryMutation,
} = discoveryApi;
