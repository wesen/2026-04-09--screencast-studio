import { baseApi } from './baseApi';
import type { DiscoveryResponse } from './types';

export const discoveryApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getDiscovery: builder.query<DiscoveryResponse, void>({
      query: () => '/discovery',
      providesTags: ['Discovery'],
    }),
  }),
});

export const { useGetDiscoveryQuery } = discoveryApi;
