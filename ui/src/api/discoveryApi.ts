import { baseApi } from './baseApi';
import { decodeProto } from './proto';
import type { DiscoveryResponse } from './types';
import { DiscoveryResponseSchema } from '@/gen/proto/screencast/studio/v1/web_pb';

export const discoveryApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getDiscovery: builder.query<DiscoveryResponse, void>({
      query: () => '/discovery',
      transformResponse: (response) => decodeProto(DiscoveryResponseSchema, response),
      providesTags: ['Discovery'],
    }),
  }),
});

export const { useGetDiscoveryQuery } = discoveryApi;
